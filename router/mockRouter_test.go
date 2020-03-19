// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package router_test

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
)

type mockController struct {
	controller.Controller
}

func (tc *mockController) Login() {

}

func (tc *mockController) Logout() {

}

type mockMiddleware struct {
}

func (tm *mockMiddleware) JWT(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("called Jwt")
	}
}

func (tm *mockMiddleware) Rbac(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("called Rbac")
	}
}

func (tm *mockMiddleware) Logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("called Logger")
	}
}

// Test router backend
type mockRouter struct {
	routes   []mockRoute
	static   map[string]string
	notFound http.Handler
	options  interface{}
}

type mockRoute struct {
	pattern    string
	controller controller.Interface
	mws        *middleware.Chain
}

var DummyTestRouter *mockRouter

func newMock(opt interface{}) router.Interface {
	DummyTestRouter = &mockRouter{}
	DummyTestRouter.options = opt
	return DummyTestRouter
}

func (t *mockRouter) NotFound(h http.Handler) {
	t.notFound = h
}

func (t *mockRouter) AddRoute(p string, c controller.Interface, m *middleware.Chain) {
	r := mockRoute{pattern: p, controller: c, mws: m}
	t.routes = append(t.routes, r)
}

func (t *mockRouter) AddPublicDir(url string, source string) {
	if t.static == nil {
		t.static = make(map[string]string)
	}

	t.static[url] = source
}

func (t *mockRouter) AddPublicFile(url string, source string) {
	if t.static == nil {
		t.static = make(map[string]string)
	}
	t.static[url] = source
}

func (t *mockRouter) Handler() http.Handler {
	ro := httprouter.New()
	return ro
}
