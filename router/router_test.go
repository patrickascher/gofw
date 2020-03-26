// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	js "github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/cache"
	_ "github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
	"github.com/patrickascher/gofw/router/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	test := assert.New(t)

	// table driven:
	var tests = []struct {
		test     string
		provider string
		fn       func(interface{}) router.Interface
		error    bool
		errorMsg string
	}{
		{test: "no provider and no name", provider: "", fn: nil, error: true, errorMsg: router.ErrNoProvider.Error()},
		{test: "no provider", provider: "mock", fn: nil, error: true, errorMsg: router.ErrNoProvider.Error()},
		{test: "no name", provider: "", fn: newMock, error: true, errorMsg: router.ErrNoProvider.Error()},
		{test: "mandatory fields pok", provider: "mock", fn: newMock, error: false, errorMsg: ""},
		{test: "register twice", provider: "mock", fn: newMock, error: true, errorMsg: fmt.Sprintf(router.ErrProviderAlreadyExists.Error(), "mock")},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			err := router.Register(tt.provider, tt.fn)
			if tt.error == true {
				test.Error(err)
				test.Equal(tt.errorMsg, err.Error())
			} else {
				test.NoError(err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	// error: no registered mock2 cache provider
	m, err := router.New("mock2", nil)
	test.Error(err)
	test.Equal(fmt.Sprintf(router.ErrUnknownProvider.Error(), "mock2"), err.Error())

	// ok
	m, err = router.New("mock", nil)
	test.NoError(err)
	test.NotNil(m)

	// ok: test if options are getting passed through
	m, err = router.New("mock", "options")
	test.NoError(err)
	test.NotNil(m)
	test.Equal("options", DummyTestRouter.options.(string))
}

// TestManager_AddPublicDir_AddPublicFile testing different combination and errors.
func TestManager_AddPublicDir_AddPublicFile(t *testing.T) {
	test := assert.New(t)

	r, err := router.New("mock", nil)
	test.NoError(err)

	// create fav.ico
	_, err = os.Create("index.html")
	test.NoError(err)

	// getting actual dir
	dir, err := os.Getwd()
	test.NoError(err)

	// table driven:
	var tests = []struct {
		test     string
		url      string
		source   string
		dir      bool
		length   int
		key      string
		error    bool
		errorMsg string
	}{
		//paths
		{test: "err: source does not exist", dir: true, url: "/assets", source: dir + "/assets", error: true, errorMsg: fmt.Sprintf(router.ErrPathDoesNotExist.Error(), dir+"/assets")},
		{test: "err: source exists but on url root level", dir: true, url: "/", source: "/", error: true, errorMsg: router.ErrRootLevel.Error()},
		{test: "err: url no prefix /", dir: true, url: "test", source: "/", error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: url empty", dir: true, url: "", source: "/", error: true, errorMsg: router.ErrUrl.Error()},
		{test: "ok: source exists on source root level", dir: true, url: "/assets", source: "/", length: 1, key: "/assets"},
		{test: "ok: source exists", dir: true, url: "/assets/", source: dir + "/httprouter/", length: 1, key: "/assets"},
		{test: "ok: source exists trailing slash", dir: true, url: "/assets/", source: dir + "/httprouter/", length: 1, key: "/assets"},
		//files
		{test: "ok: file exists", url: "/exist", source: dir + "/index.html", length: 2, key: "/exist"},
		{test: "ok: url is already defined", url: "/exist", source: dir + "/index.html", length: 2, key: "/exist"}, // its getting overwritten
		{test: "ok: url trailing slashes", url: "/trim/slash/", source: dir + "/index.html", length: 3, key: "/trim/slash"},
		{test: "ok: url / root level", url: "/", source: dir + "/index.html", length: 4, key: "/"},
		{test: "err: url empty", url: "", source: dir + "/index.html", error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: url no prefix", url: "test", source: dir + "/index.html", error: true, errorMsg: router.ErrUrl.Error()},
		{test: "ok: url empty root level", url: "//", source: dir + "/index.html", length: 4, key: "/"},
		{test: "ok: url empty root level", url: "///something///", source: dir + "/index.html", length: 5, key: "///something//"},
		{test: "err: file does not exist", url: "/404", source: dir + "/404.html", error: true, errorMsg: fmt.Sprintf(router.ErrFileDoesNotExist.Error(), dir+"/404.html")},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			var err error
			if tt.dir {
				err = r.AddPublicDir(tt.url, tt.source)
			} else {
				err = r.AddPublicFile(tt.url, tt.source)
			}

			if tt.error == true {
				if test.Error(err) {
					test.Equal(tt.errorMsg, err.Error())
				}
			} else {
				if test.NoError(err) {
					test.Equal(tt.length, len(DummyTestRouter.static))
					// trailing slashes are removed, except if its the root level
					if tt.source != "/" {
						tt.source = strings.TrimSuffix(tt.source, "/")
					}
					test.Equal(tt.source, DummyTestRouter.static[tt.key])
				}
			}
		})
	}

	// delete created fav.ico
	os.Remove("index.html")
}

// TestManager_SetFavicon testing if the path /favicon.ico is getting set correctly.
func TestManager_SetFavicon(t *testing.T) {

	r, err := router.New("mock", nil)
	assert.NoError(t, err)

	// create fav.ico
	_, err = os.Create("fav.ico")
	assert.NoError(t, err)

	// getting actual dir
	dir, err := os.Getwd()
	assert.NoError(t, err)

	// ok: add existing fav icon
	err = r.SetFavicon(dir + "/fav.ico")
	if assert.NoError(t, err) {
		assert.Equal(t, dir+"/fav.ico", DummyTestRouter.static["/favicon.ico"])
	}

	// delete created fav.ico
	os.Remove("fav.ico")
}

// TestManager_Handler testing if the router backend handler will get returned
func TestManager_Handler(t *testing.T) {
	r, err := router.New("mock", nil)
	assert.NoError(t, err)
	assert.Equal(t, r.Handler(), DummyTestRouter.Handler())
}

// TestManager_AllowHTTPMethod checks the trigger and not existing HTTP Methods.
func TestManager_AllowHTTPMethod(t *testing.T) {
	r, err := router.New("mock", nil)
	assert.NoError(t, err)

	err = r.AllowHTTPMethod(router.POST, false)
	assert.NoError(t, err)

	err = r.AllowHTTPMethod("foo", true)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf(router.ErrMethodNotAllowed.Error(), "FOO"), err.Error())
}

// TestManager_SetCache testing if the cache getting passed to the controller.
func TestManager_SetCache(t *testing.T) {
	r, err := router.New("mock", nil)
	assert.NoError(t, err)

	c, err := cache.New(cache.MEMORY, nil)
	assert.NoError(t, err)

	r.SetCache(c)
	err = r.AddPublicRoute("/", &mockController{}, router.RouteConfig{"*:Login", nil})
	assert.NoError(t, err)

	assert.Equal(t, c, DummyTestRouter.routes[0].controller.Cache())
}

// TestManager_NotFound is testing if the NotFound Handler is posted to the router backend
func TestManager_NotFound(t *testing.T) {
	r, err := router.New("mock", nil)
	assert.NoError(t, err)

	handler := js.New()
	r.NotFound(handler)

	assert.Equal(t, handler, DummyTestRouter.notFound)
}

// TestManager_AddPublicRoute_AddSecureRoute testing if the route and middleware gets added correctly.
func TestManager_AddPublicRoute_AddSecureRoute(t *testing.T) {
	test := assert.New(t)

	// get router and disable PUT global
	r, err := router.New("mock", nil)
	test.NoError(err)
	err = r.AllowHTTPMethod(router.PUT, false)
	test.NoError(err)

	// middleware
	mw := mockMiddleware{}
	mwc := &middleware.Chain{}
	mwc = mwc.Add(mw.Logger)

	// table driven:
	var tests = []struct {
		test          string
		public        bool
		pattern       string
		controller    controller.Interface
		controllerMap map[string]string
		config        router.RouteConfig
		error         bool
		errorMsg      string
	}{
		//public routes
		{test: "err: url no prefix /", public: true, pattern: "test", controller: &mockController{}, config: router.RouteConfig{"Get:Post:Login", nil}, error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: url empty", public: true, pattern: "", controller: &mockController{}, config: router.RouteConfig{"Get:Post:Login", nil}, error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: Route config has a wrong syntax", public: true, pattern: "/test", controller: &mockController{}, config: router.RouteConfig{"Get:Post:Login", nil}, error: true, errorMsg: fmt.Sprintf(router.ErrConfigPattern.Error(), "/test")},
		{test: "err: Route config empty", public: true, pattern: "/test", controller: &mockController{}, config: router.RouteConfig{}, error: true, errorMsg: fmt.Sprintf(router.ErrConfigPattern.Error(), "/test")},
		{test: "err: Controller method Api does not exist", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"*:Api", nil}, error: true, errorMsg: fmt.Sprintf(controller.ErrMethodUnknown.Error(), "Api", "router_test.mockController")},
		{test: "err: HTTP Method GET2 does not exist", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"GET2:Login", nil}, error: true, errorMsg: fmt.Sprintf(router.ErrMethodNotAllowed.Error(), "GET2")},
		{test: "err: PUT is global disabled", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"PUT:Login", nil}, error: true, errorMsg: fmt.Sprintf(router.ErrMethodNotAllowed.Error(), "PUT")},
		{test: "err: multiple config,PUT is global disabled", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"POST:Login;PUT:Login", nil}, error: true, errorMsg: fmt.Sprintf(router.ErrMethodNotAllowed.Error(), "PUT")},
		{test: "ok: Wildcard", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"*:Login", nil}, controllerMap: map[string]string{"DELETE": "Login", "GET": "Login", "HEAD": "Login", "OPTIONS": "Login", "PATCH": "Login", "POST": "Login"}},
		{test: "ok: Wildcard and post-specific config", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"*:Login;POST:Logout", nil}, controllerMap: map[string]string{"DELETE": "Login", "GET": "Login", "HEAD": "Login", "OPTIONS": "Login", "PATCH": "Login", "POST": "Logout"}},
		{test: "ok: Wildcard and pre-specific config", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"POST:Logout;*:Login", nil}, controllerMap: map[string]string{"DELETE": "Login", "GET": "Login", "HEAD": "Login", "OPTIONS": "Login", "PATCH": "Login", "POST": "Logout"}},
		{test: "ok: specific route(uppercase/lowercase) is added + multiple methods", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"POST:Login;Get:Logout", nil}, controllerMap: map[string]string{"POST": "Login", "GET": "Logout"}},
		{test: "ok: specific route is added, Multiple methods", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post:Login", nil}, controllerMap: map[string]string{"POST": "Login"}},
		{test: "ok: middleware added", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post:Login", mwc}, controllerMap: map[string]string{"POST": "Login"}},
		{test: "ok: methods with spaces in between", public: true, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post, get , option:Login", mwc}, controllerMap: map[string]string{"POST": "Login"}},

		// secure middleware (err: no secure middleware defined must be at the beginning)
		{test: "err: no secure middleware defined", public: false, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post:Login", nil}, error: true, errorMsg: router.ErrNoSecureMiddleware.Error()},
		{test: "err: url no prefix /", public: false, pattern: "test", controller: &mockController{}, config: router.RouteConfig{"post:Login", nil}, error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: url empty", public: false, pattern: "", controller: &mockController{}, config: router.RouteConfig{"post:Login", nil}, error: true, errorMsg: router.ErrUrl.Error()},
		{test: "err: config error", public: false, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"pot:Login", nil}, error: true, errorMsg: fmt.Sprintf(router.ErrConfigPattern.Error(), "/")},
		{test: "ok: secure middleware added", public: false, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post:Login", nil}, controllerMap: map[string]string{"POST": "Login"}},
		{test: "ok: additional user middleware", public: false, pattern: "/", controller: &mockController{}, config: router.RouteConfig{"post:Login", mwc}, controllerMap: map[string]string{"POST": "Login"}},
	}
	i := 0
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			// test public routes
			if tt.public {
				err := r.AddPublicRoute(tt.pattern, tt.controller, tt.config)
				if tt.error {
					test.Error(err)
					test.Equal(tt.errorMsg, err.Error())
				} else {
					if test.NoError(err) {
						test.Equal(i+1, len(DummyTestRouter.routes))                    // test if our route got added
						test.Equal(tt.pattern, DummyTestRouter.routes[i].pattern)       // test our pattern
						test.Equal(tt.controller, DummyTestRouter.routes[i].controller) // test the controller ptr
						test.Equal(tt.controllerMap, DummyTestRouter.routes[i].controller.MappingBy(tt.pattern))
						test.Equal(tt.config.Middleware, DummyTestRouter.routes[i].mws) // no middlewares exist
						if tt.config.Middleware != nil {
							test.Equal(1, len(DummyTestRouter.routes[i].mws.All()))
						}

						i++
					}
				}
			} else {
				// test secured routes
				err := r.AddSecureRoute(tt.pattern, tt.controller, tt.config)
				if tt.error {
					test.Error(err)
					mw := mockMiddleware{}
					firstMw := mw.JWT
					secondMw := mw.JWT
					r.SetSecureMiddleware(middleware.New(firstMw, secondMw))
				} else {
					if test.NoError(err) {
						test.Equal(i+1, len(DummyTestRouter.routes))                    // test if our route got added
						test.Equal(tt.pattern, DummyTestRouter.routes[i].pattern)       // test our pattern
						test.Equal(tt.controller, DummyTestRouter.routes[i].controller) // test the controller ptr
						test.Equal(tt.controllerMap, DummyTestRouter.routes[i].controller.MappingBy(tt.pattern))
						if tt.config.Middleware != nil {
							count := 2
							if tt.test == "ok: additional user middleware" {
								count = 3
							}
							test.Equal(count, len(DummyTestRouter.routes[i].mws.All()))
						}
						i++
					}
				}
			}
		})
	}

}

