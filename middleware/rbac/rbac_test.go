package rbac_test

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware"
	jwt2 "github.com/patrickascher/gofw/middleware/jwt"
	"github.com/patrickascher/gofw/middleware/rbac"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//Default Claim
type DefaultJWTClaim struct {
	User  string
	Email string
	Roles []string
	jwt.StandardClaims
}

func (c *DefaultJWTClaim) SetJid(id string) {
	c.Id = id
}
func (c *DefaultJWTClaim) SetIss(iss string) {
	c.Issuer = iss
}
func (c *DefaultJWTClaim) Iss() string {
	return c.Issuer
}
func (c *DefaultJWTClaim) SetAud(aud string) {
	c.Audience = aud
}
func (c *DefaultJWTClaim) Aud() string {
	return c.Audience
}
func (c *DefaultJWTClaim) SetSub(sub string) {
	c.Subject = sub
}
func (c *DefaultJWTClaim) Sub() string {
	return c.Subject
}
func (c *DefaultJWTClaim) SetIat(iat int64) {
	c.IssuedAt = iat
}
func (c *DefaultJWTClaim) SetExp(exp int64) {
	c.ExpiresAt = exp
}
func (c *DefaultJWTClaim) SetNbf(nbf int64) {
	c.NotBefore = nbf
}

type RoleService struct {
}

func (rs *RoleService) Allowed(uri string, action string, claims interface{}) bool {

	r := claims.(DefaultJWTClaim)

	for _, role := range r.Roles {
		if uri == "/" && role == "admin" {
			return true
		}
	}
	return false
}

func testRbac() rbac.Rbac {
	rs := RoleService{}

	rbac := rbac.Rbac{}
	rbac.SetRoleService(&rs)

	return rbac
}

func TestRbac_Middleware(t *testing.T) {
	rbac := testRbac()

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
	}

	// test request without any context - 401
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mw := middleware.New(rbac.Middleware)
	mw.Handle(handlerFunc)(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// test request with context and admin role which is allowed with the URI "/" - 200
	claim := DefaultJWTClaim{}
	claim.Roles = []string{"admin", "writer"}
	r, _ = http.NewRequest("GET", "/", nil)
	r.RequestURI = "/"
	r = r.WithContext(context.WithValue(r.Context(), jwt2.ContextName, claim))
	w = httptest.NewRecorder()
	mw.Handle(handlerFunc)(w, r)
	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)

	// test request with context and writer role which is not allowed with the URI "/" - 401
	claim = DefaultJWTClaim{}
	claim.Roles = []string{"writer"}
	r, _ = http.NewRequest("GET", "/", nil)
	r.RequestURI = "/"
	r = r.WithContext(context.WithValue(r.Context(), jwt2.ContextName, claim))
	w = httptest.NewRecorder()
	mw.Handle(handlerFunc)(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

//same tests as in TestRbac_Middleware
func TestRbac_MiddlewareJr(t *testing.T) {
	rbac := testRbac()

	handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	}

	// test request without any context - 401
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	var p []httprouter.Param

	mw := middleware.NewJR(rbac.MiddlewareJR)
	mw.Handle(handlerFunc)(w, r, p)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// test request with context and admin role which is allowed with the URI "/" - 200
	claim := DefaultJWTClaim{}
	claim.Roles = []string{"admin", "writer"}
	r, _ = http.NewRequest("GET", "/", nil)
	r.RequestURI = "/"
	r = r.WithContext(context.WithValue(r.Context(), jwt2.ContextName, claim))
	w = httptest.NewRecorder()
	p = []httprouter.Param{}
	mw.Handle(handlerFunc)(w, r, p)
	assert.Equal(t, "", w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)

	// test request with context and writer role which is not allowed with the URI "/" - 401
	claim = DefaultJWTClaim{}
	claim.Roles = []string{"writer"}
	r, _ = http.NewRequest("GET", "/", nil)
	r.RequestURI = "/"
	r = r.WithContext(context.WithValue(r.Context(), jwt2.ContextName, claim))
	w = httptest.NewRecorder()
	p = []httprouter.Param{}
	mw.Handle(handlerFunc)(w, r, p)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
