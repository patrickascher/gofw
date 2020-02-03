<h1>go-controller</h1>

go-controller offers a controller,action based handler for the router.
In this package all struct methods are called struct functions, to avoid collision with HTTP Methods.

# Install

```go
go get github.com/fullhouse-productions/go-controller
```

# Controller

## Usage

## Interface

```go
type Interface interface {
	// controller funcs
	Initialize(Interface, map[string]map[string]string) error
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
	ServeHTTPJR(rw http.ResponseWriter, r *http.Request, p httprouter.Params)

	// Methods
	HTTPMethodsByPattern(string) map[string]string //needed by router
	functionByPatternAndHTTPMethod(string, string) (func(), error)
	setSkipFuncChecks(bool)

	// Context
	setContext(ctx *Context)
	Context() *Context

	// Cache
	Cache() cacheInterface.Cache
	SetCache(cacheInterface.Cache)
	HasCache() bool

	// render types
	RenderType() string
	SetRenderType(string)

	// controller helpers
	Set(string, interface{})
	Error(int, string)
	Redirect(status int, url string)

	checkBrowserCancellation() bool
}
```

## Initialize
Initialize the controller struct.
It should get called when a route gets added. 

If it runs for the first time, the pattern and HTTPMethod<->StructFunc mapping gets created. 
If a struct function does not exist, an error will return. 

Also the controller name and render type will get set.


## ServeHTTP & ServeHTTPJR

ServeHTTP will be called with each request. There is a ServeHTTP for the normal handler and a ServeHTTPJR for the julienschmidt router.
It will create a new copy of the controller and initialize it. 
For performance reasons, it will not check if the struct function exists. This was already when the route got added.


Following procedure:
* create a new instance of the controller
* action is called
* --check if Browser is still here
* Response

!> TODO: Add a defer error class to show the error on the frontend.


## HTTPMethodsByPattern
HTTPMethodsByPattern returns a mapping of all existing HTTPMethods to struct method.
This is needed for the router.

## Context
Context returns the controller context

## Cache
The following functions are available `HasCache` (checks if a controller cache is set) `SetCache` (sets a controller cache) and `Cache (get a controller cache)
Cache gets the controller cache



```go

c := Controller{}

if c.HasCache(){
	//...
	c.SetCache(cache)
}
cache := c.Cache()

```

## RenderType & SetRenderType
A Controller can have his own render type. By default is it `JSON`. It can be defined for the whole controller in its `init` function or for each action (function).

```go
// set global render type for all actions to json
c := Controller{}
c.SetRender("json")
// or
func(c *Controller) init(){
	c.SetRender("json")
}

