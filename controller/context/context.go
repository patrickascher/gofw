// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package context provides a request and response struct which can be used in the controller.
// Request provides a lot of helper function and the Response offers a simple render function.
package context

import (
	"net/http"
)

// Context of the controller.
type Context struct {
	Request  *Request
	Response *Response
}

// New returns a context for the request and response.
func New(req *http.Request, res http.ResponseWriter) *Context {
	return &Context{
		Request:  newRequest(req),
		Response: newResponse(res),
	}
}
