// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package jwt includes a parser, generator and a middleware to checks if a request-token is valid.
// If not a StatusUnauthorized (401) will return.
//
// Claims must implement the jwt.Claimer interface. Like this its easy to extend.
// A standard Claim is defined which can get embedded in your struct to avoid rewriting all of the functions.
//
// Config struct for a simple token configuration is provided.
//
// By default the claim of a valid token will be passed to the request context by the key jwt.CLAIM.
// Its searching the Token in the cookies by the key "jwt.CookieJwt". A setCookie function exits which can be used by the
// custom authentication implementation.
//
//		cfg := jwt.Config{Issuer: "mock", Alg: jMW.HS256, Subject: "test", Audience: "gotest", Duration: 10 * time.Minutes, SignKey: "secret"}
//		jtoken, err := jwt.New(cfg, &claim)
// 		middleware.Add(jtoken.MW)
package jwt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/segmentio/ksuid"
)

// CLAIM  is added to the request ctx.
const CLAIM = "JWT"

// Algorithms
const (
	HS256 = "HS256"
	HS384 = "HS384"
	HS512 = "HS512"
)

// Error messages.
var (
	ErrConfigNotValid = errors.New("jwt: config is not valid")
	ErrSigningMethod  = errors.New("jwt: unexpected signing method: %v")
	ErrInvalidClaim   = errors.New("jwt: token claims are invalid %v")
	ErrTokenExpired   = errors.New("jwt: token is expired")
)

var allowedAlg = []string{HS256, HS384, HS512}

// RefreshConfig - @TODO this should be handled in the middleware not in the controller, create custom functions.
type RefreshConfig struct {
	Duration time.Duration // the ttl of the refresh token. 0 = infinity (suggested long lived P30DT)
}

// Config of the jwt token.
type Config struct {
	Alg          string        // algorithm (HS256, HS384, HS512)
	Issuer       string        // issuer
	Audience     string        // audience
	Subject      string        // subject
	Duration     time.Duration // the ttl of the token (suggested short lived PT15M)
	SignKey      string        // the sign key. atm only a key, later on it can also be a file path
	RefreshToken RefreshConfig // true if a refresh token should get created
}

// valid checks all mandatory field and the allowed algorithm.
func (c Config) valid() bool {

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
	if !inSlice(allowedAlg, strings.ToUpper(c.Alg)) {
		return false
	}

	return true
}

// Token struct
type Token struct {
	keyFunc jwt.Keyfunc
	config  Config
	claim   Claimer
}

// New token instance.
// Error will return if the config is invalid.
func New(config Config, claim Claimer) (*Token, error) {
	t := &Token{}

	// adding config
	if !config.valid() {
		return nil, ErrConfigNotValid
	}
	t.config = config

	// adding claim to the token
	t.claim = claim

	// adding keyFunc
	switch strings.ToUpper(t.config.Alg) {
	case HS256, HS384, HS512:
		t.keyFunc = func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(ErrSigningMethod.Error(), token.Header["alg"])
			}
			return []byte(t.config.SignKey), nil
		}
	}

	return t, nil
}

// Generate a new token with the given claim.
// A signed token and refresh token will return.
// The refresh token is a uuid.
// Error will return if the token could not get signed.
func (t *Token) Generate(claim Claimer) (string, string, error) {

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
	switch strings.ToUpper(t.config.Alg) {
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

// Parse the token, the value will get passed to a new claimer.
// For more security reasons the ISS,SUB,AUD are getting matched against the config.
// Additionally the EXP,NBF and IAT are also getting checked because a custom Claimer could have overwritten the standard valid function.
func (t *Token) Parse(token string) (Claimer, error) {

	// creating a new struct of the custom claimer
	claim := reflect.New(reflect.TypeOf(t.claim).Elem()).Interface().(Claimer)

	// parse the token
	parsedToken, err := jwt.ParseWithClaims(token, claim, t.keyFunc)
	if err != nil {
		return nil, err
	}

	// compare token value with config
	claim = parsedToken.Claims.(Claimer)
	if claim.Iss() != t.config.Issuer ||
		claim.Sub() != t.config.Subject ||
		claim.Aud() != t.config.Audience ||
		parsedToken.Header["alg"].(string) != strings.ToUpper(t.config.Alg) {
		return claim, fmt.Errorf(ErrInvalidClaim.Error(), claim) // TODO remove claim... this is used because of the refresh token and auth atm. the logic should be moved to jwt package to avoid problems.
	}

	// standard validate function could have been overwritten.
	now := time.Now().Unix()
	if now > claim.Exp() ||
		now < claim.Nbf() ||
		now < claim.Iat() {
		return claim, ErrTokenExpired // TODO remove claim... this is used because of the refresh token and auth atm. the logic should be moved to jwt package to avoid problems.
	}

	return claim, nil
}

// MW will be passed to the middleware.
func (t *Token) MW(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
		h(w, r.WithContext(context.WithValue(r.Context(), CLAIM, claim)))
	}
}

// todo remove this if a framework package slice will be created
func inSlice(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
