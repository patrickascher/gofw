// Package jwt is checking if the JWT cookie exists and if the token is valid, if not a 401 will return.
// It is available for the normal HandleFunc and the julienschmidt router.
// See https://github.com/patrickascher/go-middleware for more information and examples.
package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/segmentio/ksuid"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Cookie constants
const (
	CookieJwt     = "JWT_TOKEN"
	CookieRefresh = "REFRESH_TOKEN"
)

type key string

// ContextName for the added jwt claim
const ContextName = key("JWT")

// All Error variables are defined here.
var (
	ErrInvalidConfig = errors.New("jwt: config is not valid")
	ErrSigningMethod = errors.New("jwt: unexpected signing method: %v")
	ErrInvalidClaim  = errors.New("jwt: token claims are invalid %v")
	ErrTokenExpired  = errors.New("jwt: token is expired")
)

// Config of the jwt token.
type Config struct {
	Alg          string        // algorithm (HS256, HS384, HS512)
	Issuer       string        // issuer
	Audience     string        // audience
	Subject      string        // subject
	Duration     time.Duration // duration how long the access token should be valid (suggested short lived PT15M)
	SignKey      string        // the sign key. atm only a key, later on it can also be a file path
	RefreshToken RefreshConfig // true if a refresh token should get created
}

type RefreshConfig struct {
	Duration time.Duration // the ttl of the refresh token. If 0 = infinity (suggested long lived P30DT)
}

// valid checks if the config has all mandatory fields and if the alg is supported.
func (c *Config) valid() bool {

	// mandatory fields
	if c.Alg == "" ||
		c.Issuer == "" ||
		c.Audience == "" ||
		c.Subject == "" ||
		c.Duration == time.Duration(0) ||
		c.SignKey == "" {
		return false
	}

	// check if a supported alg is used
	c.Alg = strings.ToUpper(c.Alg)
	if !strings.Contains("HS256 HS384 HS512", c.Alg) {
		return false
	}

	return true
}

// StdClaim that implements the Claim interface.
type StdClaim struct {
	jwt.StandardClaims
}

// SetJid set the JID of the token.
func (c *StdClaim) SetJid(id string) {
	c.Id = id
}

// Jid get the JID of the token.
func (c *StdClaim) Jid() string {
	return c.Id
}

// SetIss set the ISSUER of the token.
func (c *StdClaim) SetIss(iss string) {
	c.Issuer = iss
}

// Iss get the ISSUER of the token.
func (c *StdClaim) Iss() string {
	return c.Issuer
}

// SetAud set the AUDIENCE of the token.
func (c *StdClaim) SetAud(aud string) {
	c.Audience = aud
}

// Aud get the AUDIENCE of the token.
func (c *StdClaim) Aud() string {
	return c.Audience
}

// SetSub set the SUBJECT of the token.
func (c *StdClaim) SetSub(sub string) {
	c.Subject = sub
}

// Sub get the SUBJECT of the token.
func (c *StdClaim) Sub() string {
	return c.Subject
}

// SetIat set the ISSUED AT of the token.
func (c *StdClaim) SetIat(iat int64) {
	c.IssuedAt = iat
}

// Iat get the ISSUED AT of the token.
func (c *StdClaim) Iat() int64 {
	return c.IssuedAt
}

// SetExp set the EXPIRED of the token.
func (c *StdClaim) SetExp(exp int64) {
	c.ExpiresAt = exp
}

// Exp get the EXPIRED of the token.
func (c *StdClaim) Exp() int64 {
	return c.ExpiresAt
}

// SetNbf set the NOT BEFORE of the token.
func (c *StdClaim) SetNbf(nbf int64) {
	c.NotBefore = nbf
}

// Nbf get the NOT BEFORE of the token.
func (c *StdClaim) Nbf() int64 {
	return c.NotBefore
}

// Render should return the needed claim data for the frontend.
func (c *StdClaim) Render() interface{} {
	return ""
}

// Claim interface
type Claim interface {
	SetJid(string)
	Jid() string
	SetIss(string)
	Iss() string
	SetAud(string)
	Aud() string
	SetSub(string)
	Sub() string
	Iat() int64
	SetIat(int64)
	Exp() int64
	SetExp(int64)
	Nbf() int64
	SetNbf(int64)

	Render() interface{}
	Valid() error
}

// Token is the main struct
type Token struct {
	keyFunc jwt.Keyfunc
	config  *Config
	claim   Claim
}

