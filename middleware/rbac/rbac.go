// Package rbac is offering a Role based access control list.
// It is based on the JWT middleware and checks the request context for the jwt.Claim.
package rbac

import (
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware/jwt"
	"net/http"
	"reflect"
	"strings"
)

// RoleService interface
type RoleService interface {
	Allowed(resource string, action string, claims interface{}) bool
}

// Rbac main type
type Rbac struct {
	roleService RoleService
}

// SetRoleService set your own RoleService
func (rb *Rbac) SetRoleService(r RoleService) {
	rb.roleService = r
}

// MiddlewareJR for the julienschmidt router
func (rb *Rbac) MiddlewareJR(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		rb.rbac(h, w, r, ps)
	}
}

//Middleware for a normal http.HandlerFunc
func (rb *Rbac) Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rb.rbac(h, w, r)
	}
}

// rbac middleware helper
func (rb *Rbac) rbac(handler interface{}, args ...interface{}) {

	var w http.ResponseWriter
	var r *http.Request
	var ps httprouter.Params
	var hJR httprouter.Handle
	var h http.HandlerFunc

	for k, arg := range args {
		switch k {
		case 0:
			w = arg.(http.ResponseWriter)
		case 1:
			r = arg.(*http.Request)
		case 2:
			ps = arg.(httprouter.Params)
		}

	}

	// checking if ctx exist
	ctx := r.Context().Value(jwt.ContextName)
	if ctx == nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//TODO this logic exists also in controller.SetParams -> controller.pattern.
	// move this in the controller or at least write a helper function there
	var uri string
	uri = r.RequestURI
	if len(ps) > 0 {
		uri = r.RequestURI
		for _, val := range ps {
			uri = strings.Replace(uri, val.Value, ":"+val.Key, 1)
		}

		// check if the param is a wildcard.
		// if there is no slash between the key param, then it is a wildcard
		// its only working with rules like /roles/*grid = /roles/param1/param2... not with /roles/?param1=xxx
		if !strings.Contains(uri, "/:"+ps[len(ps)-1].Key) {
			uri = strings.Replace(uri, ":"+ps[len(ps)-1].Key, "/*"+ps[len(ps)-1].Key, 1)
		}
	}

	// check if permission is granted
	if rb.roleService == nil || !rb.roleService.Allowed(uri, r.Method, ctx) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if reflect.TypeOf(handler).String() == "httprouter.Handle" {
		hJR = handler.(httprouter.Handle)
		hJR(w, r, ps)
	} else {
		h = handler.(http.HandlerFunc)
		h(w, r)
	}
}
