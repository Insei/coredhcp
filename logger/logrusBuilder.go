package logger

import (
	"io/ioutil"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

type logrusBuilder struct {
	logger *logrus.Logger
}

//NewLogrusBuilder creates new logger builder for logrus logger
func NewLogrusBuilder(logger *logrus.Logger) Builder {
	return &logrusBuilder{
		logger,
	}
}

//NewDefaultLogrusBuilder creates new logger builder for logrus logger
func NewDefaultLogrusBuilder() Builder {
	return &logrusBuilder{
		logrus.New(),
	}
}

// WithFile logs to the specified file in addition to the existing output.
func (b *logrusBuilder) WithFile(logfile string) Builder {
	b.logger.AddHook(lfshook.NewHook(logfile, &logrus.TextFormatter{}))
	return b
}

// WithNoStd disables logging to stdout/stderr.
func (b *logrusBuilder) WithNoStd() Builder {
	b.logger.SetOutput(ioutil.Discard)
	return b
}

// LogLevel sets logging level.
func (b *logrusBuilder) LogLevel(level LogLevel) Builder {
	var logLevelsFn = map[LogLevel]func(*logrus.Logger){
		None:    func(l *logrus.Logger) { l.SetOutput(ioutil.Discard) },
		Debug:   func(l *logrus.Logger) { l.SetLevel(logrus.DebugLevel) },
		Info:    func(l *logrus.Logger) { l.SetLevel(logrus.InfoLevel) },
		Warning: func(l *logrus.Logger) { l.SetLevel(logrus.WarnLevel) },
		Error:   func(l *logrus.Logger) { l.SetLevel(logrus.ErrorLevel) },
		Fatal:   func(l *logrus.Logger) { l.SetLevel(logrus.FatalLevel) },
	}
	fn, ok := logLevelsFn[level]
	if ok {
		fn(b.logger)
	}
	return b
}

//Build gets FieldLogger
func (b *logrusBuilder) Build() FieldLogger {
	b.logger.SetFormatter(GetLogrusFormatter())
	return &logrusFieldLogger{b.logger}
}
