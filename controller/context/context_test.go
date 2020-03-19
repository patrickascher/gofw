// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	"github.com/patrickascher/gofw/controller/context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeResponse struct {
	headers http.Header
	body    []byte
	status  int
}

func (r *FakeResponse) Header() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

func (r *FakeResponse) Write(body []byte) (int, error) {
	r.body = body
	return len(body), nil
}

func (r *FakeResponse) WriteHeader(status int) {
	r.status = status
}

// TestNew checks if the arguments are getting passed
func TestNew(t *testing.T) {
	req := &http.Request{}
	rw := &FakeResponse{}
	ctx := context.New(req, rw)

	assert.Equal(t, req, ctx.Request.Raw())
	assert.Equal(t, rw, ctx.Response.Raw())
}
