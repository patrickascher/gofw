// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package file_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/file"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no Filepath is provided
	log, err := file.New(file.Options{})
	test.Error(err)
	test.Nil(log)

	// error: Filepath does not exist
	log, err = file.New(file.Options{Filepath: "/does/not/exist/test.log"})
	test.Error(err)
	test.Nil(log)

	// ok
	log, err = file.New(file.Options{Filepath: "test.log"})
	test.NoError(err)
	test.NotNil(log)

	// delete created file
	err = os.Remove("test.log")
	test.NoError(err)
}

// TestFile_Write is testing if the File is created and has the correct content
func TestFile_Write(t *testing.T) {
	test := assert.New(t)

	// create entry
	ts := time.Now()
	e := logger.LogEntry{
		Level:     logger.INFO,
		Filename:  "test.log",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}
	log, err := file.New(file.Options{Filepath: "test.log"})
	test.NoError(err)

	log.Write(e)

	// sleep because its created in a goroutine
	time.Sleep(time.Duration(500) * time.Millisecond)

	// Read File content
	b, _ := ioutil.ReadFile("test.log")

	// compare
	test.Equal(fmt.Sprintf("%s %s %s:%d %s", e.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05"), e.Level.String(), filepath.Base(e.Filename), e.Line, e.Message)+"\n", string(b))

	// delete created file
	// TODO error happens because of go routine.
	err = os.Remove("test.log")
	test.NoError(err)
}
