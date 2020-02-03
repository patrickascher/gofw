// Package router is a router manager which is tightly connected to the ControllerInterface.
//
// It has special functions to create public or secured routes.
// Secure routes are getting routed through special middlewares you can set.
// You can add a global cache which will then be available in the controller.
//
// It provides a router Interface, that you can create your own router backend.
// Out of the box the router of julienschmidt is available.
// See https://github.com/patrickascher/go-router for more information and examples.
package router

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"fmt"
	cachepackage "github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
	"os"
	"path/filepath"
)

// declarations
var (
	routerStore          = make(map[string]Router)
	wildcardTag          = "*"
	separatorTag         = ";"
	separatorKeyValue    = ":"
	separatorHTTPMethods = ","
)

// HTTP Method constants.
const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
	HEAD    = "HEAD"
	TRACE   = "TRACE"
	CONNECT = "CONNECT"
)

// Errors messages are defined here.
var (
	ErrUnknownRouter          = errors.New("router: unknown router backend %#v (forgotten import?)")
	ErrNoRouter               = errors.New("router: empty router is given")
	ErrPathOrFileDoesNotExist = errors.New("router: path or file does not exist")
	ErrRouterAlreadyExists    = errors.New("router: router already exists")
	ErrMethodFormat           = errors.New("router: controller method mapping is invalid")
	ErrMethodNotAllowed       = errors.New("router: HTTP method is not allowed")
	ErrNoPatternOrHTTPMap     = errors.New("router: no pattern or http map was added")
	ErrHTTPMethodNotExist     = errors.New("router: HTTP method does not exist or is global not allowed")
)

// Register is used to register a router backend.
// This function should be called in the init function of the router backend to register itself on import.
// It returns an error if the router-name or the router-backend itself is empty or the router-backend already exists.
func Register(routerName string, router Router) error {
	if router == nil || routerName == "" {
		return ErrNoRouter
	}
	if _, dup := routerStore[routerName]; dup {
		return ErrRouterAlreadyExists
	}
	routerStore[routerName] = router
	return nil
}

// Get will return a new Router-Manger with the requested router backend.
// If the router backend does not exist, an error will return.
func Get(routerName string) (*Manager, error) {
	r, ok := routerStore[routerName]
	if !ok {
		return &Manager{}, fmt.Errorf(ErrUnknownRouter.Error(), routerName)
	}
	// allowedHTTPMethod defines the global allowed HTTP Methods.
	allowedHTTPMethod := map[string]bool{
		GET:     true,
		POST:    true,
		PUT:     true,
		DELETE:  true,
		PATCH:   true,
		OPTIONS: true,
		HEAD:    true,
		TRACE:   false, //vulnerable to XST https://www.owasp.org/index.php/Cross_Site_Tracing
		CONNECT: false,
	}
	return &Manager{router: r, allowedHTTPMethod: allowedHTTPMethod}, nil
}

//Router interface
type Router interface {
	GetHandler() http.Handler
	NotFound(http.Handler)
	AddRoute(pattern string, c controller.Interface, m *middleware.ChainJR)
	SetFavicon(string)
	SetStaticFiles(string, string)
}

// RouteConfig defines how the route is connected to the controller and if there
// are additional middlewares
//
// HTTPMethodToFunc
// Syntax: POST,OPTIONS:Login this would mean every HTTP POST and OPTIONS request will call the Login function
// Syntax Wildcard:*:Login this would link every allowed HTTP Method (see Allowed HTTP Methods) to the Login function.
//
// Middleware
// A middleware for that route can be defined. Chained middlewares are possible.
type RouteConfig struct {
	HTTPMethodToFunc string
	Middleware       *middleware.ChainJR
}

// parse the given HTTPMethodToFunc String
// Format: "POST,OPTIONS:Login"
// Format: "*:Login"
func (c *RouteConfig) parse(pattern string, m *Manager) (map[string]map[string]string, error) {
	methods := make(map[string]string)

	if len(c.HTTPMethodToFunc) == 0 || pattern == "" {
		return nil, ErrNoPatternOrHTTPMap
	}

	//Splitting http methods - struct methods
	values := strings.Split(c.HTTPMethodToFunc, separatorTag)
valueLoop:
	for _, value := range values {
		//Checking the correct string format
		keyValue := strings.Split(value, separatorKeyValue)
		if len(keyValue) != 2 {
			return map[string]map[string]string{}, ErrMethodFormat
		}

		//Splitting single http methods or handling the wildcard
		httpMethods := strings.Split(keyValue[0], separatorHTTPMethods)
		for _, httpMethod := range httpMethods {
			// wildcard - add all allowed HTTP methods to it
			if httpMethod == wildcardTag {
				for method, allowed := range m.allowedHTTPMethod {
					if allowed == true {
						methods[method] = keyValue[1]
					}
				}
				continue valueLoop
			}
			// manual added HTTP methods
			if val, ok := m.allowedHTTPMethod[strings.ToUpper(httpMethod)]; ok && val == true {
				methods[strings.ToUpper(httpMethod)] = keyValue[1]
			} else {
				return nil, ErrHTTPMethodNotExist
			}
		}
	}

	//add pattern to the return value
	rv := make(map[string]map[string]string)
	rv[pattern] = methods
	return rv, nil
}

