<h1>go-cache</h1>

go-cache is a cache manager. Its provides a cache interface and is easy to extend.
At the moment only a in-memory cache is build

## Install

```go
go get github.com/fullhouse-productions/go-cache
```

## Usage

```go 
import "github.com/fullhouse-productions/go-cache"

c, err := cache.Get(MEMORY, 3600*time.Second)

err := c.Set("key","value")
err = c.Get("key") //value
if c.Exist("key"){
// ...
}
c.GetAll() map[string]Item 
err = c.Delete("key")
err = c.DeleteAll()
// c.GC(3600*time.Second) // this is automatically called after you the cache.Get()

```

## Interface
To create your own cache backend, you have to implement the cache interface

```go
type Cache interface {
	Get(key string) (Item, error)
	GetAll() map[string]Item
	Set(key string, value interface{}, timeout time.Duration) error
	Exist(key string) bool
	Delete(key string) error
	DeleteAll() error
	GC(duration time.Duration) error
}

type Item interface {
	Value() interface{}
	Lifetime() time.Duration
}
```

## Register
Register is used to register the cache backend. 
This function should be called in the init function of the cache-backend to register itself on import.
It returns an error if the cache-name or the cache itself is empty

```go
// init registers a memory backend
func init() {
	Register("memory", &Memory{})
}
```

**Usage**
```go
import "github.com/fullhouse-productions/go-cache"
import _ "your/repo/cache/memory"
```

## Get
With `Get()` you get "create" an instance of the cache-backend by its name. It returns a ptr to the cache-backend.
The first parameter is the name of the cache-backend and the second one is a duration.

This duration parameter is used for the garbage collector, which runs in a loop every x duration.

**TODO** Think about a different naming for the whole, Register and Get functions for all packages

?> `Get()` creates the backend instance. This means no memory is wasted before.

# In-Memory Backend

All the values are stored in memory. This means after a restart of the machine the data will be gone.
All Write and Read oprations are getting locked over `sync.RWMutex` to avoid data race.

The GC will only spawn once to avoid problems.

**Get**: If the Key does not exist, an error will return 



?> If the ttl is set to `0`, the value will never expire.

!> Make sure you apply a good duration to the garbage collector. The Value should be something like `5*time.Seconds` if you just enter `5`, which is possible, it would run every 5 nanoseconds.

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-cache/issues
