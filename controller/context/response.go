package context

import (
	"encoding/json"
	"net/http"
)

// Response struct keeps the raw a ResponseWriter and the response data.
type Response struct {
	raw  http.ResponseWriter
	data map[string]interface{}
}

// NewResponse creates a new response with the raw http.ResponseWriter data.
func NewResponse(raw http.ResponseWriter) *Response {
	return &Response{raw: raw, data: make(map[string]interface{})}
}

// AddData is used to add response data
func (o *Response) AddData(key string, value interface{}) {
	o.data[key] = value
}

// Data is returning the value by key if exists.
func (o *Response) Data(key string) interface{} {
	if val, ok := o.data[key]; ok {
		return val
	}
	return nil
}

// Raw returns the *http.ResponseWriter
func (o *Response) Raw() http.ResponseWriter {
	return o.raw
}

// Render the response with the given renderType
func (o *Response) Render(renderType string) error {
	var err error

	switch renderType {
	case "json":
		err = o.renderJson()
	default:
		//TODO log type
		err = o.renderJson()
	}

	return err
}

//renderJson render the given data to json
func (o *Response) renderJson() error {

	o.Raw().Header().Set("Content-Type", "application/json")

	js, err := json.Marshal(o.data)

	if err != nil {
		return err
	}

	_, err = o.Raw().Write(js)
	return err
}
