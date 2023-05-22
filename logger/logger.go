// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package logger

import (
	"fmt"
	"io/ioutil"
	"sync"

	log_prefixed "github.com/chappjc/logrus-prefix"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var (
	globalLogger   *logrus.Logger
	getLoggerMutex sync.Mutex
)

type pluginFormatter struct {
	log_prefixed.TextFormatter
}

func (f *pluginFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	messagePrefix := ""
	server, ok := entry.Data["server"]
	if ok {
		messagePrefix += fmt.Sprintf("%s: ", server)
		delete(entry.Data, "server")
	}
	proto, ok := entry.Data["protocol"]
	if ok {
		messagePrefix += fmt.Sprintf("DHCP%s: ", proto)
		delete(entry.Data, "protocol")
	}
	plugin, ok := entry.Data["plugin"]
	if ok {
		messagePrefix += fmt.Sprintf("plugin: %s: ", plugin)
		delete(entry.Data, "plugin")
	}
	entry.Message = messagePrefix + entry.Message
	format, _ := f.TextFormatter.Format(entry)
	if plugin != nil {
		entry.Data["plugin"] = plugin
	}
	if proto != nil {
		entry.Data["protocol"] = proto
	}
	if server != nil {
		entry.Data["server"] = server
	}
	return format, nil
}

// GetLogger returns a configured logger instance
func GetLogger(prefix string) *logrus.Entry {
	if prefix == "" {
		prefix = "<no prefix>"
	}
	if globalLogger == nil {
		getLoggerMutex.Lock()
		defer getLoggerMutex.Unlock()
		logger := logrus.New()
		logger.SetFormatter(&pluginFormatter{
			TextFormatter: log_prefixed.TextFormatter{
				FullTimestamp: true,
			},
		})
		globalLogger = logger
	}
	return globalLogger.WithField("prefix", prefix)
}

// WithFile logs to the specified file in addition to the existing output.
func WithFile(log *logrus.Entry, logfile string) {
	log.Logger.AddHook(lfshook.NewHook(logfile, &logrus.TextFormatter{}))
}

// WithNoStdOutErr disables logging to stdout/stderr.
func WithNoStdOutErr(log *logrus.Entry) {
	log.Logger.SetOutput(ioutil.Discard)
}

// CreatePluginLogger returns a logger instance for the plugin
func CreatePluginLogger(serverLogger logrus.FieldLogger, pluginName string, ipv6 bool) logrus.FieldLogger {
	protocol := "v4"
	if ipv6 {
		protocol = "v6"
	}
	if serverLogger == nil {
		return GetLogger("default").WithField("plugin", pluginName).WithField("protocol", protocol)
	}
	return serverLogger.WithField("plugin", pluginName).WithField("protocol", protocol)
}
