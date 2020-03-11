// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package file implements the logger.Interface.
// The write process is done in a go routine. All operations are using a sync.RWMutex for synchronization.
//
// Check the file.Options for the available configurations.
//
// Benchmark file is available.
package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/patrickascher/gofw/logger"
)

// Error messages
var (
	ErrFilepath = errors.New("logger/file: option Filepath is mandatory")
)

// Options of the file log provider.
type Options struct {
	// The Filepath is mandatory.
	Filepath string
}

type file struct {
	lock    sync.Mutex
	options Options
}

// Write implements the logger.Interface.
// For the write process, a new go routine is spawned to avoid performance issues.
// TODO: how to handle errors, error on benchmark to delete the benchmark file?
func (c *file) Write(e logger.LogEntry) {
	go func(c *file, e logger.LogEntry) {
		c.lock.Lock()
		defer c.lock.Unlock()

		f, _ := os.OpenFile(c.options.Filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		f.WriteString(fmt.Sprintf("%s %s %s:%d %s", e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05"), e.Level.String(), filepath.Base(e.Filename), e.Line, e.Message) + "\n")
		f.Close()

	}(c, e)
}

// New creates a file log provider with the given options.
// If the option.Filepath is not set or the path does not exist, an error will return.
func New(options Options) (*file, error) {
	f := file{}
	f.options = options

	if f.options.Filepath == "" {
		return nil, ErrFilepath
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	// check if the given file exits/is createable.
	check, err := os.OpenFile(options.Filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &f, check.Close()
}
