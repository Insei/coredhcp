// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This is a generated file, edits should be made in the corresponding source file
// And this file regenerated using `coredhcp-generator --from core-plugins.txt`
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/insei/coredhcp/config"
	"github.com/insei/coredhcp/logger"
	"github.com/insei/coredhcp/plugins"
	pl_dns "github.com/insei/coredhcp/plugins/dns"
	pl_file "github.com/insei/coredhcp/plugins/file"
	pl_leasetime "github.com/insei/coredhcp/plugins/leasetime"
	pl_mtu "github.com/insei/coredhcp/plugins/mtu"
	pl_nbp "github.com/insei/coredhcp/plugins/nbp"
	pl_netmask "github.com/insei/coredhcp/plugins/netmask"
	pl_prefix "github.com/insei/coredhcp/plugins/prefix"
	pl_range "github.com/insei/coredhcp/plugins/range"
	pl_router "github.com/insei/coredhcp/plugins/router"
	pl_searchdomains "github.com/insei/coredhcp/plugins/searchdomains"
	pl_serverid "github.com/insei/coredhcp/plugins/serverid"
	pl_sleep "github.com/insei/coredhcp/plugins/sleep"
	pl_staticroute "github.com/insei/coredhcp/plugins/staticroute"
	"github.com/insei/coredhcp/server"

	flag "github.com/spf13/pflag"
)

var (
	flagLogFile     = flag.StringP("logfile", "l", "", "Name of the log file to append to. Default: stdout/stderr only")
	flagLogNoStdout = flag.BoolP("nostdout", "N", false, "Disable logging to stdout/stderr")
	flagLogLevel    = flag.StringP("loglevel", "L", "info", fmt.Sprintf("Log level. One of %v", logger.GetLogLevelsStrings()))
	flagConfig      = flag.StringP("conf", "c", "default-server.config.yml", "Use this configuration file instead of the default location")
	flagPlugins     = flag.BoolP("plugins", "P", false, "list plugins")
)

var desiredPlugins = []*plugins.Plugin{
	&pl_dns.Plugin,
	&pl_file.Plugin,
	&pl_leasetime.Plugin,
	&pl_mtu.Plugin,
	&pl_nbp.Plugin,
	&pl_netmask.Plugin,
	&pl_prefix.Plugin,
	&pl_range.Plugin,
	&pl_router.Plugin,
	&pl_searchdomains.Plugin,
	&pl_serverid.Plugin,
	&pl_sleep.Plugin,
	&pl_staticroute.Plugin,
}

func main() {
	flag.Parse()

	if *flagPlugins {
		for _, p := range desiredPlugins {
			fmt.Println(p.Name)
		}
		os.Exit(0)
	}

	logBuilder := logger.NewDefaultLogrusBuilder()
	logLevel, err := logger.LogLevelFromString(*flagLogLevel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logBuilder.LogLevel(logLevel)
	if *flagLogFile != "" {
		logBuilder.WithFile(*flagLogFile)
	}
	if *flagLogNoStdout {
		logBuilder.WithNoStd()
	}
	log := logBuilder.Build()

	// register plugins
	for _, plugin := range desiredPlugins {
		if err := plugins.RegisterPlugin(log.WithField("prefix", "main"), plugin); err != nil {
			log.Fatalf("Failed to register plugin '%s': %v", plugin.Name, err)
		}
	}

	// parse config
	parser := config.NewParser(log)
	conf, err := parser.Parse(*flagConfig)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// start server
	srv, err := server.Start(log.WithField("prefix", conf.Name), conf)
	if err != nil {
		log.Fatal(err)
	}
	if err := srv.Wait(); err != nil {
		log.Print(err)
	}
	time.Sleep(time.Second)
}
