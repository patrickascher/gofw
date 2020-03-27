// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package httprouter implements the router.Interface and wraps the julienschmidt.httprouter.
//
// All router params are getting set to the request context with the key router.PARAMS. See Options for more details.
// The matched url pattern is set to the request context with the key router.PATTERN.
package httprouter

import (
	"context"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/logger/console"
	"github.com/patrickascher/gofw/middleware/log"
	"net/http"
	"strconv"
	"strings"

	jsRouter "github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
)

// Error messages
var (
	ErrKeyValuePair = errors.New("router/httprouter: Catch-all key/value pair mismatch")
)

// init registers the router provider
func init() {
	_ = router.Register("httprouter", New)
}

// httpRouterExtended was created because the original httprouter
// it is not possible to add ctxt before the HandlerFunc is called.
type httpRouterExtended struct {
	jsRouter.Router
	additionalData map[string]string
	router         *httpRouter
}

// HandlerFunc - overwrite the original HandlerFunc - this function is the same as original but needed here.
func (h *httpRouterExtended) HandlerFunc(method, path string, handler http.HandlerFunc) {
	h.Handler(method, path, handler)
}

// Handler is adding the Pattern and Params as context.request params.
func (h *httpRouterExtended) Handler(method, path string, handler http.Handler) {
	h.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, p jsRouter.Params) {
			ctx := req.Context()
			ctx = context.WithValue(ctx, router.PATTERN, p.MatchedRoutePath())
			ctx = context.WithValue(ctx, router.PARAMS, h.paramsToMap(p, w))
			req = req.WithContext(ctx)

			handler.ServeHTTP(w, req)
		},
	)
}

// paramsToMap are mapping all router params to the the request context.
func (h *httpRouterExtended) paramsToMap(params jsRouter.Params, w http.ResponseWriter) map[string][]string {
	rv := make(map[string][]string)

	// check if its a catch-all route
	route := params.MatchedRoutePath()
	catchAllRoute := false
	if strings.Contains(route, "*") {
		catchAllRoute = true
	}

	for _, p := range params {
		if p.Key == jsRouter.MatchedRoutePathParam {
			continue
		}

		if p.Key == "filepath" {
			rv[p.Key] = []string{p.Value}
			continue
		}

		if catchAllRoute {
			urlParam := strings.Split(strings.Trim(p.Value, "/"), "/")
			for i := 0; i < len(urlParam); i++ {
				if h.router.options.CatchAllKeyValuePair {

					if i+1 >= len(urlParam) {
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(ErrKeyValuePair.Error()))
						return nil
					}

					rv[urlParam[i]] = []string{urlParam[i+1]}
					i++
					continue
				}
				rv[strconv.Itoa(i)] = []string{urlParam[i]}
			}
			continue
		}
		rv[p.Key] = []string{p.Value}
	}
	return rv
}

// newHttpRouterExtended creates the new extended httprouter.
// TODO create a OPTION function to add all httprouter options here.
func newHttpRouterExtended(ro *httpRouter) *httpRouterExtended {
	r := &httpRouterExtended{}
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true
	r.HandleMethodNotAllowed = true
	r.HandleOPTIONS = true
	r.SaveMatchedRoutePath = true
	r.router = ro
	return r
}

// New configured instance.
func New(options interface{}) router.Interface {
	r := &httpRouter{}

	if options != nil {
		r.options = options.(Options)
	}
	r.file = make(map[string]string)
	r.dir = make(map[string]string)
	return r
}

// httpRouter router provider
type httpRouter struct {
	routes   []route
	dir      map[string]string
	file     map[string]string
	notFound http.Handler
	options  Options
}

type route struct {
	pattern    string
	public     bool
	controller controller.Interface
	mws        *middleware.Chain
}

func (r *route) Pattern() string {
	return r.pattern
}

func (r *route) Public() bool {
	return r.public
}
func (r *route) Controller() controller.Interface {
	return r.controller
}
func (r *route) MW() *middleware.Chain {
	return r.mws
}

