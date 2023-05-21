// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package rangeplugin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insei/coredhcp/plugins/allocators"
	"github.com/insei/coredhcp/plugins/allocators/bitmap"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/sirupsen/logrus"
)

const pluginName = "range"

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup4: setup4,
}

//Record holds an IP lease record
type Record struct {
	IP      net.IP
	expires time.Time
}

// pluginState is the data held by an instance of the range plugin
type pluginState struct {
	// Rough lock for the whole plugin, we'll get better performance once we use leasestorage
	sync.Mutex
	// Recordsv4 holds a MAC -> IP address and lease time mapping
	Recordsv4 map[string]*Record
	LeaseTime time.Duration
	leasefile *os.File
	allocator allocators.Allocator
	log       logrus.FieldLogger
}

// Handler4 handles DHCPv4 packets for the range plugin
func (p *pluginState) Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	p.Lock()
	defer p.Unlock()
	record, ok := p.Recordsv4[req.ClientHWAddr.String()]
	if !ok {
		// Allocating new address since there isn't one allocated
		p.log.Printf("MAC address %s is new, leasing new IPv4 address", req.ClientHWAddr.String())
		ip, err := p.allocator.Allocate(net.IPNet{})
		if err != nil {
			p.log.Errorf("Could not allocate IP for MAC %s: %v", req.ClientHWAddr.String(), err)
			return nil, true
		}
		rec := Record{
			IP:      ip.IP.To4(),
			expires: time.Now().Add(p.LeaseTime),
		}
		err = saveIPAddress(p.leasefile, req.ClientHWAddr, &rec)
		if err != nil {
			p.log.Errorf("SaveIPAddress for MAC %s failed: %v", req.ClientHWAddr.String(), err)
		}
		p.Recordsv4[req.ClientHWAddr.String()] = &rec
		record = &rec
	} else {
		// Ensure we extend the existing lease at least past when the one we're giving expires
		if record.expires.Before(time.Now().Add(p.LeaseTime)) {
			record.expires = time.Now().Add(p.LeaseTime).Round(time.Second)
			err := saveIPAddress(p.leasefile, req.ClientHWAddr, record)
			if err != nil {
				p.log.Errorf("Could not persist lease for MAC %s: %v", req.ClientHWAddr.String(), err)
			}
		}
	}
	resp.YourIPAddr = record.IP
	resp.Options.Update(dhcpv4.OptIPAddressLeaseTime(p.LeaseTime.Round(time.Second)))
	p.log.Printf("found IP address %s for MAC %s", record.IP, req.ClientHWAddr.String())
	return resp, false
}

func setup4(serverLogger logrus.FieldLogger, args ...string) (handler.Handler4, error) {
	var err error
	pState := pluginState{log: logger.CreatePluginLogger(serverLogger, pluginName, false)}

	if len(args) < 4 {
		return nil, fmt.Errorf("invalid number of arguments, want: 4 (file name, start IP, end IP, lease time), got: %d", len(args))
	}
	filename := args[0]
	if filename == "" {
		return nil, errors.New("file name cannot be empty")
	}
	ipRangeStart := net.ParseIP(args[1])
	if ipRangeStart.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %v", args[1])
	}
	ipRangeEnd := net.ParseIP(args[2])
	if ipRangeEnd.To4() == nil {
		return nil, fmt.Errorf("invalid IPv4 address: %v", args[2])
	}
	if binary.BigEndian.Uint32(ipRangeStart.To4()) >= binary.BigEndian.Uint32(ipRangeEnd.To4()) {
		return nil, errors.New("start of IP range has to be lower than the end of an IP range")
	}

	pState.allocator, err = bitmap.NewIPv4Allocator(ipRangeStart, ipRangeEnd)
	if err != nil {
		return nil, fmt.Errorf("could not create an allocator: %w", err)
	}

	pState.LeaseTime, err = time.ParseDuration(args[3])
	if err != nil {
		return nil, fmt.Errorf("invalid lease duration: %v", args[3])
	}

	pState.Recordsv4, err = loadRecordsFromFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not load records from file: %v", err)
	}

	pState.log.Printf("Loaded %d DHCPv4 leases from %s", len(pState.Recordsv4), filename)

	for _, v := range pState.Recordsv4 {
		ip, err := pState.allocator.Allocate(net.IPNet{IP: v.IP})
		if err != nil {
			return nil, fmt.Errorf("failed to re-allocate leased ip %v: %v", v.IP.String(), err)
		}
		if ip.IP.String() != v.IP.String() {
			return nil, fmt.Errorf("allocator did not re-allocate requested leased ip %v: %v", v.IP.String(), ip.String())
		}
	}

	if err := registerBackingFile(&pState.leasefile, filename); err != nil {
		return nil, fmt.Errorf("could not setup lease storage: %w", err)
	}

	return pState.Handler4, nil
}
