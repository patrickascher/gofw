// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/mssola/user_agent"
	"github.com/patrickascher/gofw/middleware/jwt"
)

// ErrParameter error message
var ErrParameter = errors.New("controller/request: the parameter %#v does not exist")

// Request struct.
type Request struct {
	raw    *http.Request
	params map[string][]string
	files  map[string][]*multipart.FileHeader
	ua     UserAgent
}

// newRequest initialization the Request struct.
func newRequest(raw *http.Request) *Request {
	return &Request{raw: raw}
}

// Pattern returns the router url pattern.
// The pattern must be defined in the request context by the key router.PATTERN. This is usually done by the router.
//		Example: http://example.com/user/1
// 		/user/:id
func (req *Request) Pattern() string {
	//string is used instead of the router.PATTERN because of import cycle.
	if p := req.Raw().Context().Value("pattern"); p != nil {
		return p.(string)
	}
	return ""
}

// Raw returns the original *http.Request
func (req *Request) Raw() *http.Request {
	return req.raw
}

// Token returns the jwt token
// TODO check if needed?
func (req *Request) Token() interface{} {
	return req.Raw().Context().Value(jwt.CLAIM)
}

// Method returns the HTTP Method in uppercase.
//		GET
func (req *Request) Method() string {
	return strings.ToUpper(req.Raw().Method)
}

// Is checks the given method with the request HTTP Method. Both strings are getting set to uppercase.
func (req *Request) Is(m string) bool {
	if strings.ToUpper(m) == req.Method() {
		return true
	}
	return false
}

// IsSecure checks if the request is HTTPS.
func (req *Request) IsSecure() bool {
	return req.Scheme() == "https"
}

// IsPost checks if its a HTTP POST Method.
func (req *Request) IsPost() bool {
	return req.Is("Post")
}

// IsGet checks if its a HTTP GET Method.
func (req *Request) IsGet() bool {
	return req.Is("Get")
}

// IsPatch checks if its a HTTP PATCH Method.
func (req *Request) IsPatch() bool {
	return req.Is("Patch")
}

// IsPut checks if its a HTTP PUT Method.
func (req *Request) IsPut() bool {
	return req.Is("Put")
}

// IsDelete checks if its a HTTP DELETE Method.
func (req *Request) IsDelete() bool {
	return req.Is("Delete")
}

// File returns the requested file of a multipart POST.
// It returns a []*FileHeader because the underlying input field could be an array.
// Error will return on parse error or if the key does not exist.
func (req *Request) File(k string) ([]*multipart.FileHeader, error) {
	err := req.parse()
	if err != nil {
		return nil, err
	}

	if val, ok := req.files[k]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrParameter.Error(), k)
}

// Files returns all existing files.
// It returns a map[string][]*FileHeader because the underlying input field could be an array.
// Error will return on parse error.
func (req *Request) Files() (map[string][]*multipart.FileHeader, error) {
	err := req.parse()
	if err != nil {
		return map[string][]*multipart.FileHeader{}, err
	}
	return req.files, nil
}

// Param returns the requested parameter.
// It returns a []string because the underlying HTML input field could be an array.
// Error will return on parse error or if the key does not exist.
func (req *Request) Param(k string) ([]string, error) {
	err := req.parse()
	if err != nil {
		return nil, err
	}

	if val, ok := req.params[k]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrParameter.Error(), k)
}

// Params returns all existing parameters.
// It returns a map[string][]string because the underlying HTML input field could be an array.
// Error will return on parse error.
func (req *Request) Params() (map[string][]string, error) {
	err := req.parse()
	if err != nil {
		return nil, err
	}
	return req.params, nil
}

// IP of the request.
// First it checks the proxy X-Forwarded-For Header and takes the first entry - if exists.
// Otherwise the RemoteAddr will be returned without the Port.
// TODO Header RemoteAddr (is it official?) or X-Real-Ip
func (req *Request) IP() string {
	ips := req.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		rip, _, err := net.SplitHostPort(ips[0])
		if err != nil {
			rip = ips[0]
		}
		return rip
	}
	if ip, _, err := net.SplitHostPort(req.Raw().RemoteAddr); err == nil {
		return ip
	}
	return req.Raw().RemoteAddr
}

