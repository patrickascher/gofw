// Package julienschmidt is a wrapper for the julienschmidt htttprouter
package julienschmidt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/router"
)

// init registers the router backend
func init() {
	router.Register("julienschmidt", &Julienschmidt{})
}

// Julienschmidt router backend
type Julienschmidt struct {
	routes   []Route
	favicon  string
	statics  map[string]string
	notFound http.Handler
}

// Route contains a pattern and the pointer to the controller
type Route struct {
	pattern    string
	controller controller.Interface
	mws        *middleware.ChainJR
}

// AddRoute to the Julienschmidt struct, it will be used in the GetHandler() method
func (j *Julienschmidt) AddRoute(p string, c controller.Interface, m *middleware.ChainJR) {
	r := Route{pattern: p, controller: c, mws: m}
	j.routes = append(j.routes, r)
}

// SetFavicon to the Julienschmidt struct, it will be used in the GetHandler() method
func (j *Julienschmidt) SetFavicon(f string) {
	j.favicon = f
}

// SetStaticFiles to the Julienschmidt struct, it will be used in the GetHandler() method
func (j *Julienschmidt) SetStaticFiles(url string, source string) {
	if j.statics == nil {
		j.statics = make(map[string]string)
	}

	j.statics[url] = source
}

// GetHandler returns the mux handler for the server
// All Routes will be added and connected to the ServeHTTP method of the controller
// Favicon route will be added
// Adds all public directories - Directory listing is disabled!
func (j *Julienschmidt) GetHandler() http.Handler {

	fmt.Print("Loading Routes...")

	ro := httprouter.New()

	mw := middleware.ChainJR{}
	//add fav icon
	if j.favicon != "" {
		ro.GET("/favicon.ico", mw.Add(middleware.LoggerJR).Handle(j.faviconHandler))
		fmt.Printf("\n\x1b[32m %#v [GET]%v \x1b[49m\x1b[39m ", "/favicon.ico", j.favicon)
	}

	//add statics files
	for k, v := range j.statics {
		fileServer := http.FileServer(http.Dir(v))

		path := "/" + k + "/*filepath"
		ro.GET(path, mw.Add(middleware.LoggerJR).Handle(func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
			//disable directory listing
			req.URL.Path = ps.ByName("filepath")
			if strings.HasSuffix(req.URL.Path, "/") {
				http.NotFound(w, req)
				return
			}
			fileServer.ServeHTTP(w, req)
		}))
		fmt.Printf("\n\x1b[32m %#v [GET]%v \x1b[49m\x1b[39m ", path, http.Dir(v))
	}

	//Register Controller Routes
	for _, r := range j.routes {
		fmt.Printf("\n\x1b[32m %#v :name \x1b[49m\x1b[39m ", r.pattern)
		for method, fn := range r.controller.HTTPMethodsByPattern(r.pattern) {
			if r.mws != nil {
				ro.Handle(strings.ToUpper(method), r.pattern, r.mws.Handle(r.controller.ServeHTTPJR))
			} else {
				ro.Handle(strings.ToUpper(method), r.pattern, r.controller.ServeHTTPJR)
			}
			fmt.Printf("\x1b[32m [%v]%v name \x1b[49m\x1b[39m ", method, fn)
		}
	}

	//Not Found Handler
	if j.notFound != nil {
		ro.NotFound = j.notFound
	}

	return ro
}

//NotFound is a function to add a custom not found handler if a route does not math
func (j *Julienschmidt) NotFound(h http.Handler) {
	j.notFound = h
}

// faviconHandler is a simple serve file handler with the path to the favicon
func (j *Julienschmidt) faviconHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.ServeFile(rw, r, j.favicon)
}
