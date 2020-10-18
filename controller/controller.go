// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package controller provides a controller / action based handler for the router.
// A Controller can have different render types which write the controller data.
// If the controller is initialized by the router, a global cache will be passed through.
// Data, Redirects and Errors can be set directly in the controller.
//
// 		type AuthController struct {
//			controller.Controller
//		}
//
//		func(a *AuthController) Login()
//		{
//			// will return the controller name
//			name := a.Name()
//
//			// set a controller variable user - which will get rendered
//			a.Set("user", "John Doe")
//
//			// return the controller context.
//			// check context.request / context.response documentation.
//			ctx := a.Context()
//
//			// return the controller render type
//			rtype := a.RenderType()
//			// set the controller render type, default is JSON
//			a.SetRenderType(controller.RenderJSON)
//
//			// check if a controller cache is set
//			b := a.HasCache()
//			// get the controller cache
//			c := a.Cache()
//			// set a controller cache
//			// this is done automatic by the router if the router cache is set.
//			a.SetCache(nil)
//
//			// redirect to a different route
//			a.Redirect(301, "/forbidden")
//
//			// Set an error
//			// If the render type is JSON a json error key will be set.
//			a.Error(http.StatusUnauthorized, "you are not allowed")
//		}
package controller

import (
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/locale"
	"net/http"
	"reflect"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/controller/context"
)

// render types
const (
	RenderJSON = "json"
	RenderHTML = "html"
)

// Error messages
var (
	ErrMethodUnknown = errors.New("controller: method %#v does not exist in %#v")
)

// globalRenderType is the default value for all controllers.
var globalRenderType = RenderJSON

// Controller struct
type Controller struct {
	ctx           *context.Context
	caller        Interface
	renderType    string
	methodMapping map[string]map[string]string //map[url][HTTPMethod]ControllerMethod
	cache         cache.Interface

	actionName string
	localizer  locale.LocalizerI
}

// Interface of the controller.
type Interface interface {
	Initialize(caller Interface, mapping map[string]map[string]string, checkMethods bool) error
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
	MappingBy(pattern string) map[string]string //map[HTTPMethod]ControllerMethod

	// Context
	Context() *context.Context
	SetContext(ctx *context.Context)

	// render type
	RenderType() string
	SetRenderType(string)

	// controller helpers
	Set(string, interface{})
	Error(int, error)
	Redirect(status int, url string)
	Name() string
	Action() string

	// Translation
	T(string, ...map[string]interface{}) string
	TP(string, int, ...map[string]interface{}) string

	//Experimental
	ReadUserData(interface{}) error

	// internal helper
	checkBrowserCancellation() bool
	methodBy(pattern string, httpMethod string) (func(), error)
}

// Context returns the controller context.
func (c *Controller) Context() *context.Context {
	return c.ctx
}

// SetContext to the controller.
func (c *Controller) SetContext(ctx *context.Context) {
	c.ctx = ctx
}

// Set a controller variable by key and value.
// todo check if controller ctx is set...
func (c *Controller) Set(key string, value interface{}) {
	c.Context().Response.SetData(key, value)
}

// Initialize the controller.
// It checks if the given mapping is valid.
// Default information (name, caller, renderType) will be set.
// An error will return if the controller function does not exist.
func (c *Controller) Initialize(caller Interface, mapping map[string]map[string]string, checkMethods bool) error {

	if c.methodMapping == nil {
		c.methodMapping = make(map[string]map[string]string, len(mapping))
	}
	c.caller = caller

	// checking if the controller methods exist and merge the method mapping
	for pattern, hMethods := range mapping {
		c.methodMapping[pattern] = hMethods //pattern is unique
		if checkMethods {
			for hMethod := range hMethods {
				_, err := c.methodBy(pattern, hMethod)
				if err != nil {
					return err
				}
			}
		}
	}

	// if the name is not defined yet, some default values are set.
	c.renderType = caller.RenderType()
	if c.renderType == "" {
		c.renderType = globalRenderType
	}

	return nil
}

