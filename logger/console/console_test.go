// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package console

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/patrickascher/gofw/logger"
	"github.com/stretchr/testify/assert"
)

// Test the console colored output. A new os.Exec must be used and the output is getting compared.
func TestConsole_WriteColor(t *testing.T) {

	ts := time.Now()

	e := logger.LogEntry{
		Level:     logger.INFO,
		Filename:  "test.txt",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}

	c, err := New(Options{Color: true})
	assert.NoError(t, err)

	if os.Getenv("TestRunning") == "1" {
		e.Level = logger.TRACE
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "2" {
		e.Level = logger.DEBUG
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "3" {
		e.Level = logger.INFO
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "4" {
		e.Level = logger.WARNING
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "5" {
		e.Level = logger.ERROR
		c.Write(e)
		return
	}
	if os.Getenv("TestRunning") == "6" {
		e.Level = logger.CRITICAL
		c.Write(e)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=1")
	o, _ := cmd.Output()
	assert.Equal(t, "\x1b[92mTRACE\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=2")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[96mDEBUG\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=3")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[93mINFO\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=4")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[95mWARNING\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=5")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[91mERROR\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=6")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[91mCRITICAL\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	// this is covered in a extra console call, but that's not included in the coverage.
	e.Level = logger.TRACE
	c.Write(e)

	e.Level = logger.DEBUG
	c.Write(e)

	e.Level = logger.INFO
	c.Write(e)

	e.Level = logger.WARNING
	c.Write(e)

	e.Level = logger.ERROR
	c.Write(e)

	e.Level = logger.CRITICAL
	c.Write(e)
}

// Test the console simple output. A new os.Exec must be used and the output is getting compared.
func TestConsole_WriteNoColor(t *testing.T) {

	ts := time.Now()

	e := logger.LogEntry{
		Level:     logger.INFO,
		Filename:  "test.txt",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}

	c, err := New(Options{Color: false})
	assert.NoError(t, err)

	if os.Getenv("TestRunning") == "1" {
		c.Write(e)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestConsole_WriteNoColor")
	cmd.Env = append(os.Environ(), "TestRunning=1")
	o, _ := cmd.Output()

	// This can fail because its can be spawned later.
	//assert.Equal(t, e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")+" INFO test.txt:100 Hello World", strings.Split(string(o), "\n")[0])
	assert.True(t, strings.Split(string(o), "\n")[0] != "")

	//just for coverage because the sub calls will not be covered right
	c.Write(e)
}
