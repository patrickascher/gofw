// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package jwt_test

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/patrickascher/gofw/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jMW "github.com/patrickascher/gofw/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

//Default Claim
type mockClaim struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Permission []string `json:"permission"`
	jMW.Claim
}

var errCustom = errors.New("custom claim error")

type mockClaimWithValidFunc struct {
	mockClaim
}

func (m *mockClaimWithValidFunc) Valid() error {
	return errCustom
}

type mockClaimWithValidFunc2 struct {
	mockClaim
}

func (m *mockClaimWithValidFunc2) Valid() error {
	return nil
}

func TestNew(t *testing.T) {
	test := assert.New(t)

	tests := []struct {
		claim    mockClaim
		config   jMW.Config
		error    bool
		errorMsg string
	}{
		// ok
		{
			claim:  mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "settings"}},
			config: jMW.Config{Issuer: "mock", Alg: jMW.HS256, Subject: "test#1", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"},
		},
		// ok: alg lower case
		{
			claim:  mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "settings"}},
			config: jMW.Config{Issuer: "mock", Alg: "hs256", Subject: "test#2", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"},
		},
		// err: config subject missing
		{
			claim:    mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "settings"}},
			config:   jMW.Config{Issuer: "mock", Alg: jMW.HS256, Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"},
			error:    true,
			errorMsg: "",
		},
		// err: config alg is not allowed
		{
			claim:    mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "settings"}},
			config:   jMW.Config{Issuer: "mock", Alg: "HS100", Subject: "test#3", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"},
			error:    true,
			errorMsg: "",
		},
	}

	for _, tt := range tests {
		token, err := jMW.New(tt.config, &tt.claim)
		if tt.error {
			test.Error(err)
			test.Nil(token)
		} else {
			test.IsType(&jMW.Token{}, token)
		}
	}
}

func TestToken_Generate(t *testing.T) {

	test := assert.New(t)

	tests := []struct {
		alg      string
		error    bool
		errorMsg string
	}{
		// ok
		{alg: jMW.HS256},
		{alg: jMW.HS384},
		{alg: jMW.HS512},
		{alg: "hs256"},
		{alg: "hs384"},
		{alg: "hs512"},
		// alg is not allowed
		{alg: "hs100", error: true},
	}

	timeExecuted := time.Now().Unix()

	for _, tt := range tests {
		claim := &mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}}
		cfg := jMW.Config{Issuer: "mock", Alg: tt.alg, Subject: "test#1", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"}
		token, err := jMW.New(cfg, claim)

		if tt.error {
			test.Error(err)
			test.Nil(token)
		} else {
			if test.NoError(err) {
				jwtToken, rToken, err := token.Generate(claim)
				test.NoError(err)
				// test if the claim got the right values
				test.Equal("John Doe", claim.Name)
				test.Equal(cfg.Issuer, claim.Iss())
				test.Equal(cfg.Subject, claim.Sub())
				test.Equal(cfg.Audience, claim.Aud())
				test.Equal("john@example.com", claim.Email)
				test.Equal([]string{"/dashboard", "/settings"}, claim.Permission)
				test.True(timeExecuted+10 >= claim.Exp())
				test.True(timeExecuted >= claim.Nbf())
				test.True(timeExecuted >= claim.Iat())
				test.True(len(claim.Jid()) > 0)

				test.True(len(jwtToken) > 0)
				test.True(len(rToken) > 0)

				test.Equal("", claim.Render())
			}
		}
	}
}

