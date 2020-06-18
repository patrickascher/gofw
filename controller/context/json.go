package context

import (
	"encoding/json"
)

func init() {
	_ = Register("json", NewJson)
}

// New satisfies the config.provider interface.
func NewJson() Interface {
	return &jsonStruct{}
}

// json struct
type jsonStruct struct {
}

// renderJson render the given data to json.
// It sets an content header and marshals the data.
func (js jsonStruct) Write(r *Response) error {
	r.Raw().Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(r.data)
	if err != nil {
		return err
	}
	_, err = r.Raw().Write(j)
	return err
}
