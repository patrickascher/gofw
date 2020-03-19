// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller_test

import (
	"context"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/router"
	"github.com/patrickascher/gofw/router/httprouter"
	_ "github.com/patrickascher/gofw/router/httprouter"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofw/cache"
	_ "github.com/patrickascher/gofw/cache/memory"
	"time"
)

type mockController struct {
	controller.Controller
}

func (c *mockController) Login() {
	c.Set("user", "John Doe")
	c.Redirect(301, "/redirect")
}

func (c *mockController) RedirectFunc() {
	c.Set("redirect", "successful")
}

func (c *mockController) ErrorFunc() {
	c.Error(500, "Error message")
}

type mockControllerTimeout struct {
	mockController
}

func (c *mockControllerTimeout) Timeout() {
	time.Sleep(1 * time.Second)
	c.Set("Successful", true)
}

// Test SetCache, Cache, HasCache
func TestController_Name(t *testing.T) {
	test := assert.New(t)

	c := mockController{}
	// ok: name is only visible after initialization
	test.Equal("", c.Name())

	// ok
	err := c.Initialize(&c, nil, false)
	test.NoError(err)
	test.Equal("controller_test.mockController", c.Name())

}

// Test SetCache, Cache, HasCache
func TestController_Cache(t *testing.T) {
	test := assert.New(t)

	c := mockController{}
	// ok: no cache is defined yet
	test.Equal(false, c.HasCache())
	test.Equal(nil, c.Cache())

	mem, err := cache.New(cache.MEMORY, nil)
	test.NoError(err)

	// set memory cache
	c.SetCache(mem)
	test.Equal(mem, c.Cache())
	test.Equal(true, c.HasCache())
}

// Test RenderType and SetRenderType
func TestController_RenderType(t *testing.T) {
	c := mockController{}
	assert.Equal(t, "", c.RenderType())
	c.SetRenderType("html")
	assert.Equal(t, "html", c.RenderType())
}

func TestController_Initialize(t *testing.T) {
	test := assert.New(t)

	c := mockController{}

	// error: controller function XLogin does not exist.
	err := c.Initialize(&c, map[string]map[string]string{"/": {"GET": "XLogin"}}, true)
	test.Error(err)

	// ok: controller function XLogin does not exist but method check is disabled
	err = c.Initialize(&c, map[string]map[string]string{"/": {"GET": "XLogin"}}, false)
	test.NoError(err)
	test.Equal(map[string]string{"GET": "XLogin"}, c.MappingBy("/"))

	// ok: Login method does exist
	err = c.Initialize(&c, map[string]map[string]string{"/": {"GET": "Login"}}, true)
	test.NoError(err)
	test.Equal(map[string]string{"GET": "Login"}, c.MappingBy("/"))
	test.Equal(controller.RenderJSON, c.RenderType())
}

func TestController_MappingBy(t *testing.T) {
	c := mockController{}
	err := c.Initialize(&c, map[string]map[string]string{"/": {"GET": "Login", "POST": "Login"}}, true)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"GET": "Login", "POST": "Login"}, c.MappingBy("/"))
}

// TestController_Redirect tests if the /login route will get redirected to /redirect.
// All the defined controller variables will be lost in login.
func TestController_Redirect(t *testing.T) {
	test := assert.New(t)

	c := mockController{}
	r, err := router.New(router.HTTPROUTER, httprouter.Options{CatchAllKeyValuePair: true})
	test.NoError(err)

	//Adding test route
	err = r.AddPublicRoute("/login", &c, router.RouteConfig{HTTPMethodToFunc: "get:Login", Middleware: nil})
	test.NoError(err)
	err = r.AddPublicRoute("/redirect", &c, router.RouteConfig{HTTPMethodToFunc: "get:RedirectFunc", Middleware: nil})
	test.NoError(err)

	//creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()
	//request the url
	resp, err := http.Get(server.URL + "/login")
	test.NoError(err)

	body, err2 := ioutil.ReadAll(resp.Body)
	test.NoError(err2)

	test.Equal(200, resp.StatusCode)
	test.Equal("{\"redirect\":\"successful\"}", string(body[:]))
}

func TestController_Error(t *testing.T) {
	test := assert.New(t)
	c := mockController{}
	c.SetRenderType(controller.RenderHTML)

	r, err := router.New(router.HTTPROUTER, httprouter.Options{CatchAllKeyValuePair: true})
	test.NoError(err)

	//Adding test route
	err = r.AddPublicRoute("/error", &c, router.RouteConfig{HTTPMethodToFunc: "get:ErrorFunc", Middleware: nil})
	test.NoError(err)

	//creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	resp, err := http.Get(server.URL + "/error")
	body, error := ioutil.ReadAll(resp.Body)
	test.NoError(error)
	test.NoError(err)
	test.Equal(500, resp.StatusCode)
	test.Equal("Error message\n{}", string(body[:]))

	c.SetRenderType(controller.RenderJSON)
	resp, err = http.Get(server.URL + "/error")
	body, error = ioutil.ReadAll(resp.Body)
	test.NoError(error)
	test.NoError(err)
	test.Equal(500, resp.StatusCode)
	test.Equal("{\"error\":\"Error message\"}", string(body[:]))
}

// TestController_ServeHTTPWithCancellation is testing if the server cancels the request between the callbacks and main function call (Before 1sec sleep, Main 1sec sleep, After 1sec sleep)
func TestController_ServeHTTPWithCancellation(t *testing.T) {
	test := assert.New(t)
	//server
	c := mockControllerTimeout{}

	r, err := router.New(router.HTTPROUTER, httprouter.Options{CatchAllKeyValuePair: true})
	test.NoError(err)

	err = r.AddPublicRoute("/timeout", &c, router.RouteConfig{HTTPMethodToFunc: "get:Timeout", Middleware: nil})
	test.NoError(err)

	server := httptest.NewServer(r.Handler())
	defer server.Close()

	//set requests with
	serverTimeout(500, server.URL) //canceled after "method" call

}

func serverTimeout(milliseconds time.Duration, server string) {
	//request
	cx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequest("GET", server+"/timeout", nil)
	req = req.WithContext(cx)

	ch := make(chan error)

	// Create the request
	go func() {
		_, err := http.DefaultClient.Do(req)
		select {
		case <-cx.Done():
		default:
			ch <- err
		}

	}()

	// Simulating user cancel request after the given time
	go func() {
		time.Sleep(milliseconds * time.Millisecond)
		cancel()
	}()
	select {
	case err := <-ch:
		if err != nil {
			// HTTP error
			panic(err)
		}
	case <-cx.Done():
		//cancellation
	}

}
