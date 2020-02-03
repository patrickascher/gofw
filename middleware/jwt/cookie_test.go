package jwt_test

import (
	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCookie_Create(t *testing.T) {
	cookie := jwt.Cookie{}
	httpCookie := cookie.Create("jwt", "test")

	assert.Equal(t, "jwt", httpCookie.Name)
	assert.Equal(t, "test", httpCookie.Value)

	assert.Equal(t, true, httpCookie.HttpOnly)
	assert.Equal(t, true, httpCookie.Secure)
}

func TestCookie_Get(t *testing.T) {
	cookie := jwt.Cookie{}

	// ok
	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	r.AddCookie(cookie.Create(jwt.CookieJwt, "test"))
	jwtString, err := cookie.Get(jwt.CookieJwt, r)
	assert.NoError(t, err)
	assert.Equal(t, "test", jwtString)

	//invalid - standard cookie name does not exist
	_, err = cookie.Get("abc", r)
	assert.Error(t, err)
}
