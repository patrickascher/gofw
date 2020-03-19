// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package log logs every request. Every log provider can be used which implements the log.Interface.
// The log information is remoteAddr, HTTP Method, URL, Proto, HTTP Status, Response size and requested time.
//
//		logWriter := log.Get("console")
// 		log := New(logWriter)
// 		middleware.Add(log.MW)
package log

import (
	"fmt"
	"net/http"
	"time"

	"github.com/patrickascher/gofw/logger"
)

// Log
type Log struct {
	write *logger.Logger
}

// New returns a new Log instance.
func New(logger *logger.Logger) *Log {
	return &Log{write: logger}
}

// MW will be passed to the middleware.
func (l *Log) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// wrapped response writer to fetch the size and status.
		wrw := newResponseWriter(w)

		h(wrw, r)

		// log
		l.write.Info(fmt.Sprintf("(%s) %s %s %s %d %d %s", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, wrw.status, wrw.size, time.Since(start)))
	}
}

// responseWriter is a custom response writer to read the size and HTTP code.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader is adding the HTTP status of the response to the responseWriter struct.
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write is adding the size of the response to the responseWriter struct.
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// newResponseWriter returns a new responseWriter struct.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         200,
	}
}