// This example demonstrate the basics of the router.Interface.
// For more details check the documentation.
func Example() {
	//import "github.com/patrickascher/gofw/router/httprouter"

	// Creating a new router.
	r, err := router.New(router.HTTPROUTER, httprouter.Options{CatchAllKeyValuePair: true})
	if err != nil {
		//..
	}

	// Disable global the HTTP PATCH method.
	err = r.AllowHTTPMethod(router.PATCH, false)
	if err != nil {
		//..
	}

	// Set a cache provider - which will be added to the controller(s).
	// Simplified it for the example.
	r.SetCache(nil)

	// adding a public file
	err = r.AddPublicFile("/help", "help.pdf")
	if err != nil {
		//..
	}

	// adding a a whole directory
	err = r.AddPublicFile("/asserts", "/webpage/static")
	if err != nil {
		//..
	}

	c := mockController{}
	mw := mockMiddleware{}

	// adding a public route with a custom log middleware
	err = r.AddPublicRoute("/login", &c, router.RouteConfig{"*:Login", middleware.New(mw.Logger)})

	// adding a secure middleware for all secure routes
	r.SetSecureMiddleware(middleware.New(mw.Rbac, mw.JWT))

	// adding a secure route
	err = r.AddSecureRoute("/private", &c, router.RouteConfig{"*:Secret", nil})

	// creating a go server with the router handler
	server := http.Server{}
	server.Handler = r.Handler()
	//...
}
