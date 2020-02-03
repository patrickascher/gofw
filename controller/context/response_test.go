// Copyright 2018 (pat@fullhouse-productions.com)
// TODO check license styles
package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"net/http"
)

//HTTP
type FakeResponse struct {
	headers http.Header
	body    []byte
	status  int
}

func (r *FakeResponse) Status() int {
	return r.status
}
func (r *FakeResponse) Body() string {
	return string(r.body)
}

func (r *FakeResponse) BodyRaw() []byte {
	return r.body
}

func (r *FakeResponse) Header() http.Header {

	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

func (r *FakeResponse) Write(body []byte) (int, error) {
	r.body = body
	return len(body), nil
}

func (r *FakeResponse) WriteHeader(status int) {
	r.status = status
}

// TestResponse_Data if the data is set to the variable
func TestResponse_Data(t *testing.T) {
	resp := FakeResponse{}
	r := NewResponse(&resp)

	r.AddData("User", "Goofy")

	assert.Equal(t, map[string]interface{}{"User": "Goofy"}, r.data)
}

// TestResponse_Data if the data is set to the variable
func TestResponse_Render(t *testing.T) {
	resp := FakeResponse{}
	r := NewResponse(&resp)
	headers := resp.Header()

	r.AddData("User", "Goofy")
	//render json
	r.Render("json")
	assert.Equal(t, "{\"User\":\"Goofy\"}", resp.Body())
	assert.Equal(t, []string{"application/json"}, headers["Content-Type"])

	//render default
	r.Render("html")
	assert.Equal(t, "{\"User\":\"Goofy\"}", resp.Body())

	//json marshal error
	r.AddData("Err", make(chan int))
	err := r.Render("json")
	assert.Error(t, err)
}
