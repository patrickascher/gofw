package logger

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// TestFile_Write is testing if the File is created and has the correct content
func TestFile_Write(t *testing.T) {
	//create entry
	ts := time.Now()
	e := LogEntry{
		Level:     INFO,
		Filename:  "test.log",
		Line:      100,
		Timestamp: ts,
		Message:   "Hello World",
	}
	c := FileLogger(&FileOptions{File: "test.log"})
	c.Write(e)

	//sleep because its created in a goroutine
	time.Sleep(time.Duration(500) * time.Millisecond)

	//Read File content
	b, _ := ioutil.ReadFile("test.log")

	//compare
	assert.Equal(t, DefaultLoggingFormat(e)+"\n", string(b))

	//delete created file
	err := os.Remove("test.log")
	assert.NoError(t, err)
}
