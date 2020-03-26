// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package console implements the log.Interface.
// All operations are using a sync.RWMutex for synchronization.
//
// Check the console.Options for the available configurations.
//
// Benchmark file is available.
package console

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/patrickascher/gofw/logger"
)

// Options of the console log provider.
type Options struct {
	// adding some highlights to the console
	Color bool
}

type console struct {
	lock    sync.Mutex
	options Options
}

// colorFormat for the logging output.
func (c *console) colorFormat(e logger.LogEntry) string {
	color := ""
	switch e.Level {
	case logger.TRACE:
		color = "92" //green
	case logger.DEBUG:
		color = "96" //Blue
	case logger.INFO:
		color = "93" //Yellow
	case logger.WARNING:
		color = "95" //Lila
	case logger.ERROR, logger.CRITICAL:
		color = "91" //Red
	}

	return fmt.Sprintf("%s \x1b["+color+"m%s\x1b[39m %s:%d %s", e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05"), e.Level.String(), filepath.Base(e.Filename), e.Line, e.Message)
}

// Write implements the writer interface of the log.Interface.
// It creates a simple console output.
func (c *console) Write(e logger.LogEntry) {
	c.lock.Lock()
	if c.options.Color {
		fmt.Println(c.colorFormat(e), e.Arguments)
	} else {
		fmt.Println(fmt.Sprintf("%s %s %s:%d %s", e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05"), e.Level.String(), filepath.Base(e.Filename), e.Line, e.Message), e.Arguments)
	}
	c.lock.Unlock()
}

// New creates a console log provider with the given options.
func New(options Options) (logger.Interface, error) {
	c := console{}
	c.options = options
	return &c, nil
}
