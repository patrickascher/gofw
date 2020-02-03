package jwt_test

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware"
	JWTmiddleware "github.com/patrickascher/gofw/middleware/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

//Default Claim
type Claim struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Permission []string `json:"permission"`
	jwt.StandardClaims
}

func (c *Claim) SetJid(id string) {
	c.Id = id
}
func (c *Claim) Jid() string {
	return c.Id
}
func (c *Claim) SetIss(iss string) {
	c.Issuer = iss
}
func (c *Claim) Iss() string {
	return c.Issuer
}
func (c *Claim) SetAud(aud string) {
	c.Audience = aud
}
func (c *Claim) Aud() string {
	return c.Audience
}
func (c *Claim) SetSub(sub string) {
	c.Subject = sub
}
func (c *Claim) Sub() string {
	return c.Subject
}
func (c *Claim) SetIat(iat int64) {
	c.IssuedAt = iat
}
func (c *Claim) Iat() int64 {
	return c.IssuedAt
}
func (c *Claim) SetExp(exp int64) {
	c.ExpiresAt = exp
}
func (c *Claim) Exp() int64 {
	return c.ExpiresAt
}
func (c *Claim) SetNbf(nbf int64) {
	c.NotBefore = nbf
}
func (c *Claim) Nbf() int64 {
	return c.NotBefore
}

func (c *Claim) Render() interface{} {
	return ""
}

type ClaimWithValidFunc struct {
	Claim
}

func (c *ClaimWithValidFunc) Valid() error {
	return nil
}

func TestNewToken(t *testing.T) {
	claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	assert.NoError(t, err)
	assert.IsType(t, &JWTmiddleware.Token{}, token)

	// error because config is not valid (subject missing)
	claim = &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg = &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	_, err = JWTmiddleware.NewToken(cfg, claim)
	assert.Error(t, err)
	assert.Equal(t, JWTmiddleware.ErrInvalidConfig, err)

	// error because alg is not allowed
	claim = &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg = &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "RS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	_, err = JWTmiddleware.NewToken(cfg, claim)
	assert.Error(t, err)
	assert.Equal(t, JWTmiddleware.ErrInvalidConfig, err)
}

func TestToken_Generate(t *testing.T) {

	algs := []string{"HS256", "HS384", "HS512", "hs256", "hs384", "hs512"}

	for _, alg := range algs {
		claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
		cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: alg, Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
		token, err := JWTmiddleware.NewToken(cfg, claim)
		if assert.NoError(t, err) {
			jwtToken, _, err := token.Generate(claim)
			if assert.NoError(t, err) {
				// test if the claim got the right values
				assert.Equal(t, "Wall-E", claim.Name)
				assert.Equal(t, "fullhouse-productions.com", claim.Iss())
				assert.Equal(t, "tests", claim.Sub())
				assert.Equal(t, "go test", claim.Aud())
				assert.Equal(t, "walle@example.com", claim.Email)
				assert.Equal(t, "Wall-E", claim.Name)
				assert.Equal(t, []string{"/dashboard", "/settings"}, claim.Permission)
				assert.True(t, claim.ExpiresAt > 0)
				assert.True(t, claim.NotBefore > 0)
				assert.True(t, claim.IssuedAt > 0)
				assert.True(t, len(claim.Jid()) > 0)
				assert.IsType(t, "string", jwtToken)
			}
		}
	}
}

func TestToken_Parse_HMAC(t *testing.T) {

	algs := []string{"HS256", "HS384", "HS512"}

	for i, alg := range algs {
		claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
		cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: alg, Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
		token, err := JWTmiddleware.NewToken(cfg, claim)
		if assert.NoError(t, err) {

			if i == 0 {
				HS256 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.gvp43uET2FTew19wpzBD9AcgJ1HeS1clwDwpU-BBdoA"
				hs256Claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}, StandardClaims: jwt.StandardClaims{Audience: "go test", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053756, Issuer: "fullhouse-productions.com", NotBefore: 1549053756, Subject: "tests"}}
				claim, err := token.Parse(HS256)
				assert.NoError(t, err)
				assert.Equal(t, hs256Claim, claim)

				// jwt modified (deleted sub)
				HS256EmptySub := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6IiJ9.5j6-uoaeLbTFuAiDltouyimZoqecHffKlDpjgpg4Z8o"
				claim, err = token.Parse(HS256EmptySub)
				assert.Error(t, err)
				assert.Equal(t, fmt.Errorf(JWTmiddleware.ErrInvalidClaim.Error(), claim), err)

				//expired
				HS256Expired := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoxNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.BzfR0ttqm_NhkfxM3aEUDigda7aFRFmajfoghH023_c"
				_, err = token.Parse(HS256Expired)
				assert.Error(t, err)
				assert.Equal(t, true, strings.HasPrefix(err.Error(), "token is expired by "))

			}

			if i == 1 {
				HS384 := "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcXBxa3NUOHZyRkVpRE9KcDZDZnVjWWYiLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.eVGzxdHiLauxdH4h54fmT1JCG00n2T0EE_-GSgiEAsDNc13N8cXMQdDJB_eIttL2"
				hs384Claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}, StandardClaims: jwt.StandardClaims{Audience: "go test", ExpiresAt: 2549053806, Id: "1GapqpqksT8vrFEiDOJp6CfucYf", IssuedAt: 1549053756, Issuer: "fullhouse-productions.com", NotBefore: 1549053756, Subject: "tests"}}
				claim, err := token.Parse(HS384)
				assert.NoError(t, err)
				assert.Equal(t, hs384Claim, claim)
			}

			if i == 2 {
				HS512 := "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcXBVT281bHJXaTgyRTR6SXR0UndHWUciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.9KSSmukVseyzeXHpC9B81WML3rl1Guu203Iawb64_3j-j9LZs7xpba-z2WmR0p_SpkYC-7M64AQHI0VTVzS8gw"
				hs512Claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}, StandardClaims: jwt.StandardClaims{Audience: "go test", ExpiresAt: 2549053806, Id: "1GapqpUOo5lrWi82E4zIttRwGYG", IssuedAt: 1549053756, Issuer: "fullhouse-productions.com", NotBefore: 1549053756, Subject: "tests"}}
				claim, err := token.Parse(HS512)
				assert.NoError(t, err)
				assert.Equal(t, hs512Claim, claim)
			}

		}
	}

	// claim has his own valid function so EXP, IAT, NBF is not checked automatically
	claim := &ClaimWithValidFunc{}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	assert.NoError(t, err)
	HS256Expired := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoxNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.BzfR0ttqm_NhkfxM3aEUDigda7aFRFmajfoghH023_c"
	_, err = token.Parse(HS256Expired)
	assert.Error(t, err)
	assert.Equal(t, JWTmiddleware.ErrTokenExpired, err)
}

