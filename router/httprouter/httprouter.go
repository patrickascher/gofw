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

	"github.com/julienschmidt/httprouter"
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

// Options for the router provider
type Options struct {
	// CatchAllKeyValuePair will convert /user/*user routes param to key/value pairs
	// True: route like /user/mode/view/id/12 will be converted to params ["mode":"view" "id":"12"]
	// False: route like /user/mode/view/id/12 will be converted to params ["0":"mode" "1":"view" "2":"id" "3":"12"]
	CatchAllKeyValuePair bool
}

type route struct {
	pattern    string
	controller controller.Interface
	mws        *middleware.Chain
}

// AddRoute to the provider
func (hr *httpRouter) AddRoute(p string, c controller.Interface, m *middleware.Chain) {
	r := route{pattern: p, controller: c, mws: m}
	hr.routes = append(hr.routes, r)
}

// AddPublicDir to the provider. Directory listing is disabled.
func (hr *httpRouter) AddPublicDir(url string, source string) {
	hr.dir[url] = source
}

// AddPublicFile to the provider.
func (hr *httpRouter) AddPublicFile(url string, source string) {
	hr.file[url] = source
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
	ro := httprouter.New()
	ro.SaveMatchedRoutePath = true

	mw := middleware.Chain{}

	//adding files
	for path, file := range hr.file {
		ro.HandlerFunc("GET", path, mw.Add(l.MW).Handle(
			func(w http.ResponseWriter, req *http.Request) {
				http.ServeFile(w, req, file)
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
				req.URL.Path = httprouter.ParamsFromContext(req.Context()).ByName("filepath")
				fileServer.ServeHTTP(w, req)
			}))
		fmt.Printf("\n\x1b[32m %#v [GET]%v \x1b[49m\x1b[39m ", pattern, http.Dir(path))
	}

	//register all controller routes
	for _, r := range hr.routes {
		fmt.Printf("\n\x1b[32m %#v :name \x1b[49m\x1b[39m ", r.pattern)
		for method, fn := range r.controller.MappingBy(r.pattern) {

			h := func(w http.ResponseWriter, req *http.Request) {
				params := httprouter.ParamsFromContext(req.Context())
				ctx := context.WithValue(req.Context(), router.PATTERN, params.MatchedRoutePath())
				ctx = context.WithValue(ctx, router.PARAMS, hr.paramsToMap(params, w))
				r.controller.ServeHTTP(w, req.WithContext(ctx))
			}

			if r.mws != nil {
				ro.HandlerFunc(strings.ToUpper(method), r.pattern, r.mws.Handle(h))
			} else {
				ro.HandlerFunc(strings.ToUpper(method), r.pattern, h)
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

// paramsToMap are mapping all router params to the the request context.
func (hr *httpRouter) paramsToMap(params httprouter.Params, w http.ResponseWriter) map[string]string {
	rv := make(map[string]string)

	// check if its a catch-all route
	route := params.MatchedRoutePath()
	catchAllRoute := false
	if strings.Contains(route, "*") {
		catchAllRoute = true
	}

	for _, p := range params {
		if p.Key == httprouter.MatchedRoutePathParam {
			continue
		}

		if catchAllRoute {
			urlParam := strings.Split(strings.Trim(p.Value, "/"), "/")
			for i := 0; i < len(urlParam); i++ {
				if hr.options.CatchAllKeyValuePair {

					if i+1 >= len(urlParam) {
						w.WriteHeader(http.StatusInternalServerError)
						_, _ = w.Write([]byte(ErrKeyValuePair.Error()))
						return nil
					}

					rv[urlParam[i]] = urlParam[i+1]
					i++
					continue
				}
				rv[strconv.Itoa(i)] = urlParam[i]
			}
			continue
		}
		rv[p.Key] = p.Value
	}
	return rv
}

//NotFound is a function to add a custom not found handler if a route does not math
func (hr *httpRouter) NotFound(h http.Handler) {
	hr.notFound = h
}
