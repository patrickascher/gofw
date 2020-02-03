// Copyright 2018 (pat@fullhouse-productions.com)
// TODO check license styles
package context_test

import (
	"net/http"
	"net/url"
	"testing"

	"crypto/tls"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/controller/context"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"strings"
)

func requestHelper() (http.Request, *context.Request) {
	url, _ := url.Parse("https://test.com:8043/user?q=dotnet#test")

	header := http.Header{}
	header["Referer"] = []string{"TestSuit"}
	header["X-Forwarded-For"] = []string{"192.168.2.1"}
	header["User-Agent"] = []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}
	header["X-Forwarded-Proto"] = []string{"https"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/2",
		Header:     header,
		Method:     "GET",
		Host:       "test.com:8043",
		URL:        url,
		RequestURI: "/user?q=dotnet#test"}

	req := context.NewRequest(&r)
	return r, req
}

func requestHelperHTTP() (http.Request, *context.Request) {
	url, _ := url.Parse("http://test.com/user?q=dotnet#test")

	header := http.Header{}
	header["Remote Address"] = []string{"192.168.2.1"}
	header["User-Agent"] = []string{"Mozilla/5.0 Mobile (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/1.1",
		Header:     header,
		Method:     "POST",
		Host:       "test.com",
		URL:        url,
		RequestURI: "/user?q=dotnet"}

	r.RemoteAddr = "192.168.2.1:3000"

	req := context.NewRequest(&r)
	return r, req
}

func requestHelperIP() (http.Request, *context.Request) {
	url, _ := url.Parse("http://test.com:8043/user?q=dotnet#test")

	header := http.Header{}
	header["Referer"] = []string{"TestSuit"}
	header["Remote Address"] = []string{"192.168.2.1"}
	header["User-Agent"] = []string{"Mozilla/5.0 Mobile (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"}

	//Create a new request
	r := http.Request{
		Proto:      "HTTP/1.1",
		Header:     header,
		Method:     "POST",
		Host:       "test.com:8043",
		URL:        url,
		RequestURI: "/user?q=dotnet"}

	r.RemoteAddr = "192.168.2.1"

	req := context.NewRequest(&r)
	return r, req
}

func TestRequest_Method(t *testing.T) {
	_, req := requestHelper()
	_, req2 := requestHelperHTTP()

	// Method
	assert.Equal(t, "GET", req.Method())
	assert.Equal(t, "POST", req2.Method())

	// Is
	assert.True(t, req.Is("get"))
	assert.True(t, req.Is("GeT"))
	assert.True(t, req.Is("GET"))
	assert.False(t, req.Is("POST"))
	assert.True(t, req2.Is("POST"))
	assert.False(t, req2.Is("GET"))

	// Is...
	assert.True(t, req.IsGet())
	assert.False(t, req.IsPost())
	assert.False(t, req.IsPatch())
	assert.False(t, req.IsPut())
	assert.False(t, req.IsDelete())
	assert.True(t, req2.IsPost())
	assert.False(t, req2.IsGet())
}

func TestRequest_IsSecure(t *testing.T) {
	_, req := requestHelper()
	assert.True(t, req.IsSecure())

	_, req = requestHelperHTTP()
	assert.False(t, req.IsSecure())
}

func TestRequest_UserAgent(t *testing.T) {
	_, req := requestHelper()
	agent := req.UserAgent()

	assert.Equal(t, context.OsInfo(context.OsInfo{Name: "Mac OS X", Version: "10.13.6"}), agent.OS())
	assert.Equal(t, context.BrowserInfo(context.BrowserInfo{Name: "Chrome", Version: "69.0.3497.100"}), agent.Browser())
	assert.False(t, agent.Mobile())

	_, req2 := requestHelperHTTP()
	agent = req2.UserAgent()
	assert.True(t, agent.Mobile())
}

func TestRequest_Raw(t *testing.T) {
	r, req := requestHelper()
	assert.Equal(t, &r, req.Raw())
}

func TestRequest_IP(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "192.168.2.1", req.IP())

	_, req2 := requestHelperHTTP()
	assert.Equal(t, "192.168.2.1", req2.IP())

	_, req3 := requestHelperIP()
	assert.Equal(t, "192.168.2.1", req3.IP())
}

func TestRequest_Proxy(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, []string{"192.168.2.1"}, req.Proxy())
}

func TestRequest_Scheme(t *testing.T) {
	//X-Forwarded-Proto
	header := http.Header{}
	header["X-Forwarded-Proto"] = []string{"httpsx"}
	r := http.Request{Header: header}
	req := context.NewRequest(&r)
	assert.Equal(t, "httpsx", req.Scheme())

	//URL Scheme
	url, _ := url.Parse("https://test.com:8043/user?q=dotnet")
	url.Scheme = ""
	r = http.Request{URL: url}
	r.URL.Scheme = "httpsx"
	req = context.NewRequest(&r)
	assert.Equal(t, "httpsx", req.Scheme())

	//TLS
	url, _ = url.Parse("https://test.com:8043/user?q=dotnet")
	url.Scheme = ""
	tls_ := tls.ConnectionState{}
	r = http.Request{TLS: &tls_, URL: url}
	req = context.NewRequest(&r)
	assert.Equal(t, "https", req.Scheme())

	// else
	url, _ = url.Parse("https://test.com:8043/user?q=dotnet")
	url.Scheme = ""
	header = http.Header{}
	r = http.Request{Header: header, URL: url}
	req = context.NewRequest(&r)
	assert.Equal(t, "http", req.Scheme())
}

