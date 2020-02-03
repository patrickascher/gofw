// Package controller creates a simple handler for the go-router package.
// It supports the normal handler and the julienschmidt router handler.
package controller

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/cache"
	"net/http"
	"reflect"
)

// All error messages are defined here
var (
	defaultRender  = "json"
	ErrUnknownFunc = errors.New("controller method %#v does not exist in %#v")
)

// Controller struct
type Controller struct {
	ctx    *Context
	caller Interface

	// render
	renderType string

	// route controller info
	name                               string                       //name of the controller, used for error message
	patternHTTPMethodStructFuncMapping map[string]map[string]string //all existing httpMethods to controllerMethod (pattern,httpMethod,structMethod string)
	skipMethodChecks                   bool

	// cache
	cache cache.Cache
}

// Interface of the controller
type Interface interface {
	// controller funcs
	Initialize(Interface, map[string]map[string]string) error
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
	ServeHTTPJR(rw http.ResponseWriter, r *http.Request, p httprouter.Params)

	// Methods
	HTTPMethodsByPattern(string) map[string]string //needed by router
	functionByPatternAndHTTPMethod(string, string) (func(), error)
	setSkipFuncChecks(bool)

	// Context
	SetContext(ctx *Context)
	Context() *Context

	// Cache
	Cache() cache.Cache
	SetCache(cache.Cache)
	HasCache() bool

	// render types
	RenderType() string
	SetRenderType(string)

	// controller helpers
	Set(string, interface{})
	Error(int, string)
	Redirect(status int, url string)

	checkBrowserCancellation() bool
}

// SetCache sets the controller cache
func (c *Controller) SetCache(cache cache.Cache) {
	c.cache = cache
}

// Cache gets the controller cache
func (c *Controller) Cache() cache.Cache {
	return c.cache
}

// HasCache checks if a cache is defined
func (c *Controller) HasCache() bool {
	if c.cache == nil {
		return false
	}
	return true
}

// SetContext sets the controller context
func (c *Controller) SetContext(ctx *Context) {
	c.ctx = ctx
}

// Context returns the controller context
func (c *Controller) Context() *Context {
	return c.ctx
}

// Set controller variables
func (c *Controller) Set(key string, value interface{}) {
	c.Context().Response.AddData(key, value)
}

// isInitialized checks if a controller is initialized.
// For that the controller name gets checked.
func (c *Controller) isInitialized() bool {
	if c.name != "" {
		return true
	}
	return false
}

// setSkipFuncChecks is used to skip a struct function check.
func (c *Controller) setSkipFuncChecks(b bool) {
	c.skipMethodChecks = b
}

// Initialize the controller struct.
// It checks if the given HTTPMethod to StructMethod mapping exists.
// If not, an error will return. Also the controller name and render type will get set.
func (c *Controller) Initialize(caller Interface, httpMethodMapping map[string]map[string]string) error {

	if c.patternHTTPMethodStructFuncMapping == nil {
		c.patternHTTPMethodStructFuncMapping = make(map[string]map[string]string, 0)
	}

	for pattern, httpMethod := range httpMethodMapping {
		if !c.skipMethodChecks {
			for _, method := range httpMethod {
				_, err := getFunc(caller, method)
				if err != nil {
					return err
				}
			}
		}
		c.patternHTTPMethodStructFuncMapping[pattern] = httpMethod // a pattern is unique by the router
	}

	if !c.isInitialized() {
		c.caller = caller
		c.name = reflect.Indirect(reflect.ValueOf(c.caller)).Type().String()

		// calling user defined render type
		c.renderType = reflect.ValueOf(caller).Elem().FieldByName("renderType").String()

		if c.renderType == "" {
			c.renderType = defaultRender
		}
	}

	return nil
}

// getFunc is a helper to reflect the method of the caller controller.
// It will return an error, if the struct method does not exist.
func getFunc(caller Interface, name string) (func(), error) {
	methodVal := reflect.ValueOf(caller).MethodByName(name)
	if methodVal.IsValid() == false {
		return nil, fmt.Errorf(ErrUnknownFunc.Error(), name, reflect.Indirect(reflect.ValueOf(caller)).Type().String())
	}
	methodInterface := methodVal.Interface()
	method := methodInterface.(func())

	return method, nil
}

