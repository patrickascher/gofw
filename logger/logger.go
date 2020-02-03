// Package logger provides an alternative to the standard library log.
//
// Each LogLevels can have its own writer. That means INFO can be logged in a file, ERROR is mailed and everything else will logged in the console.
// The logger is easy to extend. At the moment a console and file writer comes out of the box
// See https://github.com/patrickascher/go-logger for more information and examples.
package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Default writer names
const (
	CONSOLE = "console"
	FILE    = "file"
)

// Log levels
const (
	UNSPECIFIED Level = iota + 1
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Level type - the higher the more critical!
type Level uint32

// String converts the level code.
func (level Level) String() string {
	switch level {
	case UNSPECIFIED:
		return ""
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
		return "<unknown>"
	}
}

// instance is the store for the log writers.
var writerStore map[string]*Logger

// once is used for the singleton.
var once sync.Once

// init initialize the default log writer.
func init() {
	writer := ConsoleLogger(&ConsoleOptions{Color: true})
	DefaultConfig := Config{Writer: writer, LogLevel: UNSPECIFIED}
	once.Do(func() {
		writerStore = make(map[string]*Logger)
		Register(CONSOLE, DefaultConfig)
	})
}

// Writer interface for the loggers.
type Writer interface {
	Write(LogEntry)
}

// LogEntry is representing the actual logger message.
type LogEntry struct {
	Level     Level
	Filename  string
	Line      int
	Timestamp time.Time
	Message   string
}

// DefaultLoggingFormat is a helper for the logging output.
func DefaultLoggingFormat(e LogEntry) string {
	ts := e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")
	//get only the filename instead of the full path
	filename := filepath.Base(e.Filename)
	if e.Level == UNSPECIFIED {
		return fmt.Sprintf("%s %s:%d %s", ts, filename, e.Line, e.Message)
	}
	return fmt.Sprintf("%s %s %s:%d %s", ts, e.Level.String(), filename, e.Line, e.Message)
}

// Config to register a new logger
type Config struct {
	Writer            Writer
	LogLevel          Level
	UnspecifiedWriter Writer
	TraceWriter       Writer
	DebugWriter       Writer
	InfoWriter        Writer
	WarningWriter     Writer
	ErrorWriter       Writer
	CriticalWriter    Writer
}

// Logger
type Logger struct {
	conf   Config
	Writer map[Level]Writer
}

// setConfig sets the config for the logger.
func (l *Logger) setConfig(c Config) {

	//set default writer for all levels
	for _, v := range []Level{UNSPECIFIED, TRACE, DEBUG, INFO, WARNING, ERROR, CRITICAL} {
		l.Writer[v] = l.conf.Writer
	}

	//check if there was a special writer for a level
	if c.UnspecifiedWriter != nil {
		l.Writer[UNSPECIFIED] = c.UnspecifiedWriter
	}
	if c.TraceWriter != nil {
		l.Writer[TRACE] = c.TraceWriter
	}
	if c.DebugWriter != nil {
		l.Writer[DEBUG] = c.DebugWriter
	}
	if c.InfoWriter != nil {
		l.Writer[INFO] = c.InfoWriter
	}
	if c.WarningWriter != nil {
		l.Writer[WARNING] = c.WarningWriter
	}
	if c.ErrorWriter != nil {
		l.Writer[ERROR] = c.ErrorWriter
	}
	if c.CriticalWriter != nil {
		l.Writer[CRITICAL] = c.CriticalWriter
	}

}

// Register adds a new log writer to the store or reconfigure it.
func Register(name string, c Config) {
	t := &Logger{conf: c}
	t.Writer = make(map[Level]Writer)
	t.setConfig(c)
	writerStore[name] = t
}

// Get the logger by its name. If the Logger does not get initialized an error is returned
func Get(name string) (*Logger, error) {
	if _, ok := writerStore[name]; ok {
		return writerStore[name], nil
	}
	return &Logger{}, fmt.Errorf("logger %v does not exist", name)
}

// log calls the Writer.Write method
func (l *Logger) log(lvl Level, msg string, args ...interface{}) {

	//don't log if log level it not allowed
	//this is tested in a different goroutine thats why there is no code coverage (TestLogger_skipLogLvl)
	if lvl < l.conf.LogLevel {
		return
	}

	//time
	timeNow := time.Now()

	//get
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}

	//Delete new line
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[0 : len(msg)-1]
	}

	//format string if there were any arguments
	fMsg := msg
	if len(args) > 0 {
		fMsg = fmt.Sprintf(msg, args...)
	}

	//create the main Entry struct
	entry := LogEntry{
		Level:     lvl,
		Filename:  file,
		Line:      line,
		Timestamp: timeNow,
		Message:   fMsg,
	}

	//call the writer
	l.Writer[lvl].Write(entry)
}

// Unspecified log message
func (l *Logger) Unspecified(msg string, args ...interface{}) {
	l.log(UNSPECIFIED, msg, args...)
}

// Trace log message
func (l *Logger) Trace(msg string, args ...interface{}) {
	l.log(TRACE, msg, args...)
}

// Debug log message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info log message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warning log message
func (l *Logger) Warning(msg string, args ...interface{}) {
	l.log(WARNING, msg, args...)
}

// Error log message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Critical log message
func (l *Logger) Critical(msg string, args ...interface{}) {
	l.log(CRITICAL, msg, args...)
}
