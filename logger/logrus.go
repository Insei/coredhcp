package logger

import (
	"fmt"

	log_prefixed "github.com/chappjc/logrus-prefix"
	"github.com/sirupsen/logrus"
)

type logrusFormatter struct {
	log_prefixed.TextFormatter
}

//GetLogrusFormatter returns logrus Formatter for setup logrus formatting in libraries
func GetLogrusFormatter() logrus.Formatter {
	return &logrusFormatter{
		TextFormatter: log_prefixed.TextFormatter{
			FullTimestamp: true,
		},
	}
}

func (f *logrusFormatter) Format(entry *logrus.Entry) ([]byte, error) {
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

type logrusFieldLogger struct {
	logrus.FieldLogger
}

func (l *logrusFieldLogger) WithField(key string, value interface{}) FieldLogger {
	return &logrusFieldLogger{l.FieldLogger.WithField(key, value)}
}
