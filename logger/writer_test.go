package logger_test

import (
	"github.com/patrickascher/gofw/logger"
	"sync"
)

// Test logger is used for testing purposes
type TestWriter struct {
	lock  sync.Mutex
	Body  string
	Entry logger.LogEntry
}

func (t *TestWriter) Write(e logger.LogEntry) {
	t.lock.Lock()
	t.Entry = e
	t.Body = logger.DefaultLoggingFormat(e)
	t.lock.Unlock()
}

func GetTestWriter() *TestWriter {
	t := TestWriter{}
	return &t
}
