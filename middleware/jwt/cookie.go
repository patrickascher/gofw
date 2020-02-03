package jwt

import (
	"net/http"
	"time"
)

// Cookie is used to transfer the jwt
type Cookie struct {
}

// Create a new cookie with the given name and value.
// Additionally some default security is added.
func (c *Cookie) Create(name string, v string) *http.Cookie {
	cookie := &http.Cookie{}
	cookie.Name = name
	cookie.Value = v

	//cookie.HttpOnly = true // not available for JS
	//cookie.Secure = true   // send only over HTTPS

	// maxAge and expires is set (for old ie browsers)
	cookie.Expires = time.Now().Add(5 * time.Hour) //GMT/UTC is handled by internals
	cookie.MaxAge = 5 * 60 * 60

	return cookie
}

// Get the token from the request
func (c *Cookie) Get(name string, r *http.Request) (string, error) {
	// get the token from cookie
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}
