package logger

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// init
func TestLogger_init(t *testing.T) {
	l, err := Get(CONSOLE)
	assert.NoError(t, err)
	assert.IsType(t, &Logger{}, l)
}

// setConfig
func TestLogger_setConfig(t *testing.T) {
	// default logger config
	consoleWriter := ConsoleLogger(&ConsoleOptions{Color: false})
	Register("testConfig", Config{Writer: consoleWriter})
	l, err := Get("testConfig")
	assert.NoError(t, err)
	assert.IsType(t, consoleWriter, l.Writer[UNSPECIFIED])
	assert.IsType(t, consoleWriter, l.Writer[TRACE])
	assert.IsType(t, consoleWriter, l.Writer[DEBUG])
	assert.IsType(t, consoleWriter, l.Writer[INFO])
	assert.IsType(t, consoleWriter, l.Writer[WARNING])
	assert.IsType(t, consoleWriter, l.Writer[ERROR])
	assert.IsType(t, consoleWriter, l.Writer[CRITICAL])

	// custom logger config
	consoleWriter = ConsoleLogger(&ConsoleOptions{Color: false})
	fileWriter := FileLogger(&FileOptions{File: ""})

	Register("testConfig", Config{
		Writer:            consoleWriter,
		UnspecifiedWriter: fileWriter,
		TraceWriter:       consoleWriter,
		DebugWriter:       fileWriter,
		InfoWriter:        consoleWriter,
		WarningWriter:     fileWriter,
		ErrorWriter:       consoleWriter,
		CriticalWriter:    fileWriter,
	})
	l, err = Get("testConfig")
	assert.NoError(t, err)
	assert.IsType(t, fileWriter, l.Writer[UNSPECIFIED])
	assert.IsType(t, consoleWriter, l.Writer[TRACE])
	assert.IsType(t, fileWriter, l.Writer[DEBUG])
	assert.IsType(t, consoleWriter, l.Writer[INFO])
	assert.IsType(t, fileWriter, l.Writer[WARNING])
	assert.IsType(t, consoleWriter, l.Writer[ERROR])
	assert.IsType(t, fileWriter, l.Writer[CRITICAL])
}

func TestLogger_log(t *testing.T) {
	Register(FILE,
		Config{
			Writer:   FileLogger(&FileOptions{File: "test.log"}),
			LogLevel: INFO,
		},
	)

	l, err := Get(FILE)
	assert.NoError(t, err)
	l.Unspecified("something")
	l.Trace("tracing")
	l.Debug("debugging")
	time.Sleep(100 * time.Millisecond)
	//no log should be written
	err = os.Remove("test.log")
	assert.Error(t, err)

	//log should be written
	l.Info("something\n")
	time.Sleep(100 * time.Millisecond)
	err = os.Remove("test.log")
	assert.NoError(t, err)

	//log should be written
	l.Warning("something")
	time.Sleep(100 * time.Millisecond)
	err = os.Remove("test.log")
	assert.NoError(t, err)

	//log should be written
	l.Error("something")
	time.Sleep(100 * time.Millisecond)
	err = os.Remove("test.log")
	assert.NoError(t, err)

	//log should be written
	l.Critical("something")
	time.Sleep(100 * time.Millisecond)
	err = os.Remove("test.log")
	assert.NoError(t, err)
}
