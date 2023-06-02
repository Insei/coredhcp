package logger

import "fmt"

//Builder is and interface for building logger instance. An additional level of abstraction over the logging setup.
type Builder interface {
	WithNoStd() Builder
	WithFile(file string) Builder
	LogLevel(level LogLevel) Builder
	Build() FieldLogger
}

//LogLevel log level
type LogLevel uint

const (
	//None level. Logging is disabled.
	None = LogLevel(iota)
	// Debug level. Usually only enabled when debugging. Very verbose logging.
	Debug
	// Info level. General operational entries about what's going on inside the
	// application.
	Info
	// Warning level. Non-critical entries that deserve eyes.
	Warning
	// Error level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	Error
	// Fatal level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	Fatal
)

var logLevels = map[string]LogLevel{
	"none":    None,
	"debug":   Debug,
	"info":    Info,
	"warning": Warning,
	"error":   Error,
	"fatal":   Fatal,
}

//GetLogLevelsStrings get available logger levels as strings
func GetLogLevelsStrings() []string {
	var levels []string
	for k := range logLevels {
		levels = append(levels, k)
	}
	return levels
}

//LogLevelFromString get LogLevel from the string
func LogLevelFromString(level string) (LogLevel, error) {
	logLevel, ok := logLevels[level]
	if !ok {
		return Debug, fmt.Errorf("invalid log level '%s', valid log levels are %v", level, GetLogLevelsStrings())
	}
	return logLevel, nil
}
