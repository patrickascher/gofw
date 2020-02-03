// Copyright 2018 (pat@fullhouse-productions.com)
// TODO check license styles
package julienschmidt

import (
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
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
}

func (c *TestController) GetMw() {
	c.Set("bar", "foo")
}

// Logger prints an info to the console
func Mw(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Write([]byte("middleware"))
		h(w, r, ps)
	}
}

type NotFound struct {
}

// Logger prints an info to the console
func (nf *NotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("notFound"))
}

// TestJulienschmidt_GetHandler creates a new testserver and test the added routes.
// Also the disallowed directory listing is tested
// route / without a middleware
// rout /mw with a middleware
// favicon.ico and asserts/favicon.ico
func TestJulienschmidt_GetHandler(t *testing.T) {
	c := TestController{}
	c.SetRenderType("html")

	r, _ := router.Get("julienschmidt")

	r.NotFound(&NotFound{})

	//Adding test route
	err := r.PublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

	//Adding test route with middleware
	mw := middleware.ChainJR{}
	err = r.PublicRoute("/mw", &c, router.RouteConfig{HTTPMethodToFunc: "get:GetMw", Middleware: mw.Add(Mw)})
	assert.NoError(t, err)

	//Adding favicon
	// create fav.ico
	_, err = os.Create("fav.ico")
	assert.NoError(t, err)
	err = r.Favicon("fav.ico")
	assert.NoError(t, err)

	//Adding static folder
	err = r.PublicDir("asserts", "../julienschmidt")
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())

	//request the url /
	resp, err := http.Get(server.URL + "/")
	body, error := ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"foo\":\"bar\"}", string(body[:]))

	resp, err = http.Get(server.URL + "/mw")
	body, error = ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "middleware{\"bar\":\"foo\"}", string(body[:]))

	//request the favicon
	favicon, _ := http.Get(server.URL + "/favicon.ico")
	assert.Equal(t, 200, favicon.StatusCode)

	//request the favicon over the static folder path
	asserts, _ := http.Get(server.URL + "/asserts/fav.ico")
	assert.Equal(t, 200, asserts.StatusCode)

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

	defer server.Close()
}
