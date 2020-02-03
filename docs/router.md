<h1>go-router</h1>

!> At the moment only the julienschmidt middleware is supported due a design mistake. To solve this problem, the go-middleware and router.go has to get rewritten. (Chain,Middleware and ChainJR,MiddlewareJR)

go-router is a router manager which is tightly connected to the `ControllerInterface`.
It has special functions to create public or secured routes. Secure routes are getting routed through special middlewares you can set.

You can add a global cache which will then be available in the controller.

It provides a router Interface, that you can create your own router backend.
Out of the box the router of `julienschmidt` is available.

## Install

```go
go get github.com/fullhouse-productions/go-router
```

## Usage 

```go
import "github.com/fullhouse-productions/go-router"
import _ "github.com/fullhouse-productions/go-router/julienschmidt"

// custom not found handler
type vue struct {
}
// If route was not found, redirect everything to the vue frontend.
func (v *vue) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	http.ServeFile(rw, r, "client/dist/index.html")
}

rm,err := Get("julienschmidt");
if err != nil{
	//...
}

// adding a fav.icon
rm.Favicon("/assets/img/fav.ico")

// add a public dir (can be called multiple times)
rm.PublicDir("/assets","/assets")

//Disable HTTP Methods globally
rm.AllowHTTPMethod(CONNECT,false)

// adding secure middleware
rm.SecureMiddleware(middleware.ChainJR(jwt.JWT, rbac.Rbac))

// not found handler
rm.NotFound(&vue{})

// ...
// adding cache
rm.Cache(cache)

// adding routes
ac := controller.AuthController{}
rm.PublicRoute("/login", &ac, router.Config{"POST,OPTIONS:Login", nil})
rm.SecureRoute("/logout", &ac, router.Config{"GET,OPTIONS:Logout", nil})

// adding the router handler to the server
server := http.Server{}
server.Addr = fmt.Sprint(":", c.Server.HTTPSPort)
server.Hanlder = rm.Handler()
defer server.Close()
//...
```

# Manager

## Get

Get will return a new Router-Manger with the requested router backend.
If the router backend does not exist, an error will return.
```go
rm,err := Get("julienschmidt")
```

## Register

Register is used to register a router backend. 
This function should be called in the init function of the router backend to register itself on import.
It returns an error if the router-name or the router-backend itself is empty or the router-backend already exists.

```go
// init registers a json reader
func init() {
	Register("julienschmidt", &Julienschmidt{})
}
```

**Usage**
```go
import "github.com/fullhouse-productions/go-router"
import _ "github.com/fullhouse-productions/go-router/julienschmidt"
```


## Router Interface
To create a own router backend, you have to implement the `Router` interface

```go
type Router interface {
	GetHandler() http.Handler
	NotFound(http.Handler)
	AddRoute(pattern string, c controller.ControllerInterface, m *middleware.ChainJR)
	SetFavicon(string)
	SetStaticFiles(string, string)
}
```

## RouteConfig

For each route an additional config can get set.

**HTTPMethodToFunc**

A HTTP Method must be linked to a Controller function and the Controller function must exist.

Syntax: `POST,OPTIONS:Login` this would mean every HTTP `POST` and `OPTIONS` request will call the `Login` function

Syntax Wildcard:`*:Login` this would link every allowed HTTP Method ([see Allowed HTTP Methods](router?id=allowed-http-methods)) to the `Login` function.

**Middleware**

A middleware for that route can be defined. 
Chained middlewares are possible.

## Allowed HTTP Methods
By default the following HTTP Methods are allowed.
For every HTTP Method a constant exists.
```go
map[string]bool{
	GET:     true,
	POST:    true,
	PUT:     true,
	DELETE:  true,
	PATCH:   true,
	OPTIONS: true,
	HEAD:    true,
	TRACE:   true,
	CONNECT: true,
}
```

You can disallow HTTP Methods global like this

```go
rm.AllowHTTPMethod(TRACE,false)
```

## Cache
A global Cache can be set. The cache has to implement the `cacheInterface.Cache`??? 
If a cache is defined, it will be available in the Controller.

```go
rm.Cache(cache)
```

## NotFound
For all routes which does not exist, a notFound Handler can get defined here.

## PublicRoute
Add a new route to the router backend.
It checks if the configured controller function exists.
Also the Controller gets initialized at this stage and the global cache gets added (if defined).

```go
ac := controller.AuthController{}
err := rm.PublicRoute("/login", &ac, router.Config{"POST,OPTIONS:Login", middleware.ChainJR(middleware.Logger)})
```

!> Url and the Config `HTTPMethodToFunc` must be given otherwise an error will return 


## SecureRoute
Does the same like `PublicRoute` but adds all middlewares to it, which are defined in [see SecureMiddleware](router?id=secureMiddleware)

```go
ac := controller.AuthController{}
err := rm.SecureRoute("/logout", &ac, router.Config{"POST,OPTIONS:Login", nil})
```

!> Url and the Config `HTTPMethodToFunc` must be given otherwise an error will return 


## SecureMiddleware
Adds special middlewares to the secure routes. This can be a JWT,Rbac,Session,... middleware.
Nothing is predefined here, so you have to take care about the security on your own.

```go
// adding secure middleware
rm.SecureMiddleware(middleware.ChainJR(jwt.JWT, rbac.Rbac))
```

!> SecureMiddleware has to be defined before you add the first SecureRoute otherwise it will get ignored!

## PublicDir
PublicDir adds an url to a directory. 
The directory will be added with the absolute path.
All leading and trailing / in the url will get removed

`PublicDir` can get set multiple times.

```go
rm.PublicDir("assets", "/path/to/assets") 
```

## Favicon
Favicon will get added with the given path

```go
rm.PublicDir("/path/to/assets/fav.ico") 
```

## Handler
Handler is returning the mux for the server

# Julienschmidt Backend

The provider is a wrapper for the julienschmifd httprouter.
The router-backend is registering itself on import.
Nothing special to say aside of:

!> Directory listing is disabled

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-router/issues