func TestToken_Parse(t *testing.T) {

	test := assert.New(t)

	// All tokens have the expired date 2050
	// Token between HS256, HS384 and HS512 have different IAT time set.
	tests := []struct {
		alg           string
		token         string
		expectedClaim mockClaim
		error         bool
		errorMsg      string
	}{
		// ok
		{
			alg:           jMW.HS256,
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.zENdytJIyGJ9HoEdq4CfC4vYlOVWCLp7P8mlTj2XhC0",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053756, Issuer: "mock", NotBefore: 1549053756, Subject: "test"}}},
		},
		// err: Token SUB got hijacked
		{
			alg:           jMW.HS256,
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6ImhpamFjayJ9.evT4KEYMnWb3AMnBtelru-cF88OffBaarsndDU83-7Y",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053756, Issuer: "mock", NotBefore: 1549053756, Subject: "test"}}},
			error:         true,
			errorMsg:      jMW.ErrInvalidClaim.Error(),
		},
		// err: Token ISS got hijacked
		{
			alg:           jMW.HS256,
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6ImhpamFjayIsIm5iZiI6MTU0OTA1Mzc1Niwic3ViIjoidGVzdCJ9.PcEcyGmqcOSCOPGuiHtU9Tg2gb1g3eUAcnPWa7tJKCs",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053756, Issuer: "mock", NotBefore: 1549053756, Subject: "test"}}},
			error:         true,
			errorMsg:      jMW.ErrInvalidClaim.Error(),
		},
		// err: Token AUD got hijacked
		{
			alg:           jMW.HS256,
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiaGlqYWNrIiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.DpuvoxFOH2Abma5LE0c7RecwqeQcfNFHfugf6584fm8",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053756, Issuer: "mock", NotBefore: 1549053756, Subject: "test"}}},
			error:         true,
			errorMsg:      jMW.ErrInvalidClaim.Error(),
		},
		{
			alg:           jMW.HS384,
			token:         "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTcsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTcsInN1YiI6InRlc3QifQ.8_4VApZL-DgHhOyC8d5Mbbdsk_uMMmjioKCK0bijJYqfcWNRGxH5vsll6Z86onqS",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053757, Issuer: "mock", NotBefore: 1549053757, Subject: "test"}}},
		},
		{
			alg:           jMW.HS512,
			token:         "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTgsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTgsInN1YiI6InRlc3QifQ.fTrLG8jkCEABrjIUhIYDV3y-PrdpGae5upR5JvxaF-wTgtErdpbcieJk9XQ2-g7-VccZHqbDgAQjEonedhsObQ",
			expectedClaim: mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}, Claim: jMW.Claim{StandardClaims: jwt.StandardClaims{Audience: "gotest", ExpiresAt: 2549053806, Id: "1GapqlYE9GYpGJfMSPrCUuuPN1w", IssuedAt: 1549053758, Issuer: "mock", NotBefore: 1549053758, Subject: "test"}}},
		},
	}

	for _, tt := range tests {
		c := &mockClaim{}
		cfg := jMW.Config{Issuer: "mock", Alg: tt.alg, Subject: "test", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"}
		token, err := jMW.New(cfg, c)

		if test.NoError(err) {
			claim, err := token.Parse(tt.token)
			if tt.error {
				test.Error(err)
				//test.Nil(claim) //TODO it should be nil after the refresh logic is moved to jwt package
				test.Equal(fmt.Sprintf(tt.errorMsg, claim), err.Error())
			} else {
				test.NoError(err)
				test.Equal(&tt.expectedClaim, claim)
			}
		}
	}

	// claim has his own valid function so EXP, IAT, NBF is not checked automatically
	claim := &mockClaimWithValidFunc{}
	cfg := jMW.Config{Issuer: "mock", Alg: jMW.HS256, Subject: "test", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"}
	token, err := jMW.New(cfg, claim)
	if test.NoError(err) {
		t := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.zENdytJIyGJ9HoEdq4CfC4vYlOVWCLp7P8mlTj2XhC0"
		_, err = token.Parse(t)
		test.Error(err)
		test.Equal(errCustom.Error(), err.Error())
	}

	// claim where the custom valid is true but the token is expired
	claim2 := &mockClaimWithValidFunc2{}
	cfg2 := jMW.Config{Issuer: "mock", Alg: jMW.HS256, Subject: "test", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"}
	token, err = jMW.New(cfg2, claim2)
	if test.NoError(err) {
		t := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoxNTQ5MDUzNzU2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.Ib1mT1jV9vX5JTtyX7Z1XvpoFsOCP4xj0xEquPEHGtA"
		_, err = token.Parse(t)
		test.Error(err)
		test.Equal(jMW.ErrTokenExpired.Error(), err.Error())
	}
}

func TestToken_MW(t *testing.T) {

	test := assert.New(t)

	// All tokens have the expired date 2050
	tests := []struct {
		token string
		error bool
	}{
		// ok
		{token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.zENdytJIyGJ9HoEdq4CfC4vYlOVWCLp7P8mlTj2XhC0"},
		// err: invalid toke, some characters were deleted
		{error: true, token: "eyJhbGcikpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJlbWFpbCI6ImpvaG5AZXhhbXBsZS5jb20iLCJwZXJtaXNzaW9uIjpbIi9kYXNoYm9hcmQiLCIvc2V0dGluZ3MiXSwiYXVkIjoiZ290ZXN0IiwiZXhwIjoyNTQ5MDUzODA2LCJqdGkiOiIxR2FwcWxZRTlHWXBHSmZNU1ByQ1V1dVBOMXciLCJpYXQiOjE1NDkwNTM3NTYsImlzcyI6Im1vY2siLCJuYmYiOjE1NDkwNTM3NTYsInN1YiI6InRlc3QifQ.zENdytJIyGJ9HoEdq4CfC4vYlOVWCLp7P8mlTj2XhC0"},
		// err: no cookie
		{error: true},
	}

	claim := &mockClaim{Name: "John Doe", Email: "john@example.com", Permission: []string{"/dashboard", "/settings"}}
	cfg := jMW.Config{Issuer: "mock", Alg: jMW.HS256, Subject: "test", Audience: "gotest", Duration: 10 * time.Second, SignKey: "secret"}
	token, err := jMW.New(cfg, claim)
	test.NoError(err)

	for _, tt := range tests {

		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("jwt:" + r.Context().Value(jMW.CLAIM).(*mockClaim).Name))
		}

		r, err := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
		test.NoError(err)

		if tt.token != "" {
			cookie := &http.Cookie{HttpOnly: true, Name: jMW.CookieJwt, Value: tt.token}
			r.AddCookie(cookie)
		}

		w := httptest.NewRecorder()
		mw := middleware.New(token.MW)
		mw.Handle(handlerFunc)(w, r)

		if tt.error {
			test.Equal("", w.Body.String())
			test.Equal(401, w.Code)
		} else {
			test.Equal("jwt:John Doe", w.Body.String())
			test.Equal(200, w.Code)
		}
	}
}
