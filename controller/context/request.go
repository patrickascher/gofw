// Package context creates a request and response context
package context

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/mssola/user_agent"
	"github.com/patrickascher/gofw/middleware/jwt"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"
)

//TODO Form Input
//TODO check CSRF

// ErrParameter error message
var ErrParameter = errors.New("the requested parameter %#v does not exist")

// Request struct
type Request struct {
	raw     *http.Request
	params  map[string][]string
	files   map[string][]*multipart.FileHeader
	pattern string
	ua      UserAgent
}

// NewRequest creates a new Request with the raw http.Request data.
func NewRequest(raw *http.Request) *Request {
	return &Request{raw: raw}
}

// Pattern returns the router pattern.
func (req *Request) Pattern() string {
	return req.pattern
}

// Token returns the jwt token.
func (req *Request) Token() interface{} {
	return req.Raw().Context().Value(jwt.ContextName)
}

// Method returns the HTTP Method in uppercase.
func (req *Request) Method() string {
	return strings.ToUpper(req.Raw().Method)
}

// Is checks the given method with the request Method. Both strings are getting set to uppercase.
func (req *Request) Is(m string) bool {
	if strings.ToUpper(m) == req.Method() {
		return true
	}
	return false
}

// IsSecure checks if the request is https.
func (req *Request) IsSecure() bool {
	return req.Scheme() == "https"
}

// IsPost checks if its a HTTP POST Method
func (req *Request) IsPost() bool {
	return req.Is("Post")
}

// IsGet checks if its a HTTP GET Method
func (req *Request) IsGet() bool {
	return req.Is("Get")
}

// IsPatch checks if its a HTTP PATCH Method
func (req *Request) IsPatch() bool {
	return req.Is("Patch")
}

// IsPut checks if its a HTTP PUT Method
func (req *Request) IsPut() bool {
	return req.Is("Put")
}

// IsDelete checks if its a HTTP DELETE Method
func (req *Request) IsDelete() bool {
	return req.Is("Delete")
}

// UserAgent parses the User-Agent header.
func (req *Request) UserAgent() *UserAgent {
	if req.ua == (UserAgent{}) {
		uaString := req.Raw().UserAgent()
		ua := user_agent.New(uaString)
		name, v := ua.Browser()
		req.ua = UserAgent{
			os:      OsInfo{Name: ua.OSInfo().Name, Version: ua.OSInfo().Version},
			browser: BrowserInfo{Name: name, Version: v},
			mobile:  ua.Mobile()}
	}

	return &req.ua
}

// UserAgent struct
type UserAgent struct {
	os      OsInfo
	browser BrowserInfo
	mobile  bool
}

// OS returns the operation system information
func (us *UserAgent) OS() OsInfo {
	return us.os
}

// Mobile returns a boolean
func (us *UserAgent) Mobile() bool {
	return us.mobile
}

// OsInfo provides the name and version of the users operating system.
type OsInfo struct {
	Name    string
	Version string
}

// Browser returns the browser information
func (us *UserAgent) Browser() BrowserInfo {
	return us.browser
}

// BrowserInfo provides the name and version of the users browser.
type BrowserInfo struct {
	Name    string
	Version string
}

// Raw returns the original *http.Request
func (req *Request) Raw() *http.Request {
	return req.raw
}

// AddJulienSchmidtRouterParams is used to add the router specific params.
// It also sets the router pattern.
func (req *Request) AddJulienSchmidtRouterParams(p httprouter.Params) error {
	err := req.parse()
	if err != nil {
		return err
	}

	var wildcard bool
	req.pattern = req.URL()

	// This is needed if a wildcard is used and no params are given
	if len(p) == 1 && p[0].Value == "/" {
		req.pattern = req.pattern + "*" + p[0].Key
		return nil
	}

	//TODO same logic in rbac.go - fix it
	if len(p) > 0 {
		for _, v := range p {
			req.params[v.Key] = []string{v.Value}
			req.pattern = strings.Replace(req.pattern, v.Value, ":"+v.Key, 1)
		}

		// check if the param is a wildcard.
		// if there is no slash between the key param, then it is a wildcard
		// its only working with rules like /roles/*grid = /roles/param1/param2... not with /roles/?param1=xxx
		if !strings.Contains(req.pattern, "/:"+p[len(p)-1].Key) {
			req.pattern = strings.Replace(req.pattern, ":"+p[len(p)-1].Key, "/*"+p[len(p)-1].Key, 1)
			wildcard = true
		}

		if wildcard {
			params := req.params[p[len(p)-1].Key][0]
			params = params[1:]
			param := strings.Split(params, "/")

			if len(params) > 0 { // ? somehow the for loop is been called also if the length is 0 and i is not smaller in that case?
				for i := 0; i < len(param); i++ {
					req.params[param[i]] = []string{param[i+1]}
					i++
				}
			}
			// delete old params entry
			delete(req.params, p[len(p)-1].Key)
		}
	}

	return nil
}

