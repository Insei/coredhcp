// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv6"
	"github.com/spf13/cast"
)

type protocolVersion int

const (
	protocolV6 protocolVersion = 6
	protocolV4 protocolVersion = 4
)

// Config holds the DHCPv6/v4 server configuration
type Config struct {
	Server6 *ServerConfig
	Server4 *ServerConfig
	Name    string
}

// New returns a new initialized instance of a Config object
func New() *Config {
	return &Config{}
}

// ServerConfig holds a server configuration that is specific to either the
// DHCPv6 server or the DHCPv4 server.
type ServerConfig struct {
	Addresses []net.UDPAddr
	Plugins   []PluginConfig
}

// PluginConfig holds the configuration of a plugin
type PluginConfig struct {
	Name string
	Args []string
}

func protoVersionCheck(v protocolVersion) error {
	if v != protocolV6 && v != protocolV4 {
		return fmt.Errorf("invalid protocol version: %d", v)
	}
	return nil
}

func parsePlugins(pluginList []interface{}) ([]PluginConfig, error) {
	plugins := make([]PluginConfig, 0, len(pluginList))
	for idx, val := range pluginList {
		conf := cast.ToStringMap(val)
		if conf == nil {
			return nil, ConfigErrorFromString("dhcpv6: plugin #%d is not a string map", idx)
		}
		// make sure that only one item is specified, since it's a
		// map name -> args
		if len(conf) != 1 {
			return nil, ConfigErrorFromString("dhcpv6: exactly one plugin per item can be specified")
		}
		var (
			name string
			args []string
		)
		// only one item, as enforced above, so read just that
		for k, v := range conf {
			name = k
			args = strings.Fields(cast.ToString(v))
			break
		}
		plugins = append(plugins, PluginConfig{Name: name, Args: args})
	}
	return plugins, nil
}

// BUG(Natolumin): listen specifications of the form `[ip6]%iface:port` or
// `[ip6]%iface` are not supported, even though they are the default format of
// the `ss` utility in linux. Use `[ip6%iface]:port` instead

// splitHostPort splits an address of the form ip%zone:port into ip,zone and port.
// It still returns if any of these are unset (unlike net.SplitHostPort which
// returns an error if there is no port)
func splitHostPort(hostport string) (ip string, zone string, port string, err error) {
	ip, port, err = net.SplitHostPort(hostport)
	if err != nil {
		// Either there is no port, or a more serious error.
		// Supply a synthetic port to differentiate cases
		var altErr error
		if ip, _, altErr = net.SplitHostPort(hostport + ":0"); altErr != nil {
			// Invalid even with a fake port. Return the original error
			return
		}
		err = nil
	}
	if i := strings.LastIndexByte(ip, '%'); i >= 0 {
		ip, zone = ip[:i], ip[i+1:]
	}
	return
}

func getListenAddress(addr string, ver protocolVersion) (*net.UDPAddr, error) {
	if err := protoVersionCheck(ver); err != nil {
		return nil, err
	}

	ipStr, ifname, portStr, err := splitHostPort(addr)
	if err != nil {
		return nil, ConfigErrorFromString("dhcpv%d: %v", ver, err)
	}

	ip := net.ParseIP(ipStr)
	if ipStr == "" {
		switch ver {
		case protocolV4:
			ip = net.IPv4zero
		case protocolV6:
			ip = net.IPv6unspecified
		default:
			panic("BUG: Unknown protocol version")
		}
	}
	if ip == nil {
		return nil, ConfigErrorFromString("dhcpv%d: invalid IP address in `listen` directive: %s", ver, ipStr)
	}
	if ip4 := ip.To4(); (ver == protocolV6 && ip4 != nil) || (ver == protocolV4 && ip4 == nil) {
		return nil, ConfigErrorFromString("dhcpv%d: not a valid IPv%d address in `listen` directive: '%s'", ver, ver, ipStr)
	}

	var port int
	if portStr == "" {
		switch ver {
		case protocolV4:
			port = dhcpv4.ServerPort
		case protocolV6:
			port = dhcpv6.DefaultServerPort
		default:
			panic("BUG: Unknown protocol version")
		}
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, ConfigErrorFromString("dhcpv%d: invalid `listen` port '%s'", ver, portStr)
		}
	}

	listener := net.UDPAddr{
		IP:   ip,
		Port: port,
		Zone: ifname,
	}
	return &listener, nil
}

// BUG(Natolumin): When listening on link-local multicast addresses without
// binding to a specific interface, new interfaces coming up after the server
// starts will not be taken into account.

func expandLLMulticast(addr *net.UDPAddr) ([]net.UDPAddr, error) {
	if !addr.IP.IsLinkLocalMulticast() && !addr.IP.IsInterfaceLocalMulticast() {
		return nil, errors.New("Address is not multicast")
	}
	if addr.Zone != "" {
		return nil, errors.New("Address is already zoned")
	}
	var needFlags = net.FlagMulticast
	if addr.IP.To4() != nil {
		// We need to be able to send broadcast responses in ipv4
		needFlags |= net.FlagBroadcast
	}

	ifs, err := net.Interfaces()
	ret := make([]net.UDPAddr, 0, len(ifs))
	if err != nil {
		return nil, fmt.Errorf("Could not list network interfaces: %v", err)
	}
	for _, iface := range ifs {
		if (iface.Flags & needFlags) != needFlags {
			continue
		}
		caddr := *addr
		caddr.Zone = iface.Name
		ret = append(ret, caddr)
	}
	if len(ret) == 0 {
		return nil, errors.New("No suitable interface found for multicast listener")
	}
	return ret, nil
}

func defaultListen(ver protocolVersion) ([]net.UDPAddr, error) {
	switch ver {
	case protocolV4:
		return []net.UDPAddr{{Port: dhcpv4.ServerPort}}, nil
	case protocolV6:
		l, err := expandLLMulticast(&net.UDPAddr{IP: dhcpv6.AllDHCPRelayAgentsAndServers, Port: dhcpv6.DefaultServerPort})
		if err != nil {
			return nil, err
		}
		l = append(l,
			net.UDPAddr{IP: dhcpv6.AllDHCPServers, Port: dhcpv6.DefaultServerPort},
			// XXX: Do we want to listen on [::] as default ?
		)
		return l, nil
	}
	return nil, errors.New("defaultListen: Incorrect protocol version")
}