type Manager struct {
	router            Router
	cache             cachepackage.Cache
	secureMiddleware  *middleware.ChainJR
	allowedHTTPMethod map[string]bool
	routes            []Route
}

type Route struct {
	Url        string
	Controller string
	Methods    map[string]string
	Public     bool
}

func newRouteFromMapping(c controller.Interface, httpMapping map[string]map[string]string, public bool) Route {
	r := Route{}

	for path, methods := range httpMapping {
		r.Url = path
		r.Public = public
		r.Methods = methods //HTTPMethod:ControllerAction
		r.Controller = strings.Split(strings.Replace(reflect.Indirect(reflect.ValueOf(c)).Type().String(), "Controller", "", -1), ".")[1]
	}

	return r
}

// Routes returns all defined routes
func (m *Manager) Routes() []Route {
	return m.routes
}

// Cache is adding a global Cache which will be available in the Controllers
func (m *Manager) Cache(cache cachepackage.Cache) {
	m.cache = cache
}

//NotFound Handler is used for all not existing routes
func (m *Manager) NotFound(handler http.Handler) {
	m.router.NotFound(handler)
}

// AllowHTTPMethod can allow or deny some HTTP Methods
// By default every HTTPMethod is allowed
// Error will return if the HTTP Method does not exist
func (m *Manager) AllowHTTPMethod(httpMethod string, allow bool) error {
	if _, ok := m.allowedHTTPMethod[httpMethod]; ok {
		m.allowedHTTPMethod[httpMethod] = allow
		return nil
	}
	return ErrMethodNotAllowed
}

// AllowedHTTPMethods returns all allowed HTTP methods
func (m *Manager) AllowedHTTPMethods() []string {
	var httpMethods []string
	for method, allowed := range m.allowedHTTPMethod {
		if allowed {
			httpMethods = append(httpMethods, method)
		}
	}
	return httpMethods
}

func (m *Manager) addRoute(pattern string, c controller.Interface, conf RouteConfig, public bool) (controller.Interface, error) {
	//initialize the controller with the url/httpMethod/structMethod mapping
	httpMapping, err := conf.parse(pattern, m)

	route := newRouteFromMapping(c, httpMapping, public)
	m.routes = append(m.routes, route)

	if err != nil {
		return nil, err
	}
	err = c.Initialize(c, httpMapping)
	c.SetCache(m.cache)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// PublicRoute add a new route to the router backend.
// It´s also checking if the given controller method from the config is existing and sets the controller mapping
func (m *Manager) PublicRoute(pattern string, c controller.Interface, conf RouteConfig) error {
	c, err := m.addRoute(pattern, c, conf, true)
	if err != nil {
		return err
	}

	//add to the router backend
	m.router.AddRoute(pattern, c, conf.Middleware)
	return nil
}

// SecureMiddleware adds special middlewares to the secure routes.
// This can be a JWT,Rbac,Session,... middleware.
// Nothing is predefined here, so you have to take care about the security on your own.
func (m *Manager) SecureMiddleware(jr *middleware.ChainJR) {
	m.secureMiddleware = jr
}

// SecureRoute does the same like PublicRoute but adds all middlewares to it, which are defined in SecureMiddleware
func (m *Manager) SecureRoute(pattern string, c controller.Interface, conf RouteConfig) error {

	c, err := m.addRoute(pattern, c, conf, false)
	if err != nil {
		return err
	}

	// Create a new middleware with the Auth handler at first position
	var mwsWithJwt *middleware.ChainJR
	mwsWithJwt = m.secureMiddleware
	// all user added middlewares to it

	if conf.Middleware != nil {
		mwsWithJwt = mwsWithJwt.Add(conf.Middleware.GetAll()...)
	}

	m.router.AddRoute(pattern, c, mwsWithJwt)
	return nil
}

// PublicDir adds an url to a directory.
// The directory will be added with the absolute path.
// All leading and trailing / in the url will get removed
func (m *Manager) PublicDir(url string, source string) error {
	url = strings.Trim(url, "/")

	dir, err := filepath.Abs(source)
	if _, errDir := os.Stat(dir); err != nil || os.IsNotExist(errDir) {
		return ErrPathOrFileDoesNotExist
	}

	m.router.SetStaticFiles(url, dir)
	return nil
}

// Favicon will get added with the given path
func (m *Manager) Favicon(icon string) error {

	file, err := filepath.Abs(icon)
	if _, errFile := os.Stat(file); err != nil || os.IsNotExist(errFile) {
		return ErrPathOrFileDoesNotExist
	}

	m.router.SetFavicon(file)
	return nil
}

//Handler is returning the mux for the server
func (m *Manager) Handler() http.Handler {
	return m.router.GetHandler()
}
