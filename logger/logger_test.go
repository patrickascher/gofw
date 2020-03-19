// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logger_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/file"
	"github.com/stretchr/testify/assert"
)

var mockLogger *mockProvider

type mockProvider struct {
	lock  sync.Mutex
	Entry logger.LogEntry
}

func (mp *mockProvider) Write(e logger.LogEntry) {
	mp.lock.Lock()
	mp.Entry = e
	mp.lock.Unlock()
}

func NewMockProvider() (*mockProvider, error) {
	mockLogger = &mockProvider{}
	return mockLogger, nil
}

// Register is testing
// - mandatory Config.Writer is missing
// - table drive test for LogLevel out of range (0-6 is allowed). 0 in that case will be set as TRACE.
func TestRegister(t *testing.T) {
	test := assert.New(t)
	mockProvider, err := NewMockProvider()
	assert.NoError(t, err) // only here for best practice

	// error: no writer is defined
	err = logger.Register(
		"mock", //log name
		logger.Config{},
	)
	test.Error(err)
	test.Equal(logger.ErrMandatoryWriter.Error(), err.Error())

	// table driven:
	// Testing LogLevel is undefined - default TRACE should be set
	// Testing Loglevel out of range - only 1-6 is allowed
	var tests = []struct {
		error   bool
		logType string
		config  logger.Config
	}{
		{error: false, logType: logger.TRACE.String(), config: logger.Config{Writer: mockProvider}},              // zero value, default Trace
		{error: false, logType: logger.TRACE.String(), config: logger.Config{Writer: mockProvider, LogLevel: 0}}, // zero value, default Trace
		{error: false, logType: logger.TRACE.String(), config: logger.Config{Writer: mockProvider, LogLevel: 1}},
		{error: false, logType: logger.DEBUG.String(), config: logger.Config{Writer: mockProvider, LogLevel: 2}},
		{error: false, logType: logger.INFO.String(), config: logger.Config{Writer: mockProvider, LogLevel: 3}},
		{error: false, logType: logger.WARNING.String(), config: logger.Config{Writer: mockProvider, LogLevel: 4}},
		{error: false, logType: logger.ERROR.String(), config: logger.Config{Writer: mockProvider, LogLevel: 5}},
		{error: false, logType: logger.CRITICAL.String(), config: logger.Config{Writer: mockProvider, LogLevel: 6}},
		{error: true, logType: "unknown log level", config: logger.Config{Writer: mockProvider, LogLevel: 7}},
	}
	for _, tt := range tests {
		t.Run(tt.logType, func(t *testing.T) {
			err = logger.Register(
				"mock", //log name
				tt.config,
			)
			if tt.error == true {
				test.Error(err)
			} else {
				test.NoError(err)
			}
		})
	}
}

// Register is testing if the different writer are getting set
func TestRegister_DifferentWriter(t *testing.T) {
	test := assert.New(t)
	// defining different log providers
	mockCommon := &mockProvider{}
	mockTrace := &mockProvider{}
	mockDebug := &mockProvider{}
	mockInfo := &mockProvider{}
	mockWarning := &mockProvider{}
	mockError := &mockProvider{}
	mockCritical := &mockProvider{}

	// error: no writer is defined
	err := logger.Register(
		"mock", //log name
		logger.Config{Writer: mockCommon,
			TraceWriter:    mockTrace,
			DebugWriter:    mockDebug,
			InfoWriter:     mockInfo,
			WarningWriter:  mockWarning,
			ErrorWriter:    mockError,
			CriticalWriter: mockCritical,
		},
	)
	test.NoError(err)

	// log messages
	l, err := logger.Get("mock")
	test.NoError(err)
	l.Trace("Trace")
	l.Debug("Debug")
	l.Info("Info")
	l.Warning("Warning")
	l.Error("Error")
	l.Critical("Critical")

	// check if the correct writer was used.
	// Writer: must be empty because it was never user
	test.Equal("", mockCommon.Entry.Message)
	// TraceWriter: must be empty because it was never user
	test.Equal("Trace", mockTrace.Entry.Message)
	// DebugWriter: must be empty because it was never user
	test.Equal("Debug", mockDebug.Entry.Message)
	// InfoWriter: must be empty because it was never user
	test.Equal("Info", mockInfo.Entry.Message)
	// WarningWriter: must be empty because it was never user
	test.Equal("Warning", mockWarning.Entry.Message)
	// ErrorWriter: must be empty because it was never user
	test.Equal("Error", mockError.Entry.Message)
	// CriticalWriter: must be empty because it was never user
	test.Equal("Critical", mockCritical.Entry.Message)
}

