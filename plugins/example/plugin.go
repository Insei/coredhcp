// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package example is example plugin
package example

// This is an example plugin that inspects a packet and prints it out. The code
// is commented in a way that should walk you through the implementation of your
// own plugins.
// Feedback is welcome!

import (
	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/sirupsen/logrus"
)

// Plugin wraps the information necessary to register a plugin.
// In the main package, you need to export a `plugins.Plugin` object called
// `Plugin`, so it can be registered into the plugin registry.
// Just import your plugin, and fill the structure with plugin name and setup
// functions:
//
// import (
//     "github.com/insei/coredhcp/plugins"
//     "github.com/insei/coredhcp/plugins/example"
// )
//
// var Plugin = plugins.Plugin{
//     Name: "example",
//     Setup6: setup6,
//     Setup4: setup4,
// }
//
// Name is simply the name used to register the plugin. It must be unique to
// other registered plugins, or the operation will fail. In other words, don't
// declare plugins with colliding names.
//
// Setup6 and Setup4 are the setup functions for DHCPv6 and DHCPv4 traffic
// handlers. They conform to the `plugins.SetupFunc6` and `plugins.SetupFunc4`
// interfaces, so they must return a `plugins.Handler6` and a `plugins.Handler4`
// respectively.
// A `nil` setup function means that that protocol won't be handled by this
// plugin.
//
// Note that importing the plugin is not enough to use it: you have to
// explicitly specify the intention to use it in the `config.yml` file, in the
// plugins section. For example:
//
// server6:
//   listen: '[::]547'
//   - example:
//   - server_id: LL aa:bb:cc:dd:ee:ff
//   - file: "leases.txt"
//

const pluginName = "example"

// Plugin represents a plugin object.
//// Setup6 and Setup4 are the setup functions for DHCPv6 and DHCPv4 handlers
//// respectively. Both setup functions can be nil.
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup6: setup6,
	Setup4: setup4,
}

type pluginState struct {
	log logrus.FieldLogger
}

// setup6 is the setup function to initialize the handler for DHCPv6
// traffic. This function implements the `plugin.SetupFunc6` interface.
// This function returns a `handler.Handler6` function, and an error if any.
// In this example we do very little in the setup function, and just return the
// `exampleHandler6` function. Such function will be called for every DHCPv6
// packet that the server receives. Remember that a handler may not be called
// for each packet, if the handler chain is interrupted before reaching it.
func setup6(serverLogger logrus.FieldLogger, args ...string) (handler.Handler6, error) {
	pState := &pluginState{
		log: logger.CreatePluginLogger(serverLogger, pluginName, true),
	}
	pState.log.Printf("loaded plugin")
	pState.log.Printf("args: %s", args)
	return pState.exampleHandler6, nil
}

// setup4 behaves like setupExample6, but for DHCPv4 packets. It
// implements the `plugin.SetupFunc4` interface.
func setup4(serverLogger logrus.FieldLogger, args ...string) (handler.Handler4, error) {
	pState := &pluginState{
		log: logger.CreatePluginLogger(serverLogger, pluginName, false),
	}
	pState.log.Printf("loaded plugin")
	pState.log.Printf("args: %s", args)
	return pState.exampleHandler4, nil
}

// exampleHandler6 handles DHCPv6 packets for the example plugin. It implements
// the `handler.Handler6` interface. The input arguments are the request packet
// that the server received from a client, and the response packet that has been
// computed so far. This function returns the response packet to be sent back to
// the client, and a boolean.
// The response can be either the same response packet received as input, a
// modified response packet, or nil. If nil, the server will not reply to the
// client, basically dropping the request.
// The returned boolean indicates to the server whether the chain of plugins
// should continue or not. If `true`, the server will stop at this plugin, and
// respond to the client (or drop the response, if nil). If `false`, the server
// will call the next plugin in the chan, using the returned response packet as
// input for the next plugin.
func (p *pluginState) exampleHandler6(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
	p.log.Printf("received DHCPv6 packet: %s", req.Summary())
	// return the unmodified response, and false. This means that the next
	// plugin in the chain will be called, and the unmodified response packet
	// will be used as its input.
	return resp, false
}

// exampleHandler4 behaves like exampleHandler6, but for DHCPv4 packets. It
// implements the `handler.Handler4` interface.
func (p *pluginState) exampleHandler4(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
	p.log.Printf("received DHCPv4 packet: %s", req.Summary())
	// return the unmodified response, and false. This means that the next
	// plugin in the chain will be called, and the unmodified response packet
	// will be used as its input.
	return resp, false
}
