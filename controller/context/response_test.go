// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/controller/context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResponse_Data if the data is set to the variable
func TestResponse_Data(t *testing.T) {
	req := &http.Request{}
	rw := &FakeResponse{}
	ctx := context.New(req, rw)

	ctx.Response.SetData("user", "John Doe")

	// ok: data returns
	assert.Equal(t, "John Doe", ctx.Response.Data("user"))
	// ok: key does not exist
	assert.Equal(t, nil, ctx.Response.Data("password"))
}

// TestResponse_Data if the data is set to the variable
func TestResponse_Render(t *testing.T) {
	test := assert.New(t)

	req := &http.Request{}
	rw := &FakeResponse{}
	ctx := context.New(req, rw)

	headers := rw.Header()

	ctx.Response.SetData("user", "John Doe")
	err := ctx.Response.Render("json")
	test.NoError(err)

	test.Equal("{\"user\":\"John Doe\"}", string(rw.body))
	test.Equal([]string{"application/json"}, headers["Content-Type"])

	//render default
	ctx.Response.Render(controller.RenderHTML)
	test.Equal("{\"user\":\"John Doe\"}", string(rw.body)) // TODO at the moment only json is defined for everything

	//json marshal error
	ctx.Response.SetData("Err", make(chan int))
	err = ctx.Response.Render(controller.RenderJSON)
	assert.Error(t, err)
}
