// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package dns

import (
	"errors"
	"net"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/sirupsen/logrus"
)

const pluginName = "dns"

// Plugin wraps the DNS plugin information.
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup6: setup6,
	Setup4: setup4,
}

type pluginState struct {
	dnsServers []net.IP
	log        logrus.FieldLogger
}

func setup6(serverLogger logrus.FieldLogger, args ...string) (handler.Handler6, error) {
	pState := &pluginState{log: logger.CreatePluginLogger(serverLogger, pluginName, true)}
	if len(args) < 1 {
		return nil, errors.New("need at least one DNS server")
	}
	for _, arg := range args {
		server := net.ParseIP(arg)
		if server.To16() == nil {
			return pState.Handler6, errors.New("expected an DNS server address, got: " + arg)
		}
		pState.dnsServers = append(pState.dnsServers, server)
	}
	pState.log.Infof("loaded %d DNS servers.", len(pState.dnsServers))
	return pState.Handler6, nil
}

func setup4(serverLogger logrus.FieldLogger, args ...string) (handler.Handler4, error) {
	pState := &pluginState{log: logger.CreatePluginLogger(serverLogger, pluginName, false)}
	pState.log.Printf("loaded plugin for DHCPv4.")
	if len(args) < 1 {
		return nil, errors.New("need at least one DNS server")
	}
	for _, arg := range args {
		DNSServer := net.ParseIP(arg)
		if DNSServer.To4() == nil {
			return pState.Handler4, errors.New("expected an DNS server address, got: " + arg)
		}
		pState.dnsServers = append(pState.dnsServers, DNSServer)
	}
	pState.log.Infof("loaded %d DNS servers.", len(pState.dnsServers))
	return pState.Handler4, nil
}

// Handler6 handles DHCPv6 packets for the dns plugin
func (p *pluginState) Handler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	decap, err := req.GetInnerMessage()
	if err != nil {
		p.log.Errorf("Could not decapsulate relayed message, aborting: %v", err)
		return nil, true
	}

	if decap.IsOptionRequested(dhcpv6.OptionDNSRecursiveNameServer) {
		resp.UpdateOption(dhcpv6.OptDNS(p.dnsServers...))
	}
	return resp, false
}

//Handler4 handles DHCPv4 packets for the dns plugin
func (p *pluginState) Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	if req.IsOptionRequested(dhcpv4.OptionDomainNameServer) {
		resp.Options.Update(dhcpv4.OptDNS(p.dnsServers...))
	}
	return resp, false
}
