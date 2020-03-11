// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package file_test

import (
	"os"
	"testing"

	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/file"
)

// BenchmarkFile_Write is writing to all log levels
func BenchmarkFile_Write(b *testing.B) {
	// Register
	fileLogger, err := file.New(file.Options{Filepath: "benchmark.log"})
	if err != nil {
		b.Error(err)
	}
	err = logger.Register("benchmark", logger.Config{Writer: fileLogger})
	if err != nil {
		b.Error(err)
	}
	log, err := logger.Get("benchmark")
	if err != nil {
		b.Error(err)
	}

	b.Run("write logs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			log.Trace("Trace")
			log.Debug("Debug")
			log.Info("Info")
			log.Warning("Warning")
			log.Error("Error")
			log.Critical("Critical")
		}
	})

	err = os.Remove("benchmark.log")
	b.Log(err)
}