func TestToken_Middleware_OK(t *testing.T) {

	// trying with no cookie
	claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	if assert.NoError(t, err) {

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("secure area-" + r.Context().Value(JWTmiddleware.ContextName).(*Claim).Name))
		}

		r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
		HS256 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.gvp43uET2FTew19wpzBD9AcgJ1HeS1clwDwpU-BBdoA"

		cookie := &http.Cookie{HttpOnly: true, Name: JWTmiddleware.CookieJwt, Value: HS256}
		r.AddCookie(cookie)

		w := httptest.NewRecorder()
		mw := middleware.New(token.Middleware)
		mw.Handle(handlerFunc)(w, r)

		assert.Equal(t, "secure area-Wall-E", w.Body.String())
		assert.Equal(t, 200, w.Code)
	}
}

func TestToken_Middleware_NoCookie(t *testing.T) {

	// trying with no cookie
	claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	if assert.NoError(t, err) {

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("secure area"))
		}

		r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)

		w := httptest.NewRecorder()
		mw := middleware.New(token.Middleware)
		mw.Handle(handlerFunc)(w, r)

		assert.Equal(t, "", w.Body.String())
		assert.Equal(t, 401, w.Code)
	}
}

func TestToken_Middleware_wrongCookie(t *testing.T) {

	// trying with no cookie
	claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	if assert.NoError(t, err) {

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("secure area"))
		}

		r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
		HS256 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.gvp43uET2FTew19wpzBD9AcgJ1HeS1clwDwpU-BBdoA"
		cookie := &http.Cookie{Name: JWTmiddleware.CookieJwt, Value: HS256, HttpOnly: true}
		r.AddCookie(cookie)

		w := httptest.NewRecorder()
		mw := middleware.New(token.Middleware)
		mw.Handle(handlerFunc)(w, r)

		assert.Equal(t, "", w.Body.String())
		assert.Equal(t, 401, w.Code)
	}
}

func TestToken_MiddlewareJR_OK(t *testing.T) {

	// trying with no cookie
	claim := &Claim{Name: "Wall-E", Email: "walle@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := &JWTmiddleware.Config{Issuer: "fullhouse-productions.com", Alg: "HS256", Subject: "tests", Audience: "go test", Duration: 50 * time.Second, SignKey: "secret"}
	token, err := JWTmiddleware.NewToken(cfg, claim)
	if assert.NoError(t, err) {

		handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			w.WriteHeader(200)
			w.Write([]byte("secure area"))
		}

		r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
		HS256 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiV2FsbC1FIiwiZW1haWwiOiJ3YWxsZUBleGFtcGxlLmNvbSIsInBlcm1pc3Npb24iOlsiL2Rhc2hib2FyZCIsIi9zZXR0aW5ncyJdLCJhdWQiOiJnbyB0ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImZ1bGxob3VzZS1wcm9kdWN0aW9ucy5jb20iLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3RzIn0.gvp43uET2FTew19wpzBD9AcgJ1HeS1clwDwpU-BBdoA"
		cookie := &http.Cookie{Name: JWTmiddleware.CookieJwt, Value: HS256, HttpOnly: true}
		r.AddCookie(cookie)

		w := httptest.NewRecorder()
		mw := middleware.NewJR(token.MiddlewareJR)
		var p []httprouter.Param

		mw.Handle(handlerFunc)(w, r, p)

		assert.Equal(t, "secure area", w.Body.String())
		assert.Equal(t, 200, w.Code)
	}
}
