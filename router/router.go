// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package router provides a router manager for any type that
// implements the router.Interface. Is tightly connected to the middleware and controller package.
//
// It offers an easy way to add public and secure routes. Secure route are getting routed through the defined middleware(s).
// Specific HTTP Methods can be global disabled.
//
// Files or directories can be added. Files are allowed on url root level, directories not. If you need more than the index and fav.ico
// on root level, a notFound handler could be used as workaround.
//
// A cache can be added which will be passed to the controller (todo: better solution?).
//
// Route params and the matched route are added as context (router.PARAMS and router.PATTERN) to the request by the available providers.
// For more details check the provider documentation. If you create your own provider, its a proposal to add this functionality as well.
package router

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/controller"
	"github.com/patrickascher/gofw/middleware"
)

// Allowed HTTP Method constants.
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

// Request context keys.
const (
	PARAMS  = "params"
	PATTERN = "pattern"
)

// pre-defined providers
const (
	HTTPROUTER = "httprouter"
)

// Errors messages are defined here.
var (
	// errors providers
	ErrUnknownProvider       = errors.New("router: unknown router-provider %q")
	ErrNoProvider            = errors.New("router: empty router-name or router-provider is nil")
	ErrProviderAlreadyExists = errors.New("router: router-provider %#v is already registered")
	// errors paths
	ErrUrl              = errors.New("router: url must begin with '/' in %v")
	ErrPathDoesNotExist = errors.New("router: path %#v does not exist")
	ErrFileDoesNotExist = errors.New("router: file %#v does not exist")
	ErrRootLevel        = errors.New("router: a public dir is not allowed on root level")
	// errors config
	ErrConfigPattern      = errors.New("router: config pattern is invalid or empty for %v")
	ErrNoSecureMiddleware = errors.New("router: no secure middleware was added")
	ErrMethodNotAllowed   = errors.New("router: HTTP method %v is not allowed")
)

// declarations
var (
	registry = make(map[string]provider)

	wildcardTag          = "*"
	separatorTag         = ";"
	separatorKeyValue    = ":"
	separatorHTTPMethods = ","
)

//Router interface
type Interface interface {
	// Handler returns the mux for http/server
	Handler() http.Handler
	// custom NotFound handler can be added
	NotFound(http.Handler)
	// AddRoute to the router.
	// pattern is already checked to start with a slash.
	AddRoute(pattern string, public bool, c controller.Interface, m *middleware.Chain)
	// AddPublicDir to the router
	// Dir is not allowed on url root level.
	AddPublicDir(url string, path string)
	// AddPublicFile to the router
	// Files are allowed on url root level.
	AddPublicFile(url string, path string)
	// Routes return all defined routes.
	Routes() []Route
}

// provider is a function which returns the router interface.
// Like this the router provider is getting initialized only when its called.
// As argument an specific router provider option can be passed.
type provider func(interface{}) Interface

// Register the router provider. This should be called in the init() of the providers.
// If the router provider/name is empty or is already registered, an error will return.
func Register(provider string, fn provider) error {
	if fn == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := registry[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists.Error(), provider)
	}
	registry[provider] = fn
	return nil
}