// HTTPMethodsByPattern returns a mapping of all existing HTTPMethods to struct method.
func (c *Controller) HTTPMethodsByPattern(pattern string) map[string]string {
	return c.patternHTTPMethodStructFuncMapping[pattern]
}

// functionByPatternAndHTTPMethod returns the struct function.
// If struct func does not exist, a error will return.
func (c *Controller) functionByPatternAndHTTPMethod(pattern string, HTTPMethod string) (func(), error) {
	//TODO error when not in map?
	return getFunc(c.caller, c.patternHTTPMethodStructFuncMapping[pattern][HTTPMethod])
}

// Redirect sets a HTTP location header and status code
func (c *Controller) Redirect(status int, url string) {
	http.Redirect(c.Context().Response.Raw(), c.Context().Request.Raw(), url, status)
}

// Error creates a HTTP error with the given code and message.
// If the render type is json, the error will be set as a controller variable.
func (c *Controller) Error(code int, msg string) {
	if c.renderType == "json" {
		c.Context().Response.Raw().WriteHeader(code)
		c.Set("error", msg)
		return
	}
	http.Error(c.Context().Response.Raw(), msg, code)
}

// RenderType will get returned
func (c *Controller) RenderType() string {
	return c.renderType
}

// SetRenderType of the controller
func (c *Controller) SetRenderType(s string) {
	c.renderType = s
}

// copyController creates a new instance of the controller itself.
func copyController(c *Controller) func() Interface {
	vc := reflect.New(reflect.TypeOf(c.caller).Elem())
	execController := vc.Interface().(Interface)
	return func() Interface {
		execController.setSkipFuncChecks(true)
		execController.SetRenderType(c.caller.RenderType())
		execController.SetCache(c.caller.Cache())

		execController.Initialize(execController, c.patternHTTPMethodStructFuncMapping)
		return execController
	}
}

/*
func (c *Controller) displayError(ctx *Context) {
	if err := recover(); err != nil {

		fmt.Println("Recovered in f", err)
		var stack string
		for i := 1; ; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			//Critical(fmt.Sprintf("%s:%d", file, line))
			stack = stack + fmt.Sprintln(fmt.Sprintf("%s:%d", file, line))

			ctx.Response.Get().Write([]byte(stack))
			ctx.Response.Get().WriteHeader(500)
		}
	}
}*/

// ServeHTTPJR for the julienschmidt router
func (c *Controller) ServeHTTPJR(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	c.serveHTTPLogic(rw, r, p)
}

// ServeHTTP for the normal http handler
func (c *Controller) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.serveHTTPLogic(rw, r)
}

// serveHTTPLogic contains the controller logic
func (c *Controller) serveHTTPLogic(args ...interface{}) {

	var w http.ResponseWriter
	var r *http.Request
	var p httprouter.Params
	for k, arg := range args {
		switch k {
		case 0:
			w = arg.(http.ResponseWriter)
		case 1:
			r = arg.(*http.Request)
		case 2:
			p = arg.(httprouter.Params)
		}
	}

	//create new instance per request
	newC := copyController(c)()
	newC.SetContext(NewContext(r, w))
	//if p != nil {
	newC.Context().Request.AddJulienSchmidtRouterParams(p)
	//}

	//TODO defer c.displayError(newC.Context())
	function, err := newC.functionByPatternAndHTTPMethod(newC.Context().Request.Pattern(), r.Method)
	if err == nil {
		function()
	} else {
		newC.Error(501, err.Error()) // can  be reached if pattern method mapping is wrong!
	}

	//checks if client is still here
	if newC.checkBrowserCancellation() {
		return
	}

	err = newC.Context().Response.Render(newC.RenderType())
	if err != nil {
		newC.Error(500, err.Error())
	}

}

// checkBrowserCancellation checking if the browser canceled the request
func (c *Controller) checkBrowserCancellation() bool {
	select {
	case <-c.Context().Request.Raw().Context().Done():
		c.Context().Response.Raw().WriteHeader(499)
		return true
	default:
	}
	return false
}
