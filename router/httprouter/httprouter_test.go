// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package httprouter_test

import (
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
	"github.com/patrickascher/gofw/router/httprouter"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type TestController struct {
	controller.Controller
}

func (c *TestController) Get() {
	c.Set("foo", "bar")
	c.Set("url", c.Context().Request.Raw().Context().Value(router.PATTERN))
	c.Set("param", c.Context().Request.Raw().Context().Value(router.PARAMS))
}

func (c *TestController) GetMw() {
	c.Set("bar", "foo")
	c.Set("url", c.Context().Request.Raw().Context().Value(router.PATTERN))
	c.Set("param", c.Context().Request.Raw().Context().Value(router.PARAMS))
}

// Logger prints an info to the console
func Mw(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("middleware"))
		h(w, r)
	}
}

type NotFound struct {
}

// Logger prints an info to the console
func (nf *NotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("notFound"))
}

// TestHttpRouter_Handler creates a new testserver and test the added routes.
// Also the disallowed directory listing is tested
func TestHttpRouter_Handler(t *testing.T) {
	c := TestController{}

	r, err := router.New("httprouter", nil)
	assert.NoError(t, err)

	r.NotFound(&NotFound{})

	//Adding test route
	err = r.AddPublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	err = r.AddPublicRoute("/user/:id/:action", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	err = r.AddPublicRoute("/grid/*grid", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	//Adding test route with middleware
	mw := middleware.Chain{}
	err = r.AddPublicRoute("/mw", &c, router.RouteConfig{HTTPMethodToFunc: "get:GetMw", Middleware: mw.Add(Mw)})
	assert.NoError(t, err)

	//Adding favicon
	// create fav.ico
	_, err = os.Create("fav.ico")
	_, err = os.Create("test.log")

	assert.NoError(t, err)
	err = r.AddPublicFile("/favicon.ico", "fav.ico")
	assert.NoError(t, err)

	//Adding static folder
	err = r.AddPublicDir("/asserts", "../httprouter")
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())

	//request the url /
	resp, err := http.Get(server.URL + "/")
	body, error := ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\",\"param\":{},\"url\":\"/\"}", string(body[:]))

	//request the url /
	resp, err = http.Get(server.URL + "/user/12/delete")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\",\"param\":{\"action\":\"delete\",\"id\":\"12\"},\"url\":\"/user/:id/:action\"}", string(body[:]))

	//request the url /
	resp, err = http.Get(server.URL + "/grid/edit/12/export")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\",\"param\":{\"0\":\"edit\",\"1\":\"12\",\"2\":\"export\"},\"url\":\"/grid/*grid\"}", string(body[:]))

	resp, err = http.Get(server.URL + "/mw")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "middleware{\"bar\":\"foo\",\"param\":{},\"url\":\"/mw\"}", string(body[:]))

	//request the favicon
	favicon, _ := http.Get(server.URL + "/favicon.ico")
	assert.Equal(t, 200, favicon.StatusCode)
	//request the favicon over the static folder path
	asserts, err := http.Get(server.URL + "/asserts/test.log")
	assert.Equal(t, 200, asserts.StatusCode, err)

	//request a directory index
	assertsDirectoryListing, _ := http.Get(server.URL + "/asserts/")
	assert.Equal(t, 404, assertsDirectoryListing.StatusCode)

	//request an url which does not exist
	resp, err = http.Get(server.URL + "/notExisting")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 404, assertsDirectoryListing.StatusCode)
	assert.Equal(t, "notFound", string(body[:]))

	// delete created fav.ico
	os.Remove("fav.ico")
	os.Remove("test.log")

	defer server.Close()
}

func TestHttpRouter_Handler2(t *testing.T) {
	c := TestController{}

	r, err := router.New("httprouter", httprouter.Options{CatchAllKeyValuePair: true})
	assert.NoError(t, err)

	err = r.AddPublicRoute("/user/:id/:action", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	err = r.AddPublicRoute("/grid/*grid", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())

	//request the url /
	resp, err := http.Get(server.URL + "/user/12/delete/")
	body, error := ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\",\"param\":{\"action\":\"delete\",\"id\":\"12\"},\"url\":\"/user/:id/:action\"}", string(body[:]))

	// error: mismatch of params (edit:12, export:?)
	resp, err = http.Get(server.URL + "/grid/edit/12/export")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	// ok:
	resp, err = http.Get(server.URL + "/grid/mode/edit/id/12/")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\",\"param\":{\"id\":\"12\",\"mode\":\"edit\"},\"url\":\"/grid/*grid\"}", string(body[:]))

	defer server.Close()
}
