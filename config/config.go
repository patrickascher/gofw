// Package config is a module to handel different configurations.
// It can handle different configuration based on the environment.
// Its easy and simple to create different readers, to handle different config backends.
// Out of the box json is supported.
// See https://github.com/patrickascher/go-config for more information and examples.
package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// readerStore saves all registered readers by its name.
var readerStore = make(map[string]config)
var env string

//JSON is a global constant for the JSON reader
const JSON = "json"

//ENV is the default string which is checked in the os.GetEnv() function
const ENV = "ENV"

// Error messages from the whole package
var (
	ErrNoReader            = errors.New("config: reader-name or the reader was not set")
	ErrReaderNotExist      = errors.New("config: reader &#v does not exist")
	ErrReaderAlreadyExists = errors.New("config: reader %#v already exists")
)

// Options is an empty interface, every reader has to handle it on its own
type Options interface {
}

type config interface {
	Parse(config interface{}, options Options) error
	Env(string)
}

// Register is used to register the readers
// This function should be called in the init function of the reader to register itself on import.
// It returns an error if the reader-name or the reader itself is empty
func Register(readerName string, reader config) error {

	if reader == nil || readerName == "" {
		return ErrNoReader
	}

	if _, ok := readerStore[readerName]; ok {
		return fmt.Errorf(ErrReaderAlreadyExists.Error(), readerName)
	}

	readerStore[readerName] = reader
	return nil
}

// Env can be used if you store your environment somewhere else than in os.GetEnv("ENV")
func Env(e string) {
	env = e
}

// Parse is calling the Parse function of the reader.
// It will return an error if the reader Parse does so.
func Parse(readerName string, config interface{}, options Options) error {

	if reader, ok := readerStore[readerName]; ok {

		// adding the environment to the reader
		if env != "" {
			reader.Env(env)
		} else {
			reader.Env(os.Getenv(ENV))
		}

		// calling the parse function of the reader with the given options
		err := reader.Parse(config, options)
		if err != nil {
			return err
		}

		return nil
	}

	return ErrReaderNotExist
}

// IsSet checks if a specific field exists and has (no) zero value (recursively).
// By default its checking if the field has an none zero value, if a zero value should be allowed, prefix the field with a 0 (example: "0User.Role.Name")
// example: User.Role.Name searches if the User struct has a field Role and the struct Role has a field Name which has no zero value.
// At the moment only struct is implemented.
func IsSet(s string, cfg interface{}) bool {
	allowZero := ""
	if strings.HasPrefix(s, "0") {
		s = s[1:]
		allowZero = "0"
	}
	search := strings.Split(s, ".")

	t := reflect.TypeOf(cfg).Kind()
	switch t {
	case reflect.Struct:
		for i, v := range search {
			rv := reflect.Indirect(reflect.ValueOf(cfg))
			if !rv.FieldByName(v).IsValid() {
				return false
			}

			if len(search)-1 > i {
				return IsSet(allowZero+strings.Join(search[i+1:], "."), rv.FieldByName(v).Interface())
			}

			if allowZero == "0" {
				return true
			}
			t := reflect.TypeOf(rv.FieldByName(v).Interface())
			return rv.FieldByName(v).Interface() != reflect.Zero(t).Interface()
		}

	}

	return false
}