// MappingBy the given pattern returns a mapping of all defined HTTP methods to controller methods.
func (c *Controller) MappingBy(pattern string) map[string]string {
	return c.methodMapping[pattern]
}

// Redirect sets a HTTP location header and status code.
// On a redirect the old controller data will be lost.
func (c *Controller) Redirect(status int, url string) {
	http.Redirect(c.Context().Response.Raw(), c.Context().Request.Raw(), url, status)
}

// Error writes a HTTP error with the given code and message.
// If the render type is json, the error will be set as a controller variable.
func (c *Controller) Error(code int, err error) {
	if c.renderType == RenderJSON {
		c.Context().Response.Raw().WriteHeader(code)
		c.Set("error", err.Error())
		return
	}
	http.Error(c.Context().Response.Raw(), err.Error(), code)
}

// RenderType of the controller.
func (c *Controller) RenderType() string {
	return c.renderType
}

// SetRenderType of the controller.
func (c *Controller) SetRenderType(s string) {
	c.renderType = s
}

// ServeHTTP handler.
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// create new instance per request
	reqController := newController(c)
	reqController.SetContext(context.New(r, w))

	// TODO defer c.displayError(newC.Context())
	function, err := reqController.methodBy(reqController.Context().Request.Pattern(), r.Method)
	if err == nil {
		function()
	} else {
		reqController.Error(501, err) // can  be reached if pattern method mapping is wrong!
	}

	// checks if client is still here
	if reqController.checkBrowserCancellation() {
		return
	}

	// render the controller data
	err = reqController.Context().Response.Render(reqController.RenderType())
	if err != nil {
		reqController.Error(500, err)
	}
}

// name returns the controller name.
func (c *Controller) Name() string {
	if c.caller == nil {
		return ""
	}
	return reflect.Indirect(reflect.ValueOf(c.caller)).Type().String()
}

// name returns the controller name.
func (c *Controller) Action() string {
	return c.actionName
}

// ReadUserData reads the user into the given interface.
// TODO in the future this will return always the user data, does not matter if its jwt or another auth service.
func (c *Controller) ReadUserData(user interface{}) error {
	token := c.Context().Request.Token()
	if token == nil {
		return errors.New("controller: jwt claim is empty")
	}

	if reflect.TypeOf(user).Kind() != reflect.Ptr {
		return errors.New("controller: ReadUserData request a pointer as argument")

	}

	reflect.ValueOf(user).Elem().Set(reflect.ValueOf(token).Elem())

	return nil
}

func (c *Controller) T(name string, template ...map[string]interface{}) string {
	l := c.Context().Request.Localizer()
	if l == nil {
		return name
	}
	if v, err := l.Translate(name, template...); err == nil {
		return v
	}
	return name
}

func (c *Controller) TP(name string, count int, template ...map[string]interface{}) string {
	l := c.Context().Request.Localizer()
	if l == nil {
		return name
	}
	if v, err := l.TranslatePlural(name, count, template...); err == nil {
		return v
	}
	return name
}

// newController creates a new instance of the controller itself.
// the render type and cache will be passed from given controller.
// Initialize is called with the methodMapping. // TODO methodMapping could be passed as variable? Benchmarks?
func newController(c *Controller) Interface {
	vc := reflect.New(reflect.TypeOf(c.caller).Elem())
	execController := vc.Interface().(Interface)
	execController.SetRenderType(c.caller.RenderType())
	execController.Initialize(execController, c.methodMapping, false)
	return execController

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

// methodBy pattern and HTTP method will return the mapped controller method.
// Error will return if the controller method does not exist.
func (c *Controller) methodBy(pattern string, HTTPMethod string) (func(), error) {
	methodName := c.methodMapping[pattern][HTTPMethod]
	c.actionName = methodName
	methodVal := reflect.ValueOf(c.caller).MethodByName(methodName)
	if methodVal.IsValid() == false {
		return nil, fmt.Errorf(ErrMethodUnknown.Error(), methodName, reflect.Indirect(reflect.ValueOf(c.caller)).Type().String())
	}
	methodInterface := methodVal.Interface()
	method := methodInterface.(func())

	return method, nil
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

			x
		}
	}
}*/