// Get checks if a log gets returned and if an error will return if the log name does not exist.
func TestGet(t *testing.T) {
	test := assert.New(t)
	mockProvider, err := NewMockProvider()
	test.NoError(err) // only here for best practice

	// setting mock again with a fresh config to avoid mistakes.
	err = logger.Register(
		"mock", //log name
		logger.Config{Writer: mockProvider},
	)

	// ok
	log, err := logger.Get("mock")
	test.NoError(err)
	test.Equal("*logger.Logger", reflect.TypeOf(log).String())

	// error: log does not exist
	log, err = logger.Get("mock2")
	test.Error(err)
	test.Equal(fmt.Sprintf(logger.ErrUnknownLogger.Error(), "mock2"), err.Error())
}

// Log is checking the following things:
// - Log for the LogLevels 1-6.
// - Checking if the LogEntry has the correct data. Timestamp, Linenumber and Filename have some minor checks.
// - On the second round the LogLevel is set to ERROR, it checks if the LogLevels getting skipped before.
func TestLogger_Log(t *testing.T) {
	test := assert.New(t)

	// startTime is used in the LogEntry to check if the Timestamp is after the startTime.
	startTime := time.Now()
	time.Sleep(100 * time.Millisecond)

	// define the log
	log, err := logger.Get("mock")
	test.NoError(err)

	// First round is the normal log for lvl 1-6
	// Second round the LogLevel is set to ERROR and CRITICAL only
	for i := 0; i < 2; i++ {

		// Reconfigure the log to log only ERROR and CRITICAL
		if i == 1 {
			mockProvider, err := NewMockProvider()
			test.NoError(err)
			err = logger.Register(
				"mock", //log name
				logger.Config{Writer: mockProvider, LogLevel: logger.ERROR},
			)
			test.NoError(err)
			log, err = logger.Get("mock")
			test.NoError(err)
		}

		// Table driven test
		var tests = []struct {
			msg  string
			args []interface{}
			fn   func(string, ...interface{})
		}{
			{msg: "TRACE", args: []interface{}{"arg1", "arg2"}, fn: log.Trace},
			{msg: "DEBUG", args: []interface{}{"arg0", "arg2"}, fn: log.Debug},
			{msg: "INFO", args: []interface{}{"arg1", "arg2"}, fn: log.Info},
			{msg: "WARNING", args: []interface{}{"arg1", "arg2"}, fn: log.Warning},
			{msg: "ERROR", args: []interface{}{"arg1", "arg2"}, fn: log.Error},
			{msg: "CRITICAL", args: []interface{}{"arg1", "arg2"}, fn: log.Critical},
		}
		for _, tt := range tests {
			t.Run(tt.msg, func(t *testing.T) {
				//log
				tt.fn(tt.msg, tt.args...)

				// checking everything on round 1
				// round two only ERROR and CRITICAL should get logged, all other levels should be empty
				if i == 0 || (i == 1 && (tt.msg == "ERROR" || tt.msg == "CRITICAL")) {
					test.Equal("logger_test.go", filepath.Base(mockLogger.Entry.Filename))
					test.Equal(tt.msg, mockLogger.Entry.Level.String())
					test.True(mockLogger.Entry.Line != 0)
					test.True(mockLogger.Entry.Timestamp.After(startTime))
					test.Equal(tt.msg, mockLogger.Entry.Message)
					test.Equal(tt.args, mockLogger.Entry.Arguments)
				} else {
					test.Equal("", mockLogger.Entry.Message)
				}
			})
		}
	}
}

// This example demonstrate the basics of the log.Interface.
// For more details check the documentation.
func Example() {

	// Register a new log with the name "access".
	fileLogger, err := file.New(file.Options{Filepath: "access.log"})
	if err != nil {
		//...
	}
	err = logger.Register(
		// log name
		"access",
		// The log should only log messages from the level WARNING and higher.
		// If the LogLevel is empty, it will start logging from TRACE.
		logger.Config{Writer: fileLogger, LogLevel: logger.WARNING},

		// Here is an example that everything should be logged in a file except CRITICAL, those should be emailed.
		// Each log level can have their own log provider.
		//
		// emailLogger = email.New(email.Options{...})
		// log.Config{Writer: logFile, CriticalWriter:emailLogger},
	)
	if err != nil {
		//...
	}

	// get the log
	log, err := logger.Get("access")
	if err != nil {
		//..
	}

	// log messages
	// The first parameter is the message it self. After that an unlimited number of arguments can follow.
	// It depends on the log provider how this is handled. Please check the provider documentation for more details.
	log.Info("User %v has successfully logged in", "John Doe") //This message will not be logged because the minimum log level is WARNING.
	log.Warning("User xy is locked because of too many login attempts")
}