// New creates the requested router provider and returns a router manager.
// Global HTTP Methods getting defined.
// If the provider is not registered an error will return.
func New(provider string, options interface{}) (*Manager, error) {
	instanceFn, ok := registry[provider]
	if !ok {
		return nil, fmt.Errorf(ErrUnknownProvider.Error(), provider)
	}

	r := instanceFn(options)
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

type Manager struct {
	router            Interface
	cache             cache.Interface
	secureMiddleware  *middleware.Chain
	allowedHTTPMethod map[string]bool
}

type Route interface {
	Pattern() string
	Public() bool
	Controller() controller.Interface
	MW() *middleware.Chain
}

// Routes return all defined routes.
func (m *Manager) Routes() []Route {
	return m.router.Routes()
}

// SetCache will pass the cache to the controllers.
// TODO check if this should be done in the request.Context or global?
func (m *Manager) SetCache(c cache.Interface) {
	m.cache = c
}

// NotFound will be executed if a given route does not exist.
// If it is not set, it depends on the provider.
// For example the httprouter provider is using http.NotFound as default.
func (m *Manager) NotFound(h http.Handler) {
	m.router.NotFound(h)
}

// AllowHTTPMethod can be used to generally allow or deny HTTP Methods for the whole router.
// Error will return if the HTTP Method does not exist
func (m *Manager) AllowHTTPMethod(httpMethod string, allow bool) error {
	httpMethod = strings.ToUpper(httpMethod)
	if _, ok := m.allowedHTTPMethod[httpMethod]; ok {
		m.allowedHTTPMethod[httpMethod] = allow
		return nil
	}
	return fmt.Errorf(ErrMethodNotAllowed.Error(), httpMethod)
}

// AddPublicRoute to the router provider.
// The pattern must start with a slash.
// An error will return if the pattern is misspelt, the controller method does not exist or the HTTP Method is not allowed.
func (m *Manager) AddPublicRoute(pattern string, c controller.Interface, conf RouteConfig) error {
	// initialize the controller with the mapping.
	c, err := m.controllerMapping(pattern, c, conf)
	if err != nil {
		return err
	}

	//add to the router provider
	m.router.AddRoute(pattern, true, c, conf.Middleware)
	return nil
}

// SetSecureMiddleware creates a global middleware for all secure routes.
// Specific routes are getting chained.
func (m *Manager) SetSecureMiddleware(c *middleware.Chain) {
	m.secureMiddleware = c
}

// AddSecureRoute to the router provider. Before usage, the router secure middleware must be set.
// The pattern must start with a slash.
// An error will return if secure middleware is not set, the pattern is misspelt, the controller method does not exist or the HTTP Method is not allowed.
func (m *Manager) AddSecureRoute(pattern string, c controller.Interface, conf RouteConfig) error {
	// check if router secure middleware is set
	if m.secureMiddleware == nil {
		return ErrNoSecureMiddleware
	}

	// initialize the controller with the mapping.
	c, err := m.controllerMapping(pattern, c, conf)
	if err != nil {
		return err
	}

	// adding custom middleware if defined.
	mw := m.secureMiddleware
	if conf.Middleware != nil {
		mw = mw.Add(conf.Middleware.All()...)
	}

	// adding to the router provider
	m.router.AddRoute(pattern, false, c, mw)
	return nil
}

// addFiles is a helper for AddPublicFile and AddPublicDir.
// It checks the url pattern and the url root level (files are allowed on root level directories not).
// If a file or directory does not exist, an error will return.
func (m *Manager) addFiles(url string, source string, dir bool) error {

	if url == "" || url[0] != '/' {
		return ErrUrl
	}

	url = strings.TrimSuffix(url, "/")
	if !dir && url == "" {
		url = "/"
	}
	if dir && url == "" {
		return ErrRootLevel
	}

	s, err := os.Executable()
	if err != nil {
		return err
	}
	path, err := filepath.Abs(path.Dir(s) + "/" + source)
	if info, errDir := os.Stat(path); err != nil || os.IsNotExist(errDir) || (info != nil && info.IsDir() != dir) {
		if dir {
			return fmt.Errorf(ErrPathDoesNotExist.Error(), source)
		}
		return fmt.Errorf(ErrFileDoesNotExist.Error(), source)
	}

	if dir {
		m.router.AddPublicDir(url, path)
		return nil
	}

	m.router.AddPublicFile(url, path)
	return nil
}

// AddPublicDir to the router provider.
// The directory will be added with the absolute path.
// Error will return if the folder does not exist or its the url root level.
// If for any reasons its necessary to add a directory to the root level, this can be done with an hack of NotFound handler.
// Otherwise only the favicon and index should be on root level. (Note: This is required because the most router providers
// add an all-catch route for directories *filepath. If that would be on root level /*filepath, it could happen that the route
// /user will never be triggered - depending on the provider)
//
// Proposal for the router providers, disable directory listing by default.
func (m *Manager) AddPublicDir(url string, source string) error {
	return m.addFiles(url, source, true)
}

// AddPublicFile to the router provider.
// Url root level is allowed.
// Error will return if the file does not exist.
func (m *Manager) AddPublicFile(url string, source string) error {
	return m.addFiles(url, source, false)
}

// SetFavicon to the router provider.
// Error will return if the file does not exist.
func (m *Manager) SetFavicon(source string) error {
	return m.addFiles("/favicon.ico", source, false)
}

// Handler returns the mux for the http/server
func (m *Manager) Handler() http.Handler {
	return m.router.Handler()
}

// controllerMapping is helper for AddPublicRoute and AddSecureRoute.
// Its initializing the controller with the given mapping and sets the cache.
// Error will return if the patter, config is wrong or the controller function does not exist.
func (m *Manager) controllerMapping(pattern string, c controller.Interface, conf RouteConfig) (controller.Interface, error) {
	if pattern == "" || pattern[0] != '/' {
		return nil, ErrUrl
	}

	//initialize the controller with the url/httpMethod/structMethod mapping
	httpMapping, err := conf.parse(pattern, m)
	if err != nil {
		return nil, err
	}

	err = c.Initialize(c, httpMapping, true)
	c.SetCache(m.cache)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// RouteConfig defines the mapping between HTTP Methods(s) and controller functions.
// Optional custom middleware(s) can be added by route.
//
// Syntax:
//
//		// GET -> controller.List, POST -> controller.Save
//		router.RouteConfig{"router.GET:List;router.POST:Save", nil}
//
//		// All allowed HTTP Methods -> controller.Login
//		router.RouteConfig{"*:Login", nil}
//
//		// All allowed HTTP Methods -> controller.List except POST -> controller.Save
//		router.RouteConfig{"*:List;router.POST:Save", nil}
type RouteConfig struct {
	HTTPMethodToFunc string
	Middleware       *middleware.Chain
}

// parse the given mapping and prepare it for the controller.
// return value syntax map[URL][HTTPMethod]ControllerFunc
func (c *RouteConfig) parse(pattern string, m *Manager) (map[string]map[string]string, error) {
	methods := make(map[string]string)

	if len(c.HTTPMethodToFunc) == 0 {
		return nil, fmt.Errorf(ErrConfigPattern.Error(), pattern)
	}

	//Splitting http methods - struct methods
	values := strings.Split(c.HTTPMethodToFunc, separatorTag)
valueLoop:
	for _, value := range values {
		//Checking the correct string format
		keyValue := strings.Split(value, separatorKeyValue)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf(ErrConfigPattern.Error(), pattern)
		}

		//Splitting single http methods or handling the wildcard
		httpMethods := strings.Split(keyValue[0], separatorHTTPMethods)
		for _, httpMethod := range httpMethods {
			httpMethod = strings.Trim(httpMethod, " ")
			// wildcard - add all allowed HTTP methods to it
			if httpMethod == wildcardTag {
				for method, allowed := range m.allowedHTTPMethod {
					if _, exist := methods[method]; !exist && allowed == true {
						methods[method] = keyValue[1]
					}
				}
				continue valueLoop
			}
			// manual added HTTP methods
			if val, ok := m.allowedHTTPMethod[strings.ToUpper(httpMethod)]; ok && val == true {
				methods[strings.ToUpper(httpMethod)] = keyValue[1]
			} else {
				return nil, fmt.Errorf(ErrMethodNotAllowed.Error(), strings.ToUpper(httpMethod))
			}
		}
	}

	//add pattern to the return value
	rv := make(map[string]map[string]string)
	rv[pattern] = methods

	return rv, nil
}
