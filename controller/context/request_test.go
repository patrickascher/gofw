// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context_test

import (
	context2 "context"
	"crypto/tls"
	"fmt"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func requestHelperHTTPS() (http.Request, *context.Context) {
	url, _ := url.Parse("https://example.com:8080/user?id=12#test")

	header := http.Header{}
	header["Referer"] = []string{"GoTest"}
	header["X-Forwarded-For"] = []string{"192.168.2.1"}
	header["User-Agent"] = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}
	header["X-Forwarded-Proto"] = []string{"https"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/2",
		Header:     header,
		Method:     "GET",
		Host:       "example.com:8080",
		URL:        url,
		RequestURI: "/user?id=12#test"}
	rw := &FakeResponse{}

	ctx := context.New(&r, rw)
	return r, ctx
}

func requestHelperHTTP() (http.Request, *context.Context) {
	url, _ := url.Parse("http://example.com:8080/user?id=12#test")

	header := http.Header{}
	header["REMOTE_ADDR"] = []string{"192.168.2.2"}
	header["User-Agent"] = []string{"Mozilla/5.0 Mobile (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}
	header["X-Forwarded-Proto"] = []string{"http"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/1.1",
		Header:     header,
		Method:     "GET",
		Host:       "example.com",
		URL:        url,
		RequestURI: "/user?id=12#test"}
	rw := &FakeResponse{}

	ctx := context.New(&r, rw)
	return r, ctx
}

func requestHelperIP() (http.Request, *context.Context) {
	url, _ := url.Parse("http://exple.com:8043/user?id=12#test")

	header := http.Header{}
	header["Referer"] = []string{"TestSuit"}
	header["Remote Address"] = []string{"192.168.2.3"}
	header["User-Agent"] = []string{"Mozilla/5.0 Mobile (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/1.1",
		Header:     header,
		Method:     "POST",
		Host:       "example.com:8080",
		URL:        url,
		RequestURI: "/user?id=12#test"}

	r.RemoteAddr = "192.168.2.3"

	rw := &FakeResponse{}

	ctx := context.New(&r, rw)
	return r, ctx
}

func requestHelperIPPort() (http.Request, *context.Context) {
	url, _ := url.Parse("http://exple.com:8043/user?id=12#test")

	header := http.Header{}
	header["Referer"] = []string{"TestSuit"}
	header["Remote Address"] = []string{"192.168.2.3"}
	header["User-Agent"] = []string{"Mozilla/5.0 Mobile (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/1.1",
		Header:     header,
		Method:     "POST",
		Host:       "example.com",
		URL:        url,
		RequestURI: "/user?id=12#test"}

	r.RemoteAddr = "192.168.2.4:8080"

	rw := &FakeResponse{}

	ctx := context.New(&r, rw)
	return r, ctx
}

func TestRequest_Method(t *testing.T) {
	test := assert.New(t)
	_, ctxHttps := requestHelperHTTPS()
	_, ctxHttp := requestHelperHTTP()

	// Method
	test.Equal("GET", ctxHttps.Request.Method())

	// Is
	test.True(ctxHttps.Request.Is("get"))
	test.True(ctxHttps.Request.Is("GeT"))
	test.True(ctxHttps.Request.Is("GET"))
	test.False(ctxHttps.Request.Is("POST"))

	// Is...
	test.True(ctxHttps.Request.IsGet())
	test.False(ctxHttps.Request.IsPost())
	test.False(ctxHttps.Request.IsPatch())
	test.False(ctxHttps.Request.IsPut())
	test.False(ctxHttps.Request.IsDelete())

	// IsSecure
	test.False(ctxHttp.Request.IsSecure())
	test.True(ctxHttps.Request.IsSecure())

}

func TestRequest_UserAgent(t *testing.T) {
	_, ctxHttps := requestHelperHTTPS()
	_, ctxHttp := requestHelperHTTP()

	agent := ctxHttps.Request.UserAgent()
	agentMobile := ctxHttp.Request.UserAgent()

	// os
	assert.Equal(t, "Mac OS X", agent.OS().Name)
	assert.Equal(t, "10.13.6", agent.OS().Version)

	// Browser
	assert.Equal(t, "Chrome", agent.Browser().Name)
	assert.Equal(t, "69.0.3497.100", agent.Browser().Version)

	// Mobile
	assert.False(t, agent.Mobile())
	assert.True(t, agentMobile.Mobile())
}

func TestRequest_Raw(t *testing.T) {
	r, ctx := requestHelperHTTP()
	assert.Equal(t, &r, ctx.Request.Raw())
}

func TestRequest_IP(t *testing.T) {
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, "192.168.2.1", ctxHttps.Request.IP()) // X-Forwarded first IP.

	//_, ctxHttp := requestHelperHTTP()
	//assert.Equal(t, "192.168.2.2", ctxHttp.Request.IP()) // REMOTE_ADDR by Header.

	_, ctxIp := requestHelperIP()
	assert.Equal(t, "192.168.2.3", ctxIp.Request.IP()) // Remote Addr set manually.

	_, ctxIpPort := requestHelperIPPort()
	assert.Equal(t, "192.168.2.4", ctxIpPort.Request.IP()) // Remote Addr set manually.
}

func TestRequest_Proxy(t *testing.T) {
	// existing X-Forwarted-For header
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, []string{"192.168.2.1"}, ctxHttps.Request.Proxy())

	// header does not exist
	_, ctxHttp := requestHelperHTTP()
	assert.Equal(t, []string{}, ctxHttp.Request.Proxy())
}

