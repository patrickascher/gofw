package callback

import (
	"errors"
	"reflect"
)

var errPtr = errors.New("callback: object must be a ptr")
var errMismatch = errors.New("callback: arguments or return value mismatch")

func StructMethod(method string, obj interface{}, args ...interface{}) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errPtr
	}

	m := reflect.ValueOf(obj).MethodByName(method)
	if m.IsValid() {
		if m.Type().NumIn() != len(args) || m.Type().NumOut() != 1 {
			return errMismatch
		}

		var arguments []reflect.Value
		for _, arg := range args {
			arguments = append(arguments, reflect.ValueOf(arg))
		}

		err := m.Call(arguments)
		if !err[0].IsNil() {
			return err[0].Interface().(error)
		}
	}

	return nil
}
