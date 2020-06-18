// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"errors"
	"fmt"
	"net/http"
)

// registry for all cache providers.
var registry = make(map[string]provider)

// Error messages
var (
	ErrUnknownProvider       = "cache: unknown response-provider %q"
	ErrNoProvider            = errors.New("cache: empty cache-name or cache-provider is nil")
	ErrProviderAlreadyExists = "cache: cache-provider %#v is already registered"
)

type provider func() Interface
type Interface interface {
	Write(response *Response) error
}

// Response struct.
type Response struct {
	raw  http.ResponseWriter
	data map[string]interface{}
}

func init() {

}

// newResponse initialization the Response struct.
func newResponse(raw http.ResponseWriter) *Response {
	return &Response{raw: raw, data: make(map[string]interface{})}
}

// SetData by key and value.
func (o *Response) SetData(key string, value interface{}) {
	o.data[key] = value
}

// Data returned by the key.
// If the key does not exist, nil will return.
func (o *Response) Data(key string) interface{} {
	if val, ok := o.data[key]; ok {
		return val
	}
	return nil
}

// Raw returns the original *http.ResponseWriter
func (o *Response) Raw() http.ResponseWriter {
	return o.raw
}

// Render the response with the given renderType.
// Error will return if the render provider is not registered.
func (o *Response) Render(renderType string) error {
	instanceFn, ok := registry[renderType]
	if !ok {
		return fmt.Errorf(ErrUnknownProvider, renderType)
	}
	instance := instanceFn()
	return instance.Write(o)
}

// Register the cache provider. This should be called in the init() of the providers.
// If the cache provider/name is empty or is already registered, an error will return.
func Register(provider string, fn provider) error {
	if fn == nil || provider == "" {
		return ErrNoProvider
	}
	if _, exists := registry[provider]; exists {
		return fmt.Errorf(ErrProviderAlreadyExists, provider)
	}
	registry[provider] = fn
	return nil
}
