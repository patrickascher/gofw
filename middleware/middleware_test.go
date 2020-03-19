// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofw/middleware"
	"github.com/stretchr/testify/assert"
)

// log test middleware.
func logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Logger-Before"))
		h(w, r)
		w.Write([]byte("Logger-After"))
	}
}

// auth test middleware.
func auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Auth-Before"))
		h(w, r)
		w.Write([]byte("Auth-After"))
	}
}

// TestNew tests the mw chain and if all middleware(s) were added correctly.
func TestNew(t *testing.T) {
	test := assert.New(t)

	// custom middleware
	controller := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Controller"))
	}

	// new middleware.chain
	r, _ := http.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()
	mw := middleware.New(logger, auth)
	mw.Handle(controller)(w, r)
	// testing the result
	test.Equal(2, len(mw.All()))
	test.Equal("Logger-BeforeAuth-BeforeControllerAuth-AfterLogger-After", w.Body.String())

	// new middleware.chain
	r, _ = http.NewRequest("GET", "/one", nil)
	w = httptest.NewRecorder()
	mw = middleware.New()
	mw.Add(auth, logger)
	mw.Handle(controller)(w, r)
	// testing the result
	test.Equal(2, len(mw.All()))
	test.Equal("Auth-BeforeLogger-BeforeControllerLogger-AfterAuth-After", w.Body.String())
}

// This example demonstrate the basics of the middleware chainer.
func Example() {
	// dummy middleware, reader, writer
	r, _ := http.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()
	controller := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Controller"))
	}

	// create a new middleware chain
	mw := middleware.New(logger)
	// add additional middleware(s) or define them already all at once middleware.New(log,auth)
	mw.Add(auth)
	// handle the middleware(s) in the order log,auth,controller
	mw.Handle(controller)(w, r)
}
