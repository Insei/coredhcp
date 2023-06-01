# coredhcp

[![Build](https://github.com/insei/coredhcp/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/insei/coredhcp/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/insei/coredhcp/branch/main/graph/badge.svg)](https://codecov.io/gh/insei/coredhcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/insei/coredhcp)](https://goreportcard.com/report/github.com/insei/coredhcp)

Fast, multithreaded, modular and extensible DHCP server written in Go

This is still a work-in-progress

## Example configuration

In CoreDHCP almost everything is implemented as a plugin. The order of plugins in the configuration matters: every request is evaluated calling each plugin in order, until one breaks the evaluation and responds to, or drops, the request.

The following configuration runs a DHCPv6-only server, listening on all the interfaces, using a custom server ID and DNS, and reading the leases from a text file.

```
server6:
    # this server will listen on all the available interfaces, on the default
    # DHCPv6 server port, and will join the default multicast groups. For more
    # control, see the `listen` directive in cmds/coredhcp/config.yml.example .
    plugins:
        - server_id: LL 00:de:ad:be:ef:00
        - file: "leases.txt"
        - dns: 8.8.8.8 8.8.4.4 2001:4860:4860::8888 2001:4860:4860::8844
```

For more complex examples, like how to listen on specific interfaces and
configure other plugins, see [default-server.config.yml.example](cmds/coredhcp/default-server.config.yml.example).

## Build and run

An example server is located under [cmds/coredhcp/](cmds/coredhcp/), so enter that
directory first. To build a server with a custom set of plugins, see the "Server
with custom plugins" section below.

Once you have a working configuration in `default-server.config.yml` (see [default-server.config.yml.example](cmds/coredhcp/default-server.config.yml.example)), you can build and run the server:
```
$ cd cmds/coredhcp
$ go build
$ sudo ./coredhcp
[2023-05-22T11:58:27+03:00]  INFO main: Setting log level to 'info'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'dns'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'file'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'lease_time'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'mtu'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'nbp'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'netmask'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'prefix'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'range'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'router'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'searchdomains'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'server_id'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'sleep'
[2023-05-22T11:58:27+03:00]  INFO main: Registering plugin 'staticroute'
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: Loading configuration
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv6: found plugin `server_id` with 2 args: [LL 00:de:ad:be:ef:00]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv6: found plugin `file` with 1 args: [leases.txt]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv6: found plugin `dns` with 2 args: [2001:4860:4860::8888 2001:4860:4860::8844]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv6: found plugin `nbp` with 1 args: [http://[2001:db8:a::1]/nbp]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv6: found plugin `prefix` with 2 args: [2001:db8::/48 64]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `lease_time` with 1 args: [3600s]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `server_id` with 1 args: [10.10.10.1]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `dns` with 2 args: [8.8.8.8 8.8.4.4]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `router` with 1 args: [192.168.1.1]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `netmask` with 1 args: [255.255.255.0]
[2023-05-22T11:58:27+03:00]  INFO config-parser: default-server: DHCPv4: found plugin `range` with 4 args: [leases.txt 10.10.10.100 10.10.10.200 60s]
[2023-05-22T11:58:27+03:00]  INFO default-server: Loading plugins...
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: loading plugin `server_id`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: plugin: server_id: loading `server_id` plugin for DHCPv6 with args: [LL 00:de:ad:be:ef:00]
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: plugin: server_id: using ll 00:de:ad:be:ef:00
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: loading plugin `file`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: plugin: file: reading leases from leases.txt
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: plugin: file: loaded 0 leases from leases.txt
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: loading plugin `dns`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: plugin: dns: loaded 2 DNS servers.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: loading plugin `nbp`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: nbp: loaded NBP plugin for DHCPv6.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv6: loading plugin `prefix`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `lease_time`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: lease_time: loading `lease_time` plugin
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `server_id`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: server_id: loading `server_id` plugin for DHCPv4 with args: [10.10.10.1]
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `dns`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: dns: loaded plugin for DHCPv4.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: dns: loaded 2 DNS servers.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `router`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: router: Loaded plugin for DHCPv4.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: router: loaded 1 router IP addresses.
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `netmask`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: netmask: loaded plugin for DHCPv4
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: netmask: loaded client netmask
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: loading plugin `range`
[2023-05-22T11:58:27+03:00]  INFO default-server: DHCPv4: plugin: range: Loaded 0 DHCPv4 leases from leases.txt
[2023-05-22T11:58:27+03:00]  INFO default-server: Starting DHCPv6 server
[2023-05-22T11:58:27+03:00]  INFO default-server: Starting DHCPv4 server
[2023-05-22T11:58:27+03:00]  INFO default-server: Listen 0.0.0.0:67
[2023-05-22T11:58:27+03:00]  INFO default-server: Listen [::]:547
...
```

Then try it with the local test client, that is located under
[cmds/client/](cmds/client):
```
$ cd cmds/client
$ go build
$ sudo ./client
INFO[2019-01-05T22:29:21Z] &{ReadTimeout:3s WriteTimeout:3s LocalAddr:[::1]:546 RemoteAddr:[::1]:547}
INFO[2019-01-05T22:29:21Z] DHCPv6Message
  messageType=SOLICIT
  transactionid=0x6d30ff
  options=[
    OptClientId{cid=DUID{type=DUID-LLT hwtype=Ethernet hwaddr=00:11:22:33:44:55}}
    OptRequestedOption{options=[DNS Recursive Name Server, Domain Search List]}
    OptElapsedTime{elapsedtime=0}
    OptIANA{IAID=[250 206 176 12], t1=3600, t2=5400, options=[]}
  ]
...
```

# Plugins

CoreDHCP is heavily based on plugins: even the core functionalities are
implemented as plugins. Therefore, knowing how to write one is the key to add
new features to CoreDHCP.

Core plugins can be found under the [plugins](/plugins/) directory. Additional
plugins can also be found in the
[coredhcp/plugins](https://github.com/insei/plugins) repository.

## Server with custom plugins

To build a server with a custom set of plugins you can use the
[coredhcp-generator](/cmds/coredhcp-generator/) tool. Head there for
documentation on how to use it.

# How to write a plugin

The best way to learn is to read the comments and source code of the
[example plugin](plugins/example/), which guides you through the implementation
of a simple plugin that prints a packet every time it is received by the server.


# Authors

* [Andrea Barberio](https://github.com/insomniacslk)
* [Anatole Denis](https://github.com/natolumin)
* [Pablo Mazzini](https://github.com/pmazzini)