// Proxy return all IPs which are in the X-Forwarded-For header.
func (req *Request) Proxy() []string {
	if ips := req.Raw().Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

// Scheme (http/https) checks the `X-Forwarded-Proto` header. If that one is empty the URL.Scheme gets checked.
// If that is also empty the request TLS will be checked.
func (req *Request) Scheme() string {
	if scheme := req.Raw().Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if req.Raw().URL.Scheme != "" {
		return req.Raw().URL.Scheme
	}
	if req.Raw().TLS == nil {
		return "http"
	}
	return "https"
}

// Host returns the host name.
// Port number will be removed if existing.
// If no host info is available, localhost will return.
//		Example: https://example.com:8080/user?id=12#test
//		example.com
func (req *Request) Host() string {
	if req.Raw().Host != "" {
		if hostPart, _, err := net.SplitHostPort(req.Raw().Host); err == nil {
			return hostPart
		}
		return req.Raw().Host
	}
	return "localhost"
}

// Protocol returns the protocol name, such as HTTP/1.1 .
func (req *Request) Protocol() string {
	return req.Raw().Proto
}

// URI returns full request url with query string, fragment.
//		Example: https://example.com:8080/user?id=12#test
//		/user?id=12#test
func (req *Request) URI() string {
	return req.Raw().RequestURI
}

// URL returns request url path without the  query string and fragment.
//		Example: https://example.com:8080/user?id=12#test
//		/user
func (req *Request) URL() string {
	return req.Raw().URL.Path
}

// FullURL returns the schema,host,port,uri
//		Example: https://example.com:8080/user?id=12#test
//		https://example.com:8080/user?id=12#test
func (req *Request) FullURL() string {
	s := req.Site()
	if req.Port() != 80 {
		s = fmt.Sprintf("%v:%v%v", s, req.Port(), req.URI())
	}
	return s
}

// Site returns base site url as scheme://domain type without the port.
//		Example: https://example.com:8080/user?id=12#test
//		https://example.com
func (req *Request) Site() string {
	return req.Scheme() + "://" + req.Domain()
}

// Domain is an alias of Host method.
//		Example: https://example.com:8080/user?id=12#test
//		example.com
func (req *Request) Domain() string {
	return req.Host()
}

// Port returns request client port.
// On error or if its empty the standard port 80 will return.
func (req *Request) Port() int {
	if _, portPart, err := net.SplitHostPort(req.Raw().Host); err == nil {
		port, _ := strconv.Atoi(portPart)
		return port
	}
	return 80
}

// Referer returns the Referer Header
func (req *Request) Referer() string {
	return req.Raw().Referer()
}

// UserAgent parses the User-Agent header and return a UserAgent struct.
func (req *Request) UserAgent() *UserAgent {
	if req.ua == (UserAgent{}) {
		uaString := req.Raw().UserAgent()
		ua := user_agent.New(uaString)
		name, v := ua.Browser()
		req.ua = UserAgent{
			os:      osInfo{Name: ua.OSInfo().Name, Version: ua.OSInfo().Version},
			browser: browserInfo{Name: name, Version: v},
			mobile:  ua.Mobile()}
	}

	return &req.ua
}

// UserAgent struct.
type UserAgent struct {
	os      osInfo
	browser browserInfo
	mobile  bool
}

// OS returns the operation system information which contain the Name and Version.
func (us *UserAgent) OS() osInfo {
	return us.os
}

// Mobile returns a boolean
func (us *UserAgent) Mobile() bool {
	return us.mobile
}

// OsInfo provides the name and version of the users operating system.
type osInfo struct {
	Name    string
	Version string
}

// Browser returns the browser information which contain the Name and Version.
func (us *UserAgent) Browser() browserInfo {
	return us.browser
}

// BrowserInfo provides the name and version of the users browser.
type browserInfo struct {
	Name    string
	Version string
}

// parse is adding all GET params and POST Form data in req.params
// its only called once if the method "Param" or "Params" is called
// TODO how to handle the url params? same logic?
// TODO set body limit.
// TODO set filesize limit.
func (req *Request) parse() error {

	if req.params == nil {
		req.params = make(map[string][]string)
		req.files = make(map[string][]*multipart.FileHeader)
	} else {
		//already parsed
		return nil
	}

	// adding router params
	if req.Raw().Context().Value("params") != nil { // could not use ROUTER.PARAMS because of import cycle.
		req.params = req.Raw().Context().Value("params").(map[string][]string)
	}

	// Handling GET Params
	if req.IsGet() || req.IsDelete() {
		getParams := req.Raw().URL.Query()
		for param, val := range getParams {
			req.params[param] = val
		}
	}

	// Handling Form Post Params
	if req.IsPost() || req.IsPut() || req.IsPatch() {
		if strings.HasPrefix(req.Raw().Header.Get("Content-Type"), "multipart/form-data") {
			if err := req.Raw().ParseMultipartForm(16 * 1024 * 1024); err != nil { //TODO make this customizeable 16MB
				return err
			}
			for file, val := range req.Raw().MultipartForm.File {
				req.files[file] = val
			}
			for param, val := range req.Raw().MultipartForm.Value {
				req.params[param] = val
			}
		} else {
			if err := req.Raw().ParseForm(); err != nil {
				return err
			}
			getParams := req.Raw().PostForm
			for param, val := range getParams {
				req.params[param] = val
			}
		}

	}

	return nil
}