// NewToken creates a new instance of token.
// It first validates the given config.
// Then adds the keyFunc for HSxxx (only HMAC is supported at the moment)
func NewToken(config *Config, claim Claim) (*Token, error) {
	t := &Token{}

	// adding config
	if !config.valid() {
		return nil, ErrInvalidConfig
	}
	t.config = config

	// adding claim to the token
	t.claim = claim

	// adding KeyFunc, if there would be a different ALG it would already fail in the config.valid function
	switch t.config.Alg {
	case "HS256", "HS384", "HS512":
		t.keyFunc = func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(ErrSigningMethod.Error(), token.Header["alg"])
			}
			return []byte(t.config.SignKey), nil
		}
	}

	return t, nil
}

// Generate a new token with the given claim
// ALG has to be HS256, HS384 or HS512
func (t *Token) Generate(claim Claim) (string, string, error) {

	// setting default claim values
	claim.SetJid(ksuid.New().String())                     // Token ID
	claim.SetIat(time.Now().Unix())                        // IAT
	claim.SetNbf(time.Now().Unix())                        // NBF
	claim.SetExp(time.Now().Add(t.config.Duration).Unix()) // EXP
	claim.SetIss(t.config.Issuer)                          // ISS
	claim.SetSub(t.config.Subject)                         // Sub
	claim.SetAud(t.config.Audience)                        // AUD

	// creating token - no other value supported and would fail already at NewToken
	var token *jwt.Token
	switch t.config.Alg {
	case "HS256":
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	case "HS384":
		token = jwt.NewWithClaims(jwt.SigningMethodHS384, claim)
	case "HS512":
		token = jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	}

	// signing token
	tokenString, err := token.SignedString([]byte(t.config.SignKey))
	if err != nil {
		return "", "", err
	}

	// random string for uuid
	refreshToken := ksuid.New()
	return tokenString, refreshToken.String(), nil
}

// Parse the given token.
// First the jwt library validates the token and as addition security reasons, its also checked against our configuration.
func (t *Token) Parse(token string) (Claim, error) {

	claim := t.newClaim()
	parsedToken, err := jwt.ParseWithClaims(token, claim, t.keyFunc)
	if err != nil {
		return nil, err
	}

	// additional Security
	// check if the Claim has the ISS,AUD,SUB and ALG as configured
	claim = parsedToken.Claims.(Claim)
	if claim.Iss() != t.config.Issuer ||
		claim.Sub() != t.config.Subject ||
		claim.Aud() != t.config.Audience ||
		parsedToken.Header["alg"].(string) != strings.ToUpper(t.config.Alg) {
		return claim, fmt.Errorf(ErrInvalidClaim.Error(), claim)
	}

	// additional Security
	// It could be that the custom claim has his own Valid function and then the Valid of jwt.StandardClaim is getting overwritten.
	// That's why here we are checking again for EXP,NBF and IAT
	now := time.Now().Unix()
	if now > claim.Exp() ||
		now < claim.Nbf() ||
		now < claim.Iat() {
		return claim, ErrTokenExpired
	}

	return claim, nil
}

// newClaim creates a new instance of the user claim
func (t *Token) newClaim() Claim {
	return reflect.New(reflect.TypeOf(t.claim).Elem()).Interface().(Claim)
}

// MiddlewareJR for the julienschmidt router
func (t *Token) MiddlewareJR(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		t.jwt(h, w, r, ps)
	}
}

//Middleware for a normal http.HandlerFunc
func (t *Token) Middleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.jwt(h, w, r)
	}
}

// jwt is a helper function to handel julienschmidt router and a normal http HandlerFunc
func (t *Token) jwt(handler interface{}, args ...interface{}) {

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

	// get the token from cookie
	cookie := Cookie{}
	token, err := cookie.Get(CookieJwt, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// parse token
	claim, err := t.Parse(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO if parse response with an expired, check config for refresh token.
	// if refresh token allowed, generate a new token (where to get the claim data???)
	// TODO refresh TOKEN if expired and Refresh Key exists

	// adding claim to the request context
	req := r.WithContext(context.WithValue(r.Context(), ContextName, claim))

	if reflect.TypeOf(handler).String() == "httprouter.Handle" {
		hJR = handler.(httprouter.Handle)
		hJR(w, req, ps)
	} else {
		h = handler.(http.HandlerFunc)
		h(w, req)
	}
}