// set render type to html for the login action
func (c *Controller) Login(){
	c.SetRender("html")
}
```

## Set

Set is a helper to set response data. In the background its calling `c.Context().Response.AddDate()`.
The key has to be a string and the value must be of any type (interface) 

```go
c := Controller{}
c.Set("user",User{Name:"Wall-E",Type:"Robot"})
```

## Error

Error is calling `http.Error` or sets an error variable if the render type is `json`. There are two arguments, the first one is the HTTP Code and the second one is the error message.
```go
c := Controller{}
c.Error(401,"No access granted"})
```

## Redirect

Redirect is creating a HTTP redirect. The first argument must be the HTTP Status and the second argument is the redirect address.

```go
// set global render type for all actions to json
c := Controller{}
c.Redirect(301,"https://www.google.com"})
```

## functionByPatternAndHTTPMethod
functionByPatternAndHTTPMethod returns the struct function.
If the struct function does not exist, an error will return.

## setContext
setContext sets the controller context

## checkBrowserCancellation

If a connection is cancelled a `499 Client Closed Request ` will respond.


## isInitialized
checks if the controller is already initialized. For that it checks if the controller.name variable is set.

## setSkipFuncChecks
This method is set before a controller is copied to avoid the reflect and checks of the struct functions.

## getFunc 
getFunc is a helper to reflect the method of the caller controller.
It will return an error, if the struct method does not exist.

## copyController
copyController creates a new instance of the controller itself.

# Context

Context is including the Request and Response.

## Request

?> TODO Tests for Pattern, Token, Julienschmidt Wildcard, Multipart form

### Pattern
returns the router pattern
`/user/:test/*grid`

### Token (JWT)

!> maybe delete it?!?!

### Method
Returns the HTTP Method in uppercase

```go
req.Method() // GET
```

### Is
Checks the given method with the request Method. Both strings are getting set to uppercase.
There are the following helper functions available:

* IsPost
* IsGet
* IsPatch
* IsPut
* IsDelete

```go
req.Is("GET") // boolean
req.IsGet() // boolean
```
### IsSecure
Checks if the request is https

```go
req.IsSecure() // boolean
```

### Raw
Returns the original *http.Request

```go
req.Raw() // *http.Request
```

### UserAgent

#### OS
Name and Version are available of the OperationSystem.

```go
req.UserAgent().OsInfo // Name:Mac OS X Version:10.13.6
```

#### Mobile
boolean

```go
req.UserAgent().Mobile // boolean
```

#### Browser
Name and Version are available of the Browser.

```go
req.UserAgent().BrowserInfo // Name:Chrome Version:69.0.3497.100
```

### IP
It tries to find the IP Address in th X-Forwarded-For Header. If it exists, it takes the first entry.
If it does not exist, it returns the RemoteAddr without Port.


```go
req.IP() // 192.168.0.1
```

?> Create a better solution like here described:https://husobee.github.io/golang/ip-address/2015/12/17/remote-ip-go.html


### Proxy
Return all IPs which are in the X-Forwarded-For header.

```go
req.Proxy() // []string{192.168.0.1}
```

### Scheme
First it checks the `X-Forwarded-Proto` header. If that one is empty the URL.Scheme gets checked.
If that is also empty the request TLS gets checked.

```go
req.Scheme() // http or https
```

### Host
Returns the host name. If no host info is available, localhost will return. Port number will be removed if existing.
If no host info exists, it will return localhost.
```go
req.Host() // example.com
```

### Protocol
Returns the request HTTP Protocol name.

```go
req.Protocol() // HTTP/1.1
```

### URI
URI returns the request url with query string and fragment.
```go
// https://test.com:8043/user?q=dotnet#test
req.URI() // /user?q=dotnet#test
```

### URL
URL returns request url path (without query string, fragment).

```go
// https://test.com:8043/user?q=dotnet#test
req.URL() // /user
```

### FullURL
FullURL returns the schema,host,port,uri

```go
// https://test.com:8043/user?q=dotnet#test
req.FullURL() // https://test.com:8043/user?q=dotnet#test
```

### Site
Site returns base site url as scheme://domain type without the port.

```go
// https://test.com:8043/user?q=dotnet#test
req.Site() // https://test.com
```

### Domain
Domain returns host name.
Alias of Host method.

```go
// https://test.com:8043/user?q=dotnet#test
req.Domain() // test.com
```

### Port
Port returns request client port.
when error or empty, return 80.

```go
// https://test.com:8043/user?q=dotnet#test
req.Port() // 8043
```

### Referer
Returns the Referer Header

### AddJulienSchmidtRouterParams
AddJulienSchmidtRouterParams is used to add the router specific params.
It also sets the router pattern.

### Param
Param returns the requested parameter.
If it was not found, an error will return.
It returns a []string because the underlying HTML input field could be an array.
```go
p,err := req.Param("user")
```

### Params
Params returns all existing parameters.
It will return an error if there was something wrong with the parsing.
It returns a []string because the underlying HTML input field could be an array.

```go
p,err := req.Params()
```




## Response

### Usage:

```go
r := NewResponse(rw)
err := r.Render("json")
```

### AddData
Add a new key value pair for the response.

### Raw
Get the raw http.ResponseWriter. 

### Render

Render the response with the given render type.
At the momen only `Json` is defined.
If there is a problem an error will return

### renderJson
The header Content-Type gets set to `application/json` and all the response data will get marshaled to json.
If there is a problem an error will return

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-controller/issues
