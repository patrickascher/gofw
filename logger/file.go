package logger

import (
	"os"
	"sync"
)

// FileOptions can be defined here
type FileOptions struct {
	File string
}

type file struct {
	lock    sync.Mutex
	options *FileOptions
}

// Write implements the writer interface of the logger.
// It creates a new goroutine to avoid performance problems
func (c *file) Write(e LogEntry) {
	go func(c *file, e LogEntry) {
		c.lock.Lock()
		defer c.lock.Unlock()

		f, err := os.OpenFile(c.options.File, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err = f.WriteString(DefaultLoggingFormat(e) + "\n"); err != nil {
			panic(err)
		}
	}(c, e)
}

// FileLogger creates a File Logger with the given Config.
// This is the entry point for this logger
func FileLogger(options *FileOptions) *file {
	f := file{}
	f.options = options

	return &f
}
