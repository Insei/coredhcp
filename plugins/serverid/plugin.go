// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package serverid

import (
	"errors"
	"net"
	"strings"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/insomniacslk/dhcp/iana"
	"github.com/sirupsen/logrus"
)

const pluginName = "server_id"

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup6: setup6,
	Setup4: setup4,
}

type pluginStateV6 struct {
	// v6ServerID is the DUID of the v6 server
	serverID *dhcpv6.Duid
	log      logrus.FieldLogger
}

type pluginStateV4 struct {
	serverID net.IP
	log      logrus.FieldLogger
}

// Handler6 handles DHCPv6 packets for the server_id plugin.
func (p pluginStateV6) Handler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	if p.serverID == nil {
		p.log.Fatal("BUG: Plugin is running uninitialized!")
		return nil, true
	}

	msg, err := req.GetInnerMessage()
	if err != nil {
		// BUG: this should already have failed in the main handler. Abort
		p.log.Error(err)
		return nil, true
	}

	if sid := msg.Options.ServerID(); sid != nil {
		// RFC8415 §16.{2,5,7}
		// These message types MUST be discarded if they contain *any* ServerID option
		if msg.MessageType == dhcpv6.MessageTypeSolicit ||
			msg.MessageType == dhcpv6.MessageTypeConfirm ||
			msg.MessageType == dhcpv6.MessageTypeRebind {
			return nil, true
		}

		// Approximately all others MUST be discarded if the ServerID doesn't match
		if !sid.Equal(*p.serverID) {
			p.log.Infof("requested server ID does not match this server's ID. Got %v, want %v", sid, *p.serverID)
			return nil, true
		}
	} else if msg.MessageType == dhcpv6.MessageTypeRequest ||
		msg.MessageType == dhcpv6.MessageTypeRenew ||
		msg.MessageType == dhcpv6.MessageTypeDecline ||
		msg.MessageType == dhcpv6.MessageTypeRelease {
		// RFC8415 §16.{6,8,10,11}
		// These message types MUST be discarded if they *don't* contain a ServerID option
		return nil, true
	}
	dhcpv6.WithServerID(*p.serverID)(resp)
	return resp, false
}

// Handler4 handles DHCPv4 packets for the server_id plugin.
func (p pluginStateV4) Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	if p.serverID == nil {
		p.log.Fatal("BUG: Plugin is running uninitialized!")
		return nil, true
	}
	if req.OpCode != dhcpv4.OpcodeBootRequest {
		p.log.Warningf("not a BootRequest, ignoring")
		return resp, false
	}
	if req.ServerIPAddr != nil &&
		!req.ServerIPAddr.Equal(net.IPv4zero) &&
		!req.ServerIPAddr.Equal(p.serverID) {
		// This request is not for us, drop it.
		p.log.Infof("requested server ID does not match this server's ID. Got %v, want %v", req.ServerIPAddr, p.serverID)
		return nil, true
	}
	resp.ServerIPAddr = make(net.IP, net.IPv4len)
	copy(resp.ServerIPAddr[:], p.serverID)
	resp.UpdateOption(dhcpv4.OptServerIdentifier(p.serverID))
	return resp, false
}

func setup4(serverLogger logrus.FieldLogger, args ...string) (handler.Handler4, error) {
	pState := &pluginStateV4{log: logger.CreatePluginLogger(serverLogger, pluginName, false)}
	pState.log.Printf("loading `server_id` plugin for DHCPv4 with args: %v", args)
	if len(args) < 1 {
		return nil, errors.New("need an argument")
	}
	serverID := net.ParseIP(args[0])
	if serverID == nil {
		return nil, errors.New("invalid or empty IP address")
	}
	if serverID.To4() == nil {
		return nil, errors.New("not a valid IPv4 address")
	}
	pState.serverID = serverID.To4()
	return pState.Handler4, nil
}

func setup6(serverLogger logrus.FieldLogger, args ...string) (handler.Handler6, error) {
	pState := &pluginStateV6{log: logger.CreatePluginLogger(serverLogger, pluginName, true)}
	pState.log.Printf("loading `server_id` plugin for DHCPv6 with args: %v", args)
	if len(args) < 2 {
		return nil, errors.New("need a DUID type and value")
	}
	duidType := args[0]
	if duidType == "" {
		return nil, errors.New("got empty DUID type")
	}
	duidValue := args[1]
	if duidValue == "" {
		return nil, errors.New("got empty DUID value")
	}
	duidType = strings.ToLower(duidType)
	hwaddr, err := net.ParseMAC(duidValue)
	if err != nil {
		return nil, err
	}
	var v6ServerID *dhcpv6.Duid
	switch duidType {
	case "ll", "duid-ll", "duid_ll":
		v6ServerID = &dhcpv6.Duid{
			Type: dhcpv6.DUID_LL,
			// sorry, only ethernet for now
			HwType:        iana.HWTypeEthernet,
			LinkLayerAddr: hwaddr,
		}
	case "llt", "duid-llt", "duid_llt":
		v6ServerID = &dhcpv6.Duid{
			Type: dhcpv6.DUID_LLT,
			// sorry, zero-time for now
			Time: 0,
			// sorry, only ethernet for now
			HwType:        iana.HWTypeEthernet,
			LinkLayerAddr: hwaddr,
		}
	case "en", "uuid":
		return nil, errors.New("EN/UUID DUID type not supported yet")
	default:
		return nil, errors.New("Opaque DUID type not supported yet")
	}
	pState.serverID = v6ServerID
	pState.log.Printf("using %s %s", duidType, duidValue)
	return pState.Handler6, nil
}
