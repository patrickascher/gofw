// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package logger provides an alternative to the standard log library.
//
// Each LogLevels can have its own log provider. That means INFO can be logged in a file, ERROR is mailed and everything else will logged in the console.
// Different loggers with different log levels can be created. This means you can have a logger for an importer and a different logger for the application it self.
//
// The logger is easy to extend by implementing the logger.Interface.
package logger

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

// Log levels
const (
	TRACE level = iota + 1
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Error messages
var (
	ErrLogLevel        = errors.New("logger: LogLevel is unknown %#v")
	ErrMandatoryWriter = errors.New("logger: writer is mandatory")
	ErrUnknownLogger   = errors.New("logger: %v does not exist")
	errUnknownLogLevel = "unknown log level"
)

// registry for the defined logger.
// TODO check if ptr or value should be used
var registry map[string]*logger

// Interface is used by log providers.
type Interface interface {
	Write(LogEntry)
}

// Level - the higher the more critical
type level uint32

// String converts the level code.
func (lvl level) String() string {
	switch lvl {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	default:
		return errUnknownLogLevel
	}
}

// LogEntry is representing the actual logger message.
type LogEntry struct {
	Level     level
	Filename  string
	Line      int
	Timestamp time.Time
	Message   string
	Arguments []interface{}
}

// Config for the logger instance.
// Writer is mandatory, all others are optional.
// If the LogLevel is empty, TRACE will be set as default.
type Config struct {
	LogLevel       level
	Writer         Interface
	TraceWriter    Interface
	DebugWriter    Interface
	InfoWriter     Interface
	WarningWriter  Interface
	ErrorWriter    Interface
	CriticalWriter Interface
}

type logger struct {
	writer map[level]Interface
}

// setConfig for the logger.
// It skips the writer for lower log levels to safe memory.
// Checks if a specific logger is set, otherwise the default Writer is taken.
// Improvement: Set only the specific loggers if set, and dont set the default writer instead -> to safe memory - internal logic must be changed.
func (l *logger) setConfig(c Config) {

	//set default writer for all levels
	for _, lvl := range []level{TRACE, DEBUG, INFO, WARNING, ERROR, CRITICAL} {

		// skip writers if they are not needed
		if lvl < c.LogLevel {
			continue
		}

		// setting specific writers
		if c.TraceWriter != nil && lvl == TRACE {
			l.writer[lvl] = c.TraceWriter
			continue
		}
		if c.DebugWriter != nil && lvl == DEBUG {
			l.writer[lvl] = c.DebugWriter
			continue
		}
		if c.InfoWriter != nil && lvl == INFO {
			l.writer[lvl] = c.InfoWriter
			continue
		}
		if c.WarningWriter != nil && lvl == WARNING {
			l.writer[lvl] = c.WarningWriter
			continue
		}
		if c.ErrorWriter != nil && lvl == ERROR {
			l.writer[lvl] = c.ErrorWriter
			continue
		}
		if c.CriticalWriter != nil && lvl == CRITICAL {
			l.writer[lvl] = c.CriticalWriter
			continue
		}

		// setting writer to the default
		l.writer[lvl] = c.Writer
	}
}

// Register adds a new log provider to the registry or reconfigure it.
// If the name already exists, it will be overwritten.
func Register(name string, c Config) error {
	t := &logger{writer: make(map[level]Interface)}

	// Checking the config.
	// The main writer is mandatory.
	if c.Writer == nil {
		return ErrMandatoryWriter
	}

	// If no log level is set, Trace will be set as default
	if c.LogLevel == 0 {
		c.LogLevel = TRACE
	}
	// If the log level is out of range, an error will return.
	if c.LogLevel > 6 {
		return fmt.Errorf(ErrLogLevel.Error(), c.LogLevel)
	}

	// configure the logger
	t.setConfig(c)

	// adding the logger to the registry
	if registry == nil {
		registry = make(map[string]*logger)
	}
	registry[name] = t

	return nil
}

// Get the logger by its name.
// If the logger was not registered, an error will return.
func Get(name string) (*logger, error) {
	if _, ok := registry[name]; ok {
		return registry[name], nil
	}
	return nil, fmt.Errorf(ErrUnknownLogger.Error(), name)
}

// log calls the Writer.Write method
func (l *logger) log(lvl level, msg string, args ...interface{}) {

	// writer is not defined if the minimum log level is higher.
	if l.writer[lvl] == nil {
		return
	}

	// get file and line number of the parent caller.
	// If it was not possible to recover the information, the file string will be empty and line number will be 0.
	_, file, line, _ := runtime.Caller(2)

	//create the main Entry struct
	entry := LogEntry{
		Level:     lvl,
		Filename:  file,
		Line:      line,
		Timestamp: time.Now(),
		Message:   msg,
		Arguments: args,
	}

	//call the writer
	l.writer[lvl].Write(entry)
}

// Trace log message
func (l *logger) Trace(msg string, args ...interface{}) {
	l.log(TRACE, msg, args...)
}

// Debug log message
func (l *logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info log message
func (l *logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warning log message
func (l *logger) Warning(msg string, args ...interface{}) {
	l.log(WARNING, msg, args...)
}

// Error log message
func (l *logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Critical log message
func (l *logger) Critical(msg string, args ...interface{}) {
	l.log(CRITICAL, msg, args...)
}
