package router_test

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
	"net/http"
)

type TestController struct {
	controller.Controller
}

func (tc *TestController) Login() {

}

func (tc *TestController) Logout() {

}

type TestMiddleware struct {
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) JWT(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Jwt")
	}
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) Rbac(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Rbac")
	}
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) Logger(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Logger")
	}
}

// Test router backend
type TestRouter struct {
	routes   []TestRoute
	favicon  string
	statics  map[string]string
	notFound http.Handler
}

// TestRoute contains a pattern and the pointer to the controller
type TestRoute struct {
	pattern    string
	controller controller.Interface
	mws        *middleware.ChainJR
}

var DummyTestRouter *TestRouter

func init() {
	DummyTestRouter = &TestRouter{}
	router.Register("TestRouter", DummyTestRouter)
}
func (t *TestRouter) reset() {
	new := &TestRouter{}
	t.routes = new.routes
	t.favicon = new.favicon
	t.statics = new.statics
	t.notFound = new.notFound
}

func (t *TestRouter) NotFound(h http.Handler) {
	t.notFound = h
}

//AddRoute to the Julienschmidt struct, it will be used in the GetHandler() method
func (t *TestRouter) AddRoute(p string, c controller.Interface, m *middleware.ChainJR) {
	r := TestRoute{pattern: p, controller: c, mws: m}
	t.routes = append(t.routes, r)
}

//SetFavicon to the Julienschmidt struct, it will be used in the GetHandler() method
func (t *TestRouter) SetFavicon(f string) {
	t.favicon = f
}

//SetStaticFiles to the Julienschmidt struct, it will be used in the GetHandler() method
func (t *TestRouter) SetStaticFiles(url string, source string) {
	if t.statics == nil {
		t.statics = make(map[string]string)
	}

	t.statics[url] = source
}

// GetHandler
func (t *TestRouter) GetHandler() http.Handler {
	ro := httprouter.New()
	return ro
}
