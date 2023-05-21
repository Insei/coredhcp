// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package router

import (
	"errors"
	"net"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/sirupsen/logrus"
)

const pluginName = "router"

// Plugin wraps plugin registration information
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup4: setup4,
}

type pluginState struct {
	routers []net.IP
	log     logrus.FieldLogger
}

func setup4(serverLogger logrus.FieldLogger, args ...string) (handler.Handler4, error) {
	pState := &pluginState{routers: []net.IP{}, log: logger.CreatePluginLogger(serverLogger, pluginName, false)}
	pState.log.Printf("Loaded plugin for DHCPv4.")
	if len(args) < 1 {
		return nil, errors.New("need at least one router IP address")
	}
	for _, arg := range args {
		router := net.ParseIP(arg)
		if router.To4() == nil {
			return pState.Handler4, errors.New("expected an router IP address, got: " + arg)
		}
		pState.routers = append(pState.routers, router)
	}
	pState.log.Infof("loaded %d router IP addresses.", len(pState.routers))
	return pState.Handler4, nil
}

// Handler4 handles DHCPv4 packets for the router plugin
func (p *pluginState) Handler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	resp.Options.Update(dhcpv4.OptRouter(p.routers...))
	return resp, false
}