// Options for the router provider
type Options struct {
	// CatchAllKeyValuePair will convert /user/*user routes param to key/value pairs
	// True: route like /user/mode/view/id/12 will be converted to params ["mode":"view" "id":"12"]
	// False: route like /user/mode/view/id/12 will be converted to params ["0":"mode" "1":"view" "2":"id" "3":"12"]
	CatchAllKeyValuePair bool
}

// Routes returns all defined routes.
func (hr *httpRouter) Routes() []router.Route {
	var rv []router.Route
	for _, r := range hr.routes {
		r := r
		rv = append(rv, &r)
	}
	return rv
}

// AddRoute to the provider
func (hr *httpRouter) AddRoute(p string, public bool, c controller.Interface, m *middleware.Chain) {
	r := route{pattern: p, public: public, controller: c, mws: m}
	hr.routes = append(hr.routes, r)
}

// AddPublicDir to the provider. Directory listing is disabled.
func (hr *httpRouter) AddPublicDir(url string, source string) {
	hr.dir[url] = source
}

// AddPublicFile to the provider.
func (hr *httpRouter) AddPublicFile(url string, source string) {
	hr.file[url] = source
	fmt.Println("+++++++", hr.file)
}

// Handler returns the mux handler for the server.
// All defined files, directories and controller routes will be added.
// Custom NotFound handler will get set - if defined.
// Directory listing is disabled.
func (hr *httpRouter) Handler() http.Handler {

	c, _ := console.New(console.Options{Color: true})
	_ = logger.Register("console", logger.Config{Writer: c})
	cLogger, _ := logger.Get("console")
	l := log.New(cLogger)

	fmt.Print("Loading Routes...")
	//add files in a directory
	ro := newHttpRouterExtended(hr)

	mw := middleware.Chain{}

	//adding files
	for path, file := range hr.file {
		ro.HandlerFunc("GET", path, mw.Add(l.MW).Handle(
			func(w http.ResponseWriter, req *http.Request) {
				http.ServeFile(w, req, hr.file[req.Context().Value(router.PATTERN).(string)])
			}))
		fmt.Printf("\n\x1b[32m %#v [GET]%v \x1b[49m\x1b[39m ", path, file)
	}

	// adding directories
	for k, path := range hr.dir {
		fileServer := http.FileServer(http.Dir(path))
		pattern := k + "/*filepath"
		ro.HandlerFunc("GET", pattern, mw.Add(l.MW).Handle(
			func(w http.ResponseWriter, req *http.Request) {
				//disable directory listing
				if strings.HasSuffix(req.URL.Path, "/") {
					http.NotFound(w, req)
					return
				}
				if val, ok := req.Context().Value(router.PARAMS).(map[string][]string)["filepath"]; ok {
					req.URL.Path = val[0]
					fileServer.ServeHTTP(w, req)
					return
				}
				http.NotFound(w, req)
				return

			}))
		fmt.Printf("\n\x1b[32m %#v [GET]%v \x1b[49m\x1b[39m ", pattern, http.Dir(path))
	}

	//register all controller routes
	for _, r := range hr.routes {
		fmt.Printf("\n\x1b[32m %#v :name \x1b[49m\x1b[39m ", r.pattern)
		for method, fn := range r.controller.MappingBy(r.pattern) {
			if r.mws != nil {
				ro.HandlerFunc(strings.ToUpper(method), r.pattern, r.mws.Handle(r.controller.ServeHTTP)) //TODO ????? error no url pattern
			} else {
				ro.HandlerFunc(strings.ToUpper(method), r.pattern, r.controller.ServeHTTP)
			}
			fmt.Printf("\x1b[32m [%v]%v name \x1b[49m\x1b[39m ", method, fn)
		}
	}

	//Not Found Handler
	if hr.notFound != nil {
		ro.NotFound = hr.notFound
	}

	return ro
}

//NotFound is a function to add a custom not found handler if a route does not math
func (hr *httpRouter) NotFound(h http.Handler) {
	hr.notFound = h
}
