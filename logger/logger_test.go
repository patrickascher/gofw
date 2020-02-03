package logger_test

import (
	"fmt"
	"github.com/patrickascher/gofw/logger"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLevel_StringString(t *testing.T) {
	assert.Equal(t, "", logger.Level(1).String())
	assert.Equal(t, "TRACE", logger.Level(2).String())
	assert.Equal(t, "DEBUG", logger.Level(3).String())
	assert.Equal(t, "INFO", logger.Level(4).String())
	assert.Equal(t, "WARNING", logger.Level(5).String())
	assert.Equal(t, "ERROR", logger.Level(6).String())
	assert.Equal(t, "CRITICAL", logger.Level(7).String())
	assert.Equal(t, "<unknown>", logger.Level(999).String())
}

func TestDefaultLoggingFormat(t *testing.T) {
	ts := time.Now()

	entry := logger.LogEntry{}
	entry.Level = logger.TRACE
	entry.Timestamp = ts
	entry.Message = "Test"
	entry.Line = 100
	entry.Filename = "logger_test.go"

	assert.Equal(t, entry.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")+" TRACE logger_test.go:100 Test", logger.DefaultLoggingFormat(entry))

	// UNSPECIFIED is not logging the log-level
	entry.Level = logger.UNSPECIFIED
	assert.Equal(t, entry.Timestamp.In(time.UTC).Format("2006-01-02 15:04:05")+" logger_test.go:100 Test", logger.DefaultLoggingFormat(entry))
}

func TestRegisterAndGet(t *testing.T) {

	logger.Register(
		"test", //logger name
		logger.Config{
			Writer:   GetTestWriter(),
			LogLevel: logger.UNSPECIFIED,
		},
	)

	writer, err := logger.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, "*logger.Logger", reflect.TypeOf(writer).String())

	// Logger does not exist
	_, err = logger.Get("doesNotExist")
	assert.Error(t, err)
}

func TestLogger_Unspecified(t *testing.T) {

	TestWriter := GetTestWriter()
	logger.Register(
		"testLogger", //logger name
		logger.Config{
			Writer:     TestWriter, //default writer for all lvls
			LogLevel:   logger.UNSPECIFIED,
			InfoWriter: logger.FileLogger(&logger.FileOptions{File: "test.log"}), //custom writer for this logger-lvl
		},
	)

	log, err := logger.Get("testLogger")
	assert.NoError(t, err)

	log.Unspecified("Unspecified %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)

	log.Trace("Trace %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)

	log.Debug("Debug %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)

	//Info gets logged into a the test.log
	log.Info("Info %v", "log")
	// wait for the go routine to write in the file
	time.Sleep(500 * time.Millisecond)
	//read file
	file, err := os.OpenFile("test.log", os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer file.Close()
	// read file, line by line
	var text = make([]byte, 52)
	for {
		_, err = file.Read(text)

		// break if finally arrived at end of file
		if err == io.EOF {
			break
		}

		// break if error occurred
		if err != nil && err != io.EOF {
			assert.NoError(t, err)
		}
	}

	TestWriter.Entry.Level = logger.INFO
	TestWriter.Entry.Line = 90 // TODO better solution - this can fail as soon as the test gets edit
	TestWriter.Entry.Message = strings.Replace(TestWriter.Entry.Message, "Debug", "Info", 1)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), strings.TrimSpace(string(text)))

	//delete file - error if it does not exist
	err = os.Remove("test.log")
	assert.NoError(t, err)

	log.Warning("Warning %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)

	log.Error("Error %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)

	log.Critical("Critical %v", "log")
	fmt.Println(TestWriter.Body)
	assert.Equal(t, logger.DefaultLoggingFormat(TestWriter.Entry), TestWriter.Body)
}
