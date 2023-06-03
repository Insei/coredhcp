// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package leasetime

import (
	"errors"
	"time"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
)

const pluginName = "lease_time"

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name: pluginName,
	// currently not supported for DHCPv6
	Setup6: nil,
	Setup4: setup4,
}

type pluginState struct {
	leaseTime time.Duration
}

// Handler4 handles DHCPv4 packets for the lease_time plugin.
func (p *pluginState) Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	if req.OpCode != dhcpv4.OpcodeBootRequest {
		return resp, false
	}
	// Set lease time unless it has already been set
	if !resp.Options.Has(dhcpv4.OptionIPAddressLeaseTime) {
		resp.Options.Update(dhcpv4.OptIPAddressLeaseTime(p.leaseTime))
	}
	return resp, false
}

func setup4(serverLogger logger.FieldLogger, args ...string) (handler.Handler4, error) {
	plog := logger.CreatePluginLogger(serverLogger, pluginName, false)
	plog.Print("loading `lease_time` plugin")
	if len(args) < 1 {
		plog.Error("No default lease time provided")
		return nil, errors.New("lease_time failed to initialize")
	}

	leaseTime, err := time.ParseDuration(args[0])
	if err != nil {
		plog.Errorf("invalid duration: %v", args[0])
		return nil, errors.New("lease_time failed to initialize")
	}
	pState := pluginState{leaseTime: leaseTime}

	return pState.Handler4, nil
}
