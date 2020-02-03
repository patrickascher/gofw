<h1>go-config</h1>

go-config is a module to handel different configurations. 
It can handle different configuration based on the environment.
Its simple to create different readers, to handle different config backends.
Out of the box json is supported.

<h1>Features:</h1>

 - Config Interface
 - Easy to extend
 - Env based

# Config

Config is using a globals reader store.
This means you have to register your own readers before you can use them.

Out of the box, a json reader is available.

## Install

```go
go get github.com/fullhouse-productions/go-config
```

## Usage

```go
import "github.com/fullhouse-productions/go-config"
import _ "github.com/fullhouse-productions/go-config/json"

type Cfg struct{
	Adapter string
	Host string
	//...
}

cfg := Cfg{}
err := config.Parse(JSON,&cfg,Options{Path:"config/conf.json"})
//...
```

?> By default the environment variable `ENV` is taken. You can choose a different one through the options in the `Parse` function


## Interface
To create a own reader, you have to implement the config Interface

```go
type config interface {
	Parse(config interface{}, options options) error
	Env(string)
} 
```

## Register
Register is used to register the reader. 
This function should be called in the init function of the reader to register itself on import.
It returns an error if the reader-name or the reader itself is empty

```go
// init registers a json reader
func init() {
	Register("json", &Json{})
}
```

**Usage**
```go
import "github.com/fullhouse-productions/go-config"
import _ "your/repo/config/reader"
```

## Parse
Parse is calling the Parse function of the reader.
It will return an error if the reader Parse does so.

## IsSet
IsSet checks recursively if a field is existing and has "no" zero value in a struct.
If a zero value should be allowed, prefix the field name with a 0.
This can be used to check if a specific configuration exists. 

At the moment only structs are supported.
```go
// no zero values allowed
if config.IsSet("User.Role.Name",cfg){
	//...
}

// zero values allowed
if config.IsSet("0User.Role.Name",cfg){
	//...
}
```

# Json Provider

The json reader is loading the given file if it exists.
It also checks if a file with the environment as prefix exists. If so it will get loaded and merged.

Example:
Main file `config.json` will get loaded, then it checks if the `{env}.config.json` file exists and tries to merge it together.

!> Parser will return only an error if the main file does not exist.

!> The option param must be a Ptr

## Options

| Option      | example            | description |
|-------------|--------------------|-------------|
| File | "config/config.json" | Path to the json file



## Usage

```go
import "github.com/fullhouse-productions/go-config"
import _ "github.com/fullhouse-productions/go-config/json"

type Cfg struct{
	Adapter string
	Host string
	//...
}

cfg := Cfg{}
err := config.Parse(JSON,&cfg,&Options{File:"config/conf.json"})
//...
```

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-config/issues