func TestRequest_Scheme(t *testing.T) {

	// Header X-Forwarded-Proto
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, "https", ctxHttps.Request.Scheme())

	// Scheme URL.Scheme
	_, ctxIP := requestHelperIP()
	assert.Equal(t, "http", ctxIP.Request.Scheme())

	//TLS
	url, _ := url.Parse("https://test.com:8043/user?q=dotnet")
	url.Scheme = ""
	tls_ := tls.ConnectionState{}
	r := http.Request{TLS: &tls_, URL: url}
	rw := &FakeResponse{}
	req := context.New(&r, rw)
	assert.Equal(t, "https", req.Request.Scheme())

	//no TLS
	url, _ = url.Parse("https://test.com:8043/user?q=dotnet")
	url.Scheme = ""
	tls_ = tls.ConnectionState{}
	r = http.Request{TLS: nil, URL: url}
	rw = &FakeResponse{}
	req = context.New(&r, rw)
	assert.Equal(t, "http", req.Request.Scheme())
}

func TestRequest_Host(t *testing.T) {
	r := http.Request{}
	rw := &FakeResponse{}

	req := context.New(&r, rw)
	assert.Equal(t, "localhost", req.Request.Host())

	r = http.Request{Host: "example.com"}
	req = context.New(&r, rw)
	assert.Equal(t, "example.com", req.Request.Host())

	r = http.Request{Host: "example.com:3000"}
	req = context.New(&r, rw)
	assert.Equal(t, "example.com", req.Request.Host())
}

func TestRequest_Protocol(t *testing.T) {
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, "HTTP/2", ctxHttps.Request.Protocol())

	_, ctxHttp := requestHelperHTTP()
	assert.Equal(t, "HTTP/1.1", ctxHttp.Request.Protocol())
}

func TestRequest_URI(t *testing.T) {
	_, ctx := requestHelperHTTPS()
	assert.Equal(t, "/user?id=12#test", ctx.Request.URI())
}

func TestRequest_URL(t *testing.T) {
	_, ctx := requestHelperHTTPS()
	assert.Equal(t, "/user", ctx.Request.URL())
}

func TestRequest_FullURL(t *testing.T) {
	_, ctx := requestHelperHTTPS()
	assert.Equal(t, "https://example.com:8080/user?id=12#test", ctx.Request.FullURL())
}

func TestRequest_Site(t *testing.T) {
	_, ctx := requestHelperHTTPS()
	assert.Equal(t, "https://example.com", ctx.Request.Site())
}

func TestRequest_Domain(t *testing.T) {
	_, ctx := requestHelperHTTPS()
	assert.Equal(t, "example.com", ctx.Request.Domain())
}

func TestRequest_Port(t *testing.T) {
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, 8080, ctxHttps.Request.Port())

	_, ctxHttp := requestHelperHTTP()
	assert.Equal(t, 80, ctxHttp.Request.Port())
}

func TestRequest_Referer(t *testing.T) {
	_, ctxHttps := requestHelperHTTPS()
	assert.Equal(t, "GoTest", ctxHttps.Request.Referer())

	_, ctxHttp := requestHelperHTTP()
	assert.Equal(t, "", ctxHttp.Request.Referer())
}

func TestRequest_Token(t *testing.T) {
	r := http.Request{Host: "example.com:3000"}
	rw := &FakeResponse{}

	req := context.New(r.WithContext(context2.WithValue(r.Context(), jwt.CLAIM, "abcd")), rw)
	assert.Equal(t, "abcd", req.Request.Token())
}

func TestRequest_Pattern(t *testing.T) {
	r := http.Request{Host: "example.com:3000"}
	rw := &FakeResponse{}

	req := context.New(&r, rw)
	assert.Equal(t, "", req.Request.Pattern())

	req = context.New(r.WithContext(context2.WithValue(r.Context(), "pattern", "/user/:id")), rw)
	assert.Equal(t, "/user/:id", req.Request.Pattern())
}

func TestRequest_parseForm(t *testing.T) {
	rw := &FakeResponse{}

	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	form := url.Values{}
	form.Add("username", "Mike")

	r := httptest.NewRequest("POST", "https://test.com:8043/user?q=dotnet#test", strings.NewReader(form.Encode()))
	r.Header = header

	req := context.New(r, rw)

	params, err := req.Request.Params()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"username": {"Mike"}}, params)

	param, err := req.Request.Param("username")
	assert.NoError(t, err)
	assert.Equal(t, []string{"Mike"}, param)

	param, err = req.Request.Param("password")
	assert.Error(t, err)
	assert.Nil(t, param)
}

func TestRequest_parseGet(t *testing.T) {
	rw := &FakeResponse{}

	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	form := url.Values{}
	form.Add("username", "Mike")

	r := httptest.NewRequest("GET", "https://test.com:8043/user?id=12#test", strings.NewReader(form.Encode()))
	r.Header = header

	req := context.New(r, rw)

	params, err := req.Request.Params()
	//recall render should not be called twice
	_, _ = req.Request.Params()

	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"id": {"12#test"}}, params)

	// id
	param, err := req.Request.Param("id")
	assert.NoError(t, err)
	assert.Equal(t, []string{"12#test"}, param)

	//parameter does not exist
	_, err2 := req.Request.Param("password")
	assert.Error(t, err2)
	assert.Equal(t, err2, fmt.Errorf(context.ErrParameter.Error(), "password"))
}
