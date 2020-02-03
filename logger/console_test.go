package logger

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConsole_WriteColor(t *testing.T) {

	ts := time.Now()

	e := LogEntry{
		Level:     INFO,
		Filename:  "test.txt",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}

	c := ConsoleLogger(&ConsoleOptions{Color: true})
	assert.Equal(t, true, c.options.Color)

	if os.Getenv("TestRunning") == "1" {
		e.Level = UNSPECIFIED
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "2" {
		e.Level = TRACE
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "3" {
		e.Level = DEBUG
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "4" {
		e.Level = INFO
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "5" {
		e.Level = WARNING
		c.Write(e)
		return
	}

	if os.Getenv("TestRunning") == "6" {
		e.Level = ERROR
		c.Write(e)
		return
	}
	if os.Getenv("TestRunning") == "7" {
		e.Level = CRITICAL
		c.Write(e)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=1")
	o, _ := cmd.Output()
	assert.Equal(t, "test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=2")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[92mTRACE\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=3")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[96mDEBUG\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=4")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[93mINFO\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=5")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[95mWARNING\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=6")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[91mERROR\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	cmd = exec.Command(os.Args[0], "-test.run=TestConsole_WriteColor")
	cmd.Env = append(os.Environ(), "TestRunning=7")
	o, _ = cmd.Output()
	assert.Equal(t, "\x1b[91mCRITICAL\x1b[39m test.txt:100 Hello World", strings.Split(string(o), "\n")[0][20:])

	// this is covered in a extra console call, but that's not included in the coverage.
	e.Level = UNSPECIFIED
	c.Write(e)

	e.Level = TRACE
	c.Write(e)

	e.Level = DEBUG
	c.Write(e)

	e.Level = INFO
	c.Write(e)

	e.Level = WARNING
	c.Write(e)

	e.Level = ERROR
	c.Write(e)

	e.Level = CRITICAL
	c.Write(e)
}

func TestConsole_WriteNoColor(t *testing.T) {

	ts := time.Now()

	e := LogEntry{
		Level:     INFO,
		Filename:  "test.txt",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}

	c := ConsoleLogger(&ConsoleOptions{Color: false})
	assert.Equal(t, false, c.options.Color)

	if os.Getenv("TestRunning") == "1" {
		c.Write(e)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestConsole_WriteNoColor")
	cmd.Env = append(os.Environ(), "TestRunning=1")
	o, _ := cmd.Output()

	assert.Equal(t, e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")+" INFO test.txt:100 Hello World", strings.Split(string(o), "\n")[0])
	//just for coverage because the sub calls will not be covered right
	c.Write(e)
}
