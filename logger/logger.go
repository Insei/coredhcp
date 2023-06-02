// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package logger is package with helpers, interfaces and implementations for coredhcp logging
package logger

//Logger interface that describes logger functionality
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

//FieldLogger logger interface that support filed logging
type FieldLogger interface {
	Logger
	WithField(key string, value interface{}) FieldLogger
}

// CreatePluginLogger returns a FieldLogger instance for the plugin
func CreatePluginLogger(serverLogger FieldLogger, pluginName string, ipv6 bool) FieldLogger {
	protocol := "v4"
	if ipv6 {
		protocol = "v6"
	}
	return serverLogger.WithField("plugin", pluginName).WithField("protocol", protocol)
}
