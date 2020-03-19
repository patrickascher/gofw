// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package log_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/file"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/middleware/log"
	"github.com/stretchr/testify/assert"
)

// Test the middleware.
// TODO check the result more percise as soon as the logger.file works correctly.
func TestLog_MW(t *testing.T) {
	test := assert.New(t)

	fileWriter, err := file.New(file.Options{Filepath: "access.log"})
	test.NoError(err)

	err = logger.Register("file", logger.Config{Writer: fileWriter})
	test.NoError(err)

	logger, err := logger.Get("file")
	test.NoError(err)

	log := log.New(logger)

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("new Writer"))
	}

	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	w := httptest.NewRecorder()

	mw := middleware.New(log.MW)
	mw.Handle(handlerFunc)(w, r)

	assert.Equal(t, "new Writer", w.Body.String())

	b, err := ioutil.ReadFile("access.log")
	test.NoError(err)
	test.Contains(string(b), "GET /path")

	err = os.Remove("access.log")
	test.NoError(err)
}
