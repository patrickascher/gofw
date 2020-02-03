package controller_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/patrickascher/gofw/controller"
	"github.com/stretchr/testify/assert"
)

type FakeResponse struct {
	headers http.Header
	body    []byte
	status  int
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

// TestNewContext testing if a new Context is getting created
func TestNewContext(t *testing.T) {
	rw := &FakeResponse{}
	ctx := controller.NewContext(&http.Request{}, rw)

	assert.Equal(t, "controller.Context", reflect.TypeOf(ctx).Elem().String())
	assert.Equal(t, "context.Response", reflect.TypeOf(ctx.Response).Elem().String())
	assert.Equal(t, "context.Request", reflect.TypeOf(ctx.Request).Elem().String())
	assert.Equal(t, "controller_test.FakeResponse", reflect.TypeOf(ctx.Response.Raw()).Elem().String())
}
