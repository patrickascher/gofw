// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package json implements the config.Interface and registers a json provider.
// All operations are using a sync.RWMutex for synchronization.
// Benchmark file is available.
//
// Check the json.Options for the available configurations.
package json

import (
	encJson "encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/patrickascher/gofw/config"
)

// Error messages
var (
	ErrFilepath = errors.New("config/json: Filepath is missing")
)

// json config provider
type json struct {
	mux sync.Mutex
}

// Options for the json config reader
type Options struct {
	// The Filepath is mandatory.
	// The given file will get decoded and marshaled into the given config struct.
	// If a environment file exists, it will be merged with a higher priority.
	// This means a common configuration could be written in conf.json and the env configuration is getting merged together.
	// Prefix the file with the environment followed by a dot. (dev.conf.json, staging.conf.json, production.conf.json, ...)
	Filepath string
}

// init registers the json provider
func init() {
	_ = config.Register(config.JSON, New)
}

// New satisfies the config.provider interface.
func New() config.Interface {
	return &json{}
}

// Parse the given file into the config struct. sync.mux is used for synchronisation.
// File and env file are getting decoded/marshaled into the config struct (please see json.Options for more details).
// If the filepath is not set, file does not exist or the json can not get decoded, an error will return.
func (j *json) Parse(config interface{}, env string, options interface{}) error {
	// checking if the config Filepath is set
	opt := options.(Options)
	if opt.Filepath == "" {
		return ErrFilepath
	}

	//sync.Mutex is getting locked.
	j.mux.Lock()
	defer j.mux.Unlock()

	// opening filepath and write it to the config struct
	err := fileOpen(opt.Filepath, config)
	if err != nil {
		return err
	}

	// check if an env file exists and is no dir
	envFile := fmt.Sprintf("%v%v%v%v", filepath.Dir(opt.Filepath), string(filepath.Separator), env+".", filepath.Base(opt.Filepath))
	if info, err := os.Stat(envFile); err == nil && !info.IsDir() {
		//ignore error because its only the env file
		err = fileOpen(envFile, config)
		if err != nil {
			return err
		}
	}

	return nil
}

// fileOpen checks if a file exists and read it into the given config struct.
func fileOpen(f string, c interface{}) (err error) {
	//Open the config file
	configFile, err := os.Open(f)
	if err != nil {
		return
	}

	// like this the configFile.Close error is getting handled.
	defer func() {
		cErr := configFile.Close()
		if err == nil {
			err = cErr
		}
	}()

	//Read and convert the Json to the given struct
	jsonParser := encJson.NewDecoder(configFile)
	return jsonParser.Decode(&c)
}