// parse is adding all GET params and POST Form data in req.params
// its only called once if the method "Param" or "Params" is called
func (req *Request) parse() error {

	if req.params == nil {
		req.params = make(map[string][]string)
		req.files = make(map[string][]*multipart.FileHeader)
	} else {
		//already parsed
		return nil
	}

	// TODO set body limit

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

// File returns the requested file of a multipart post.
// If it was not found, an error will return.
// It returns a []*FileHeader because the underlying input field could be an array.
func (req *Request) File(k string) ([]*multipart.FileHeader, error) {
	err := req.parse()
	if err != nil {
		return []*multipart.FileHeader{}, err
	}

	if val, ok := req.files[k]; ok {
		return val, nil
	}
	return []*multipart.FileHeader{}, fmt.Errorf(ErrParameter.Error(), k)
}

// Files returns all existing files.
// It will return an error if there was something wrong with the parsing.
// It returns a map[string][]*FileHeader because the underlying input field could be an array.
func (req *Request) Files() (map[string][]*multipart.FileHeader, error) {
	err := req.parse()
	if err != nil {
		return map[string][]*multipart.FileHeader{}, err
	}
	return req.files, nil
}

// Param returns the requested parameter.
// If it was not found, an error will return.
// It returns a []string because the underlying HTML input field could be an array.
func (req *Request) Param(k string) ([]string, error) {
	err := req.parse()
	if err != nil {
		return []string{}, err
	}

	if val, ok := req.params[k]; ok {
		return val, nil
	}
	return []string{}, fmt.Errorf(ErrParameter.Error(), k)
}

// Params returns all existing parameters.
// It will return an error if there was something wrong with the parsing.
// It returns a map[string][]string because the underlying HTML input field could be an array.
func (req *Request) Params() (map[string][]string, error) {
	err := req.parse()
	if err != nil {
		return map[string][]string{}, err
	}
	return req.params, nil
}

// IP tries to find the IP Address in th X-Forwarded-For Header. If it exists, it takes the first entry.
// If it does not exist, it returns the RemoteAddr without Port.
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

// Scheme checks the `X-Forwarded-Proto` header. If that one is empty the URL.Scheme gets checked.
// If that is also empty the request TLS gets checked.
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

// Host eturns the host name. If no host info is available, localhost will return. Port number will be removed if existing.
// If no host info exists, it will return localhost.
func (req *Request) Host() string {
	if req.Raw().Host != "" {
		if hostPart, _, err := net.SplitHostPort(req.Raw().Host); err == nil {
			return hostPart
		}
		return req.Raw().Host
	}
	return "localhost"
}

// Protocol returns request protocol name, such as HTTP/1.1 .
func (req *Request) Protocol() string {
	return req.Raw().Proto
}

// URI returns full request url with query string, fragment.
func (req *Request) URI() string {
	return req.Raw().RequestURI
}

// URL returns request url path (without query string, fragment).
func (req *Request) URL() string {
	return req.Raw().URL.Path
}

// FullURL returns the schema,host,port,uri
func (req *Request) FullURL() string {
	s := req.Site()
	if req.Port() != 80 {
		s = fmt.Sprintf("%v:%v%v", s, req.Port(), req.URI())
	}
	return s
}

// Site returns base site url as scheme://domain type without the port.
func (req *Request) Site() string {
	return req.Scheme() + "://" + req.Domain()
}

// Domain returns host name.
// Alias of Host method.
func (req *Request) Domain() string {
	return req.Host()
}

// Port returns request client port.
// when error or empty, return 80.
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
