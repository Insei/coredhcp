// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/insei/coredhcp/logger"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

//Parser is a struct for parsing yml config to Config via Viper
type Parser struct {
	config *Config
	logger logger.FieldLogger
	v      *viper.Viper
}

//NewParser creates Viper based parser for Config
func NewParser(logger logger.FieldLogger) *Parser {
	return &Parser{
		config: New(),
		logger: logger.WithField("prefix", "config-parser"),
		v:      viper.New(),
	}
}

// Parse reads a configuration file and returns a Config object, or an error if
// any.
func (p *Parser) Parse(path string) (*Config, error) {
	filename := filepath.Base(path)
	if !strings.HasSuffix(filename, ".config.yml") {
		return nil, fmt.Errorf("incorrect config name, correct: <server-name>.config.yml")
	}
	p.config.Name = filename[:len(filename)-len(".config.yml")]
	p.logger = p.logger.WithField("server", p.config.Name)

	p.logger.Print("Loading configuration")
	p.v.SetConfigType("yml")
	p.v.SetConfigFile(path)

	if err := p.v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := p.parseConfig(protocolV6); err != nil {
		return nil, err
	}
	if err := p.parseConfig(protocolV4); err != nil {
		return nil, err
	}
	if p.config.Server6 == nil && p.config.Server4 == nil {
		return nil, ConfigErrorFromString("need at least one valid config for DHCPv6 or DHCPv4")
	}
	return p.config, nil
}

func (p *Parser) parseListen(ver protocolVersion) ([]net.UDPAddr, error) {
	if err := protoVersionCheck(ver); err != nil {
		return nil, err
	}

	listen := p.v.Get(fmt.Sprintf("server%d.listen", ver))

	// Provide an emulation of the old keyword "interface" to avoid breaking config files
	if iface := p.v.Get(fmt.Sprintf("server%d.interface", ver)); iface != nil && listen != nil {
		return nil, ConfigErrorFromString("interface is a deprecated alias for listen, " +
			"both cannot be used at the same time. Choose one and remove the other.")
	} else if iface != nil {
		listen = "%" + cast.ToString(iface)
	}

	if listen == nil {
		return defaultListen(ver)
	}

	addrs, err := cast.ToStringSliceE(listen)
	if err != nil {
		addrs = []string{cast.ToString(listen)}
	}

	listeners := []net.UDPAddr{}
	for _, a := range addrs {
		l, err := getListenAddress(a, ver)
		if err != nil {
			return nil, err
		}

		if l.Zone == "" && (l.IP.IsLinkLocalMulticast() || l.IP.IsInterfaceLocalMulticast()) {
			// link-local multicast specified without interface gets expanded to listen on all interfaces
			expanded, err := expandLLMulticast(l)
			if err != nil {
				return nil, err
			}
			listeners = append(listeners, expanded...)
			continue
		}

		listeners = append(listeners, *l)
	}
	return listeners, nil
}

func (p *Parser) getPlugins(ver protocolVersion) ([]PluginConfig, error) {
	if err := protoVersionCheck(ver); err != nil {
		return nil, err
	}
	pluginList := cast.ToSlice(p.v.Get(fmt.Sprintf("server%d.plugins", ver)))
	if pluginList == nil {
		return nil, ConfigErrorFromString("dhcpv%d: invalid plugins section, not a list or no plugin specified", ver)
	}
	return parsePlugins(pluginList)
}

func (p *Parser) parseConfig(ver protocolVersion) error {
	if err := protoVersionCheck(ver); err != nil {
		return err
	}
	if exists := p.v.Get(fmt.Sprintf("server%d", ver)); exists == nil {
		// it is valid to have no server configuration defined
		return nil
	}
	// read plugin configuration
	plugins, err := p.getPlugins(ver)
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		p.logger.Printf("DHCPv%d: found plugin `%s` with %d args: %v", ver, plugin.Name, len(plugin.Args), plugin.Args)
	}

	listeners, err := p.parseListen(ver)
	if err != nil {
		return err
	}

	sc := ServerConfig{
		Addresses: listeners,
		Plugins:   plugins,
	}
	if ver == protocolV6 {
		p.config.Server6 = &sc
	} else if ver == protocolV4 {
		p.config.Server4 = &sc
	}
	return nil
}
