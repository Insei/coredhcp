// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package sleep

// This plugin introduces a delay in the DHCP response.

import (
	"fmt"
	"time"

	"github.com/insei/coredhcp/handler"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
)

const pluginName = "sleep"

// Example configuration of the `sleep` plugin:
//
// server4:
//   plugins:
//     - sleep 300ms
//     - file: "leases4.txt"
//
// server6:
//   plugins:
//     - sleep 1s
//     - file: "leases6.txt"
//
// For the duration format, see the documentation of `time.ParseDuration`,
// https://golang.org/pkg/time/#ParseDuration .

// Plugin contains the `sleep` plugin data.
var Plugin = plugins.Plugin{
	Name:   pluginName,
	Setup6: setup6,
	Setup4: setup4,
}

type pluginState struct {
	delay time.Duration
	log   logger.FieldLogger
}

func setup6(serverLogger logger.FieldLogger, args ...string) (handler.Handler6, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want exactly one argument, got %d", len(args))
	}
	delay, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}
	pState := &pluginState{
		delay: delay,
		log:   logger.CreatePluginLogger(serverLogger, pluginName, true),
	}
	pState.log.Printf("loaded plugin for DHCPv6.")
	return makeSleepHandler6(pState), nil
}

func setup4(serverLogger logger.FieldLogger, args ...string) (handler.Handler4, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("want exactly one argument, got %d", len(args))
	}
	delay, err := time.ParseDuration(args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}
	pState := &pluginState{
		delay: delay,
		log:   logger.CreatePluginLogger(serverLogger, pluginName, true),
	}
	pState.log.Printf("loaded plugin for DHCPv4.")
	return makeSleepHandler4(pState), nil
}

func makeSleepHandler6(pState *pluginState) handler.Handler6 {
	return func(req, resp dhcpv6.DHCPv6) (dhcpv6.DHCPv6, bool) {
		pState.log.Printf("introducing delay of %s in response", pState.delay)
		// return the unmodified response, and instruct coredhcp to continue to
		// the next plugin.
		time.Sleep(pState.delay)
		return resp, false
	}
}

func makeSleepHandler4(pState *pluginState) handler.Handler4 {
	return func(req, resp *dhcpv4.DHCPv4) (*dhcpv4.DHCPv4, bool) {
		pState.log.Printf("introducing delay of %s in response", pState.delay)
		// return the unmodified response, and instruct coredhcp to continue to
		// the next plugin.
		time.Sleep(pState.delay)
		return resp, false
	}
}
