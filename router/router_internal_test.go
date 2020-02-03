package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	cachepackage "github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

// mock cache entry. TODO better solution for this
type cache struct {
}

type cacheItem struct {
}

func (c *cache) Get(string) (cachepackage.Item, error) {
	return &cacheItem{}, nil
}
func (c *cache) GetAll() map[string]cachepackage.Item {
	return make(map[string]cachepackage.Item)
}
func (c *cache) Set(key string, value interface{}, timeout time.Duration) error {
	return nil
}
func (c *cache) Exist(key string) bool {
	return true
}
func (c *cache) Delete(key string) error {
	return nil
}
func (c *cache) DeleteAll() error {
	return nil
}
func (c *cache) GC(duration time.Duration) error {
	return nil
}
func (c *cacheItem) Value() interface{} {
	return ""
}
func (c *cacheItem) Lifetime() time.Duration {
	return 5 * time.Millisecond
}

type TestMiddleware struct {
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) JWT(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Jwt")
	}
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) Rbac(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Rbac")
	}
}

//Auth adds a jwt and rbac to the request
func (tm *TestMiddleware) Logger(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Println("called Logger")
	}
}

func TestManager_Cache(t *testing.T) {
	cache := &cache{}
	rm := Manager{}
	rm.Cache(cache)

	assert.Equal(t, cache, rm.cache)
}

func TestManager_AllowHttpMethod(t *testing.T) {
	rm := Manager{}
	rm.allowedHTTPMethod = map[string]bool{
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

	err := rm.AllowHTTPMethod(GET, false)
	assert.NoError(t, err)
	assert.Equal(t, false, rm.allowedHTTPMethod[GET])
	assert.Equal(t, true, rm.allowedHTTPMethod[POST])

	err = rm.AllowHTTPMethod("something", false)
	assert.Error(t, err)

}

func TestManager_SecureMiddleware(t *testing.T) {
	rm := Manager{}
	mw := TestMiddleware{}
	rm.SecureMiddleware(middleware.NewJR(mw.JWT, mw.Rbac))

	assert.Equal(t, 2, len(rm.secureMiddleware.GetAll()))
}
