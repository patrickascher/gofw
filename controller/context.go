package controller

import (
	"github.com/patrickascher/gofw/controller/context"
	"net/http"
)

// Context is the controller context for the request (input) and response (output)
type Context struct {
	Request  *context.Request
	Response *context.Response
}

// NewContext creates a new context with the given request and response
func NewContext(req *http.Request, res http.ResponseWriter) *Context {
	return &Context{
		Request:  context.NewRequest(req),
		Response: context.NewResponse(res),
	}
}
