package logger

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// ConsoleOptions can be defined here
type ConsoleOptions struct {
	Color bool
}

type console struct {
	lock    sync.Mutex
	options *ConsoleOptions
}

//colorFormat for the logging output
func (c *console) colorFormat(e LogEntry) string {
	ts := e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")
	//get only the filename instead of the full path
	filename := filepath.Base(e.Filename)

	color := ""
	switch e.Level {
	case UNSPECIFIED, TRACE:
		color = "92" //green
	case DEBUG:
		color = "96" //Blue
	case INFO:
		color = "93" //Yellow
	case WARNING:
		color = "95" //Lila
	case ERROR, CRITICAL:
		color = "91" //Red
	}

	if e.Level == UNSPECIFIED {
		return fmt.Sprintf("%s %s:%d %s", ts, filename, e.Line, e.Message)
	}

	return fmt.Sprintf("%s \x1b["+color+"m%s\x1b[39m %s:%d %s", ts, e.Level.String(), filename, e.Line, e.Message)
}

// Write implements the writer interface of the logger.
// It creates an output with or without color
func (c *console) Write(e LogEntry) {
	c.lock.Lock()
	if c.options.Color {
		fmt.Println(c.colorFormat(e))
	} else {
		fmt.Println(DefaultLoggingFormat(e))
	}
	c.lock.Unlock()
}

// ConsoleLogger creates a Console Logger with the given Config.
// This is the entry point for this logger
func ConsoleLogger(options *ConsoleOptions) *console {
	c := console{}
	c.options = options

	return &c
}
