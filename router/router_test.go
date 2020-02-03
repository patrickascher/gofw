package router_test

import (
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

// TestRouter_Register testing if the Provider is getting registered
// it also tests if the provider already exist or the register name or provider is missing
func TestGetAndRegister(t *testing.T) {

	// Testing get with an existing Router
	_, err := router.Get("TestRouter")
	assert.NoError(t, err)

	// Testing get with a non existing Router
	_, err = router.Get("notExisting")
	assert.Error(t, err)

	//Testing error - no provider name exist or provider is nil
	err = router.Register("", &TestRouter{})
	assert.Equal(t, router.ErrNoRouter, err)
	assert.Error(t, err)
	err = router.Register("Memory", nil)
	assert.Equal(t, router.ErrNoRouter, err)
	assert.Error(t, err)

	//Testing normal Register - no error should happen
	err = router.Register("Memory", &TestRouter{})
	assert.NoError(t, err)

	//Testing error - provider already exist
	err = router.Register("Memory", &TestRouter{})
	assert.Equal(t, router.ErrRouterAlreadyExists, err)
}

func TestManager_PublicDirAndFavicon(t *testing.T) {

	r, err := router.Get("TestRouter")
	assert.NoError(t, err)

	// create fav.ico
	_, err = os.Create("fav.ico")
	assert.NoError(t, err)

	// getting actual dir
	dir, err := os.Getwd()
	assert.NoError(t, err)

	// add fav icon
	err = r.Favicon(dir + "/fav.ico")
	assert.NoError(t, err)
	assert.Equal(t, dir+"/fav.ico", DummyTestRouter.favicon)
	// Fav icon does not exist
	err = r.Favicon(dir + "/favicon.ico")
	assert.Error(t, err)

	// Dir does not exist
	err = r.PublicDir("/assets", dir+"/assets")
	assert.Error(t, err)
	// Dir exists
	err = r.PublicDir("/assets", "/")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.statics))
	assert.Equal(t, "/", DummyTestRouter.statics["assets"])
	// Direct file
	err = r.PublicDir("/img", dir+"/fav.ico")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(DummyTestRouter.statics))
	assert.Equal(t, dir+"/fav.ico", DummyTestRouter.statics["img"])

	// delete created fav.ico
	os.Remove("fav.ico")
}

// TestManager_Handler testing if the router backend handler will get returned
func TestManager_Handler(t *testing.T) {
	r, err := router.Get("TestRouter")
	assert.NoError(t, err)
	assert.Equal(t, r.Handler(), DummyTestRouter.GetHandler())
}

// TestManager_NotFound is testing if the NotFound Handler is posted to the router backend
func TestManager_NotFound(t *testing.T) {
	r, err := router.Get("TestRouter")
	assert.NoError(t, err)

	handler := httprouter.New()
	r.NotFound(handler)

	assert.Equal(t, handler, DummyTestRouter.notFound)
}

