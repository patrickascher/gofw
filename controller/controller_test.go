package controller_test

import (
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/router"
	_ "github.com/patrickascher/gofw/router/backend/julienschmidt"
	"github.com/stretchr/testify/assert"
	"testing"

	"context"
	"github.com/patrickascher/gofw/cache"
	_ "github.com/patrickascher/gofw/cache/memory"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
)

type TestController struct {
	controller.Controller
}

func (c *TestController) Get() {
	c.Redirect(307, "/redirect")
	c.Set("link", "Get")
}

func (c *TestController) RedirectFunc() {
	c.Set("link", "RedirectFunc")
}

func (c *TestController) ErrorFunc() {
	c.Error(500, "Error message")
}

func (c *TestController) ErrorJSONFunc() {
	c.Error(500, "Error message")
}

type TestControllerTimeout struct {
	TestController
}

func (c *TestControllerTimeout) Timeout() {
	time.Sleep(1 * time.Second)
	c.Set("Get", true)
}

// Test SetCache, Cache, HasCache
func TestController_Cache(t *testing.T) {
	c := TestController{}
	assert.Equal(t, false, c.HasCache())
	assert.Equal(t, nil, c.Cache())

	mem, err := cache.Get("memory", 60*time.Second)
	assert.NoError(t, err)

	c.SetCache(mem)
	assert.Equal(t, mem, c.Cache())
	assert.Equal(t, true, c.HasCache())

}

func TestController_Initialize(t *testing.T) {

	c := TestController{}
	err := c.Initialize(&c, map[string]map[string]string{"/": {"GET": "XGet"}})
	assert.Error(t, err)

	err = c.Initialize(&c, map[string]map[string]string{"/": {"GET": "Get"}})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"GET": "Get"}, c.HTTPMethodsByPattern("/"))
	assert.Equal(t, "json", c.RenderType())
}

func TestController_HTTPMethodsByPattern(t *testing.T) {
	c := TestController{}
	err := c.Initialize(&c, map[string]map[string]string{"/": {"GET": "Get"}})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"GET": "Get"}, c.HTTPMethodsByPattern("/"))
}

// Test RenderType and SetRenderType
func TestController_RenderType(t *testing.T) {
	c := TestController{}
	assert.Equal(t, "", c.RenderType())
	c.SetRenderType("html")
	assert.Equal(t, "html", c.RenderType())
}

// Redirect and Set is tested here
func TestController_Redirect(t *testing.T) {
	c := TestController{}
	r, _ := router.Get("julienschmidt")
	//Adding test route
	err := r.PublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)
	err = r.PublicRoute("/redirect", &c, router.RouteConfig{HTTPMethodToFunc: "get:RedirectFunc", Middleware: nil})
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()
	//request the url
	resp, err := http.Get(server.URL + "/")
	assert.NoError(t, err)
	body, err2 := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err2)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "{\"link\":\"RedirectFunc\"}", string(body[:]))

}

// is also testing should test - ServeHTTP

func TestController_Error_HTML(t *testing.T) {
	c := TestController{}
	c.SetRenderType("html")
	r, _ := router.Get("julienschmidt")

	//Adding test route
	err := r.PublicRoute("/error", &c, router.RouteConfig{HTTPMethodToFunc: "get:ErrorFunc", Middleware: nil})
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	//request the url
	resp, err := http.Get(server.URL + "/error")
	body, error := ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "Error message\n{}", string(body[:]))
}

// is also testing ServeHTTPJR
func TestController_Error_JSON(t *testing.T) {
	c := TestController{}
	r, _ := router.Get("julienschmidt")

	//Adding test route
	err := r.PublicRoute("/errorJson", &c, router.RouteConfig{HTTPMethodToFunc: "get:ErrorJSONFunc", Middleware: nil})
	assert.NoError(t, err)

	//creating go test server
	server := httptest.NewServer(r.Handler())
	defer server.Close()

	//request the url
	resp, err := http.Get(server.URL + "/errorJson")
	body, error := ioutil.ReadAll(resp.Body)
	assert.NoError(t, error)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "{\"error\":\"Error message\"}", string(body[:]))

}

// TestController_ServeHTTPWithCancellation is testing if the server cancels the request between the callbacks and main function call (Before 1sec sleep, Main 1sec sleep, After 1sec sleep)
func TestController_ServeHTTPWithCancellation(t *testing.T) {
	//server
	c := TestControllerTimeout{}
	r, _ := router.Get("julienschmidt")
	err := r.PublicRoute("/test", &c, router.RouteConfig{HTTPMethodToFunc: "get:Get", Middleware: nil})
	assert.NoError(t, err)

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
