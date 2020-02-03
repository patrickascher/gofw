<h1>go-server</h1>

Server is a default http Server with some hooks.

## Install

```go
go get github.com/fullhouse-productions/go-server
```

## Usage

```go

	// Getting the config
	err := cfg.Parse(cfg.JSON, &config.Config, &reader.JsonOptions{File: "backend/config/app.json"})
	if err != nil {
		fmt.Println(err)
	}

	// initialize the server & hooks
	err = go_server.Initialize(&config.Config, go_server.LOGGER, go_server.BUILDER, go_server.CACHE, go_server.ROUTER)
	if err != nil {
		fmt.Println(err)
	}

	// run the application
	err = go_server.Run()
	if err != nil {
		fmt.Println(err)
	}
```

# Server 

## Initialize
Initialize is used to set-up the server with some hooks.
As argument the needed hooks can get defined. For each hook a constant is defined.

```go

// This would initialize the hooks `Logger`, `Builder` and `Router`.
Initialize(&cfg, LOGGER, BUILDER, ROUTER)
```

Each hook can get used like this in the application later on:

```go

// your application

logger := server.Logger()
builder := server.Builder()
router := server.Router()
//...

```

## Run
Is starting the HTTP/HTTPS server. If `ForceHTTPS` is set, all HTTP requests will get redirected to HTTPS.

?> Check some best practice configs.

# Config

Config has some default structs defined.
You have to embed this struct into your application configuration struct.

**example**
```go
// your config struct
type AppConfig struct{
	server.Cfg
	Title string
	//...
}
```

**defined in server module**
```go
type Cfg struct {
	Database    *sqlquery.Config    `json:"database"`
	Server       Server        `json:"server"`
	Router       Router        `json:"router"`
	CacheManager CacheProvider `json:"cache"`
}

type Server struct {
	HTTPPort   int    `json:"httpPort"`
	HTTPSPort  int    `json:"httpsPort"`
	ForceHTTPS bool   `json:"forceHttps"`
	CertFile   string `json:"certFile"`
	KeyFile    string `json:"keyFile"`
}

type Router struct {
	Provider    string      `json:"provider"`
	Favicon     string      `json:"favicon"`
	PublicDirs []Directory `json:"directories"`
}

type Directory struct {
	Url    string `json:"url"`
	Source string `json:"source"`
}

type CacheProvider struct {
	Provider string `json:"provider"`
	GCCycle  int64  `json:"cycle"`
}
```


# Hooks

Hooks can be loaded by the `server.Initialize()` method.
This is a list with the pre-defined hooks.

## Logger
Logger is defined by default. It is returning the default logger which is a console logger.

## Cache
If a CacheProvider is defined, a cache will be created. 
The GCCycle must be set in seconds and is the time when the garbage collector is running every x seconds.

## Builder
If a database is defined, a global Builder will be created.

?> At the moment only one database is config-able. Its easy to change this by simply create a slice out of the Database struct and make minimal changes on the `Builder` method.

## Router
If a router is defined, a RouterManager will be created.




# Issues & Ideas

- [ ] Proposal: Initialize should return a server Struct. on the struct u can call Cache, Builde,Logger,... and Start.

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-server/issues
