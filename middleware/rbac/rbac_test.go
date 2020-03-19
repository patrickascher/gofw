// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package rbac_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/patrickascher/gofw/middleware"
	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/patrickascher/gofw/middleware/rbac"
	"github.com/patrickascher/gofw/router"
	"github.com/stretchr/testify/assert"
)

type mockClaim struct {
	User  string
	Email string
	Roles []string
	jwt.Claim
}

type roleService struct {
}

func (rs *roleService) Allowed(uri string, action string, claims interface{}) bool {

	r := claims.(mockClaim)

	for _, role := range r.Roles {
		if uri == "/" && role == "admin" {
			return true
		}
	}
	return false
}

func TestRbac_MW(t *testing.T) {
	test := assert.New(t)

	// mock roleService
	rs := roleService{}

	// controller
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
	}

	// error: no role service is defined
	rbacMw := rbac.New(nil)
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mw := middleware.New(rbacMw.MW)
	mw.Handle(handlerFunc)(w, r)
	test.Equal(http.StatusInternalServerError, w.Code)

	// error: no claim is defined
	rbacMw = rbac.New(&rs)
	r, _ = http.NewRequest("GET", "/", nil)
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	w = httptest.NewRecorder()
	mw = middleware.New(rbacMw.MW)
	mw.Handle(handlerFunc)(w, r)
	test.Equal(http.StatusUnauthorized, w.Code)

	// ok: no claim is defined
	rbacMw = rbac.New(&rs)
	r, _ = http.NewRequest("GET", "/", nil)
	claim := mockClaim{}
	claim.Roles = []string{"admin", "writer"}
	r = r.WithContext(context.WithValue(r.Context(), jwt.CLAIM, claim))
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	w = httptest.NewRecorder()
	mw = middleware.New(rbacMw.MW)
	mw.Handle(handlerFunc)(w, r)
	test.Equal(http.StatusOK, w.Code)

	// error: role writer is not allowed
	rbacMw = rbac.New(&rs)
	r, _ = http.NewRequest("GET", "/", nil)
	claim = mockClaim{}
	claim.Roles = []string{"writer"}
	r = r.WithContext(context.WithValue(r.Context(), jwt.CLAIM, claim))
	r = r.WithContext(context.WithValue(r.Context(), router.PATTERN, "/"))
	w = httptest.NewRecorder()
	mw = middleware.New(rbacMw.MW)
	mw.Handle(handlerFunc)(w, r)
	test.Equal(http.StatusForbidden, w.Code)
}