func TestRequest_Host(t *testing.T) {
	r := http.Request{}
	req := context.NewRequest(&r)
	assert.Equal(t, "localhost", req.Host())

	r = http.Request{Host: "test.com"}
	req = context.NewRequest(&r)
	assert.Equal(t, "test.com", req.Host())

	r = http.Request{Host: "test2.com:3000"}
	req = context.NewRequest(&r)
	assert.Equal(t, "test2.com", req.Host())
}

func TestRequest_Protocol(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "HTTP/2", req.Protocol())

	_, req2 := requestHelperHTTP()
	assert.Equal(t, "HTTP/1.1", req2.Protocol())
}

func TestRequest_URI(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "/user?q=dotnet#test", req.URI())
}

func TestRequest_URL(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "/user", req.URL())
}

func TestRequest_FullURL(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "https://test.com:8043/user?q=dotnet#test", req.FullURL())
}

func TestRequest_Site(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "https://test.com", req.Site())
}

func TestRequest_Domain(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "test.com", req.Domain())
}

func TestRequest_Port(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, 8043, req.Port())

	_, req2 := requestHelperHTTP()
	assert.Equal(t, 80, req2.Port())
}

func TestRequest_Referer(t *testing.T) {
	_, req := requestHelper()
	assert.Equal(t, "TestSuit", req.Referer())

	_, req2 := requestHelperHTTP()
	assert.Equal(t, "", req2.Referer())
}

func TestRequest_parseForm(t *testing.T) {
	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	form := url.Values{}
	form.Add("username", "Mike")

	r := httptest.NewRequest("POST", "https://test.com:8043/user?q=dotnet#test", strings.NewReader(form.Encode()))
	r.Header = header

	req := context.NewRequest(r)

	params, err := req.Params()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"username": {"Mike"}}, params)
}

func TestRequest_parseGet(t *testing.T) {
	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	form := url.Values{}
	form.Add("username", "Mike")

	r := httptest.NewRequest("GET", "https://test.com:8043/user?q=dotnet#test", strings.NewReader(form.Encode()))
	r.Header = header

	req := context.NewRequest(r)

	params, err := req.Params()
	//recall render should not be called twice
	_, _ = req.Params()

	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"q": {"dotnet#test"}}, params)

	param, err := req.Param("q")
	assert.NoError(t, err)
	assert.Equal(t, []string{"dotnet#test"}, param)

	//parameter does not exist
	_, err2 := req.Param("userXXX")
	assert.Error(t, err2)
	assert.Equal(t, err2, fmt.Errorf(context.ErrParameter.Error(), "userXXX"))
}

func TestRequest_AddJulienSchmidtRouterParams(t *testing.T) {
	url, _ := url.Parse("http://test.com:8043/delete/root")
	r := http.Request{URL: url}
	req := context.NewRequest(&r)

	// check URL
	assert.Equal(t, "/delete/root", req.URL())

	params := httprouter.Params{}
	params = append(params, httprouter.Param{Key: "user", Value: "root"})
	req.AddJulienSchmidtRouterParams(params)

	// check param and router pattern
	param, err := req.Param("user")
	assert.NoError(t, err)
	assert.Equal(t, []string{"root"}, param)
	assert.Equal(t, "/delete/:user", req.Pattern())

	//parameter does not exist
	_, err2 := req.Param("userXXX")
	assert.Error(t, err2)
	assert.Equal(t, err2, fmt.Errorf(context.ErrParameter.Error(), "userXXX"))
}

// TODO when we have a working example again
/*func TestRequest_AddJulienSchmidtRouterWildcard(t *testing.T) {
	url, _ := url.Parse("http://test.com:8043/delete/edit/user/1")
	r := http.Request{URL: url}
	req := context.NewRequest(&r)

	// check URL
	assert.Equal(t, "/delete/edit/user/1", req.URL())

	params := httprouter.Params{}
	params = append(params, httprouter.Param{"1", "edit"})
	params = append(params, httprouter.Param{"2", "user"})
	params = append(params, httprouter.Param{"3", "1"})

	req.AddJulienSchmidtRouterParams(params)

	// check param and router pattern
	param, err := req.Param("user")
	assert.NoError(t, err)
	assert.Equal(t, []string{"root"}, param)
	assert.Equal(t, "/delete/:user", req.Pattern())

	//parameter does not exist
	param, err = req.Param("userXXX")
	assert.Error(t, err)
	assert.Equal(t, err, fmt.Errorf(context.ErrParameter.Error(),"userXXX"))
}*/
