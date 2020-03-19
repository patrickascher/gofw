// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

func TestCookie_Create(t *testing.T) {
	test := assert.New(t)

	cookie := jwt.Cookie{}
	httpCookie := cookie.Create("jwt", "test")

	test.Equal("jwt", httpCookie.Name)
	test.Equal("test", httpCookie.Value)
	// test.Equal(true, httpCookie.HttpOnly)
	// test.Equal(true, httpCookie.Secure)
	test.Equal("jwt", httpCookie.Name)
	test.True(time.Now().Before(httpCookie.Expires))
	test.Equal(5*60*60, httpCookie.MaxAge)
}

func TestCookie_Get(t *testing.T) {
	test := assert.New(t)

	cookie := jwt.Cookie{}

	// ok
	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(cookie.Create(jwt.CookieJwt, "test"))
	jwtString, err := cookie.Get(jwt.CookieJwt, r)
	test.NoError(err)
	test.Equal("test", jwtString)

	//invalid - cookie key does not exist
	_, err = cookie.Get("abc", r)
	test.Error(err)
	test.Equal(http.ErrNoCookie.Error(), err.Error())

}