// TestManager_PublicRoute testing if the route and middleware gets added correctly
func TestManager_PublicRoute(t *testing.T) {
	r, err := router.Get("TestRouter")
	assert.NoError(t, err)

	// Error RouteConfig/HTTPMethodToFunc has a wrong syntax
	c := TestController{}
	err = r.PublicRoute("test", &c, router.RouteConfig{HTTPMethodToFunc: "Wrong:Syntax:Login", Middleware: nil})
	assert.Error(t, err)

	// Error Controller Method Api does not exist
	c = TestController{}
	err = r.PublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "*:Api", Middleware: nil})
	assert.Error(t, err)

	// Error Pattern is empty
	c = TestController{}
	err = r.PublicRoute("", &c, router.RouteConfig{HTTPMethodToFunc: "*:Api", Middleware: nil})
	assert.Error(t, err)

	// Error HTTP Method GET2 does not exist
	c = TestController{}
	err = r.PublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "GET2:Api", Middleware: nil})
	assert.Error(t, err)

	//Disallow GET globally
	r.AllowHTTPMethod(router.GET, false)
	// Error HTTP Method GET is not allowed
	c = TestController{}
	err = r.PublicRoute("/", &c, router.RouteConfig{HTTPMethodToFunc: "GET:Api", Middleware: nil})
	assert.Error(t, err)

	// OK - Controller function exist, no middleware added, wildcard
	DummyTestRouter.reset()
	c = TestController{}
	err = r.PublicRoute("test", &c, router.RouteConfig{HTTPMethodToFunc: "*:Login", Middleware: nil})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.routes))                            // test if our route got added
	assert.Equal(t, "test", DummyTestRouter.routes[0].pattern)                 // test our pattern
	assert.Equal(t, &c, DummyTestRouter.routes[0].controller)                  // test the controller ptr
	assert.Equal(t, (*middleware.ChainJR)(nil), DummyTestRouter.routes[0].mws) // no middlewares exist

	// OK - Controller function exist, no middleware added, specific HTTP Method
	DummyTestRouter.reset()
	c = TestController{}
	err = r.PublicRoute("test", &c, router.RouteConfig{HTTPMethodToFunc: "post:Login", Middleware: nil})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.routes))                            // test if our route got added
	assert.Equal(t, "test", DummyTestRouter.routes[0].pattern)                 // test our pattern
	assert.Equal(t, &c, DummyTestRouter.routes[0].controller)                  // test the controller ptr
	assert.Equal(t, (*middleware.ChainJR)(nil), DummyTestRouter.routes[0].mws) // no middlewares exist

	// OK - Controller function exist, no middleware added
	DummyTestRouter.reset()
	c = TestController{}
	err = r.PublicRoute("test", &c, router.RouteConfig{HTTPMethodToFunc: "*:Login", Middleware: nil})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.routes))                            // test if our route got added
	assert.Equal(t, "test", DummyTestRouter.routes[0].pattern)                 // test our pattern
	assert.Equal(t, &c, DummyTestRouter.routes[0].controller)                  // test the controller ptr
	assert.Equal(t, (*middleware.ChainJR)(nil), DummyTestRouter.routes[0].mws) // no middlewares exist

	// OK - Controller function exist, middleware added
	DummyTestRouter.reset()
	our := TestMiddleware{}
	mws := &middleware.ChainJR{}
	mws = mws.Add(our.Logger)

	c = TestController{}
	err = r.PublicRoute("/logout", &c, router.RouteConfig{HTTPMethodToFunc: "*:Logout", Middleware: mws})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.routes))
	assert.Equal(t, "/logout", DummyTestRouter.routes[0].pattern)
	assert.Equal(t, 1, len(DummyTestRouter.routes[0].mws.GetAll()))
	assert.Equal(t, &c, DummyTestRouter.routes[0].controller)
}

// TestManager_SecureRoute is doing the same thing as PublicRoute just some secure routes are added
func TestManager_SecureRoute(t *testing.T) {

	r, err := router.Get("TestRouter")
	assert.NoError(t, err)

	// Error Controller pattern is empty
	c := TestController{}
	err = r.SecureRoute("", &c, router.RouteConfig{HTTPMethodToFunc: "", Middleware: nil})
	assert.Error(t, err)

	// OK - testing if the middlewares are getting added correctly
	our := TestMiddleware{}

	firstMw := our.JWT
	secondMw := our.JWT
	thirdMw := our.Logger
	r.SecureMiddleware(middleware.NewJR(firstMw, secondMw))

	DummyTestRouter.reset()
	c = TestController{}
	mws := &middleware.ChainJR{}
	mws = mws.Add(thirdMw)
	err = r.SecureRoute("/logout", &c, router.RouteConfig{HTTPMethodToFunc: "*:Logout", Middleware: mws})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(DummyTestRouter.routes))
	assert.Equal(t, "/logout", DummyTestRouter.routes[0].pattern)
	assert.Equal(t, 3, len(DummyTestRouter.routes[0].mws.GetAll()))

	// testing the position of the middleware. TODO find a better solution
	for i, mws := range DummyTestRouter.routes[0].mws.GetAll() {
		switch i {
		case 0:
			assert.Equal(t, reflect.ValueOf(firstMw).Pointer(), reflect.ValueOf(mws).Pointer())
		case 1:
			assert.Equal(t, reflect.ValueOf(secondMw).Pointer(), reflect.ValueOf(mws).Pointer())
		case 2:
			assert.Equal(t, reflect.ValueOf(thirdMw).Pointer(), reflect.ValueOf(mws).Pointer())
		}
	}
	assert.Equal(t, &c, DummyTestRouter.routes[0].controller)
}
