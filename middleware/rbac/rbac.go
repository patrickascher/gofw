// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package rbac provides a role based access control list.
// It is build on top of the JWT middleware.
// A roleService must be set, to check against your business logic.
//
//		rs := &RoleService{} // custom implementation
// 		rbac := New(rs)
// 		middleware.Add(jwt.MW,rbac.MW) // jwt.MW must be set before the rbac.MW middleware.
package rbac

import (
	"net/http"

	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/patrickascher/gofw/router"
)

// RoleService interface
type RoleService interface {
	// Allowed returns a boolean if the access is granted.
	// For the given url, HTTP method and jwt claim which includes specific user information.
	Allowed(url string, HTTPMethod string, claims interface{}) bool
}

// Rbac type
type Rbac struct {
	roleService RoleService
}

// New returns a rbac.
func New(r RoleService) *Rbac {
	return &Rbac{roleService: r}
}

// MW will be passed to the middleware.
func (rb *Rbac) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// checking the request context for the required keys
		claim := r.Context().Value(jwt.CLAIM)
		urlPattern := r.Context().Value(router.PATTERN)

		// application or configuration errors
		if rb.roleService == nil || urlPattern == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// normally the jwt.MW is taking care of this.
		// Its just here if a developer forgot to add the jwt.MW.
		if claim == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !rb.roleService.Allowed(urlPattern.(string), r.Method, claim) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h(w, r)
	}
}
