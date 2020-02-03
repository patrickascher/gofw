package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

//Default Claim
type TestClaim struct {
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Permission []string `json:"permission"`
	jwt.StandardClaims
}

func (c *TestClaim) SetJid(id string) {
	c.Id = id
}
func (c *TestClaim) Jid() string {
	return c.Id
}
func (c *TestClaim) SetIss(iss string) {
	c.Issuer = iss
}
func (c *TestClaim) Iss() string {
	return c.Issuer
}
func (c *TestClaim) SetAud(aud string) {
	c.Audience = aud
}
func (c *TestClaim) Aud() string {
	return c.Audience
}
func (c *TestClaim) SetSub(sub string) {
	c.Subject = sub
}
func (c *TestClaim) Sub() string {
	return c.Subject
}
func (c *TestClaim) SetIat(iat int64) {
	c.IssuedAt = iat
}
func (c *TestClaim) Iat() int64 {
	return c.IssuedAt
}
func (c *TestClaim) SetExp(exp int64) {
	c.ExpiresAt = exp
}
func (c *TestClaim) Exp() int64 {
	return c.ExpiresAt
}
func (c *TestClaim) SetNbf(nbf int64) {
	c.NotBefore = nbf
}
func (c *TestClaim) Nbf() int64 {
	return c.NotBefore
}
func (c *TestClaim) Render() interface{} {
	return ""
}

func helperConfig() *Config {
	return &Config{
		Alg:      "HS256",
		Issuer:   "fullhouse-productions.com",
		Audience: "auth",
		Subject:  "wall-e",
		Duration: 10 * time.Second,
		SignKey:  "secret",
	}
}

func TestConfig(t *testing.T) {
	cfg := helperConfig()

	// true - everything ok
	assert.True(t, cfg.valid())

	// true - valid ALG type
	cfg = helperConfig()
	cfg.Alg = "HS256"
	assert.True(t, cfg.valid())
	cfg.Alg = "HS384"
	assert.True(t, cfg.valid())
	cfg.Alg = "HS512"
	assert.True(t, cfg.valid())

	// false - ALG empty
	cfg = helperConfig()
	cfg.Alg = ""
	assert.False(t, cfg.valid())

	// false - ALG wrong
	cfg = helperConfig()
	cfg.Alg = "XYZ"
	assert.False(t, cfg.valid())

	// false - ISS empty
	cfg = helperConfig()
	cfg.Issuer = ""
	assert.False(t, cfg.valid())

	// false - AUD empty
	cfg = helperConfig()
	cfg.Audience = ""
	assert.False(t, cfg.valid())

	// false - SUB empty
	cfg = helperConfig()
	cfg.Subject = ""
	assert.False(t, cfg.valid())

	// false - EXP empty
	cfg = helperConfig()
	cfg.Duration = time.Duration(0)
	assert.False(t, cfg.valid())

	// false - SignKey empty
	cfg = helperConfig()
	cfg.SignKey = ""
	assert.False(t, cfg.valid())
}

func TestToken_newClaim(t *testing.T) {
	token := Token{}
	token.claim = &TestClaim{}
	assert.Equal(t, token.claim, token.newClaim())
}
