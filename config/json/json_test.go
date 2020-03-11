// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package json_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/patrickascher/gofw/config/json"
	"github.com/stretchr/testify/assert"
)

type mockServerConfig struct {
	Ip   string
	Port int
}

type mockUserConfig struct {
	Name     string
	Password string
}

type mockConfig struct {
	Server mockServerConfig
	User   mockUserConfig
}

// mockFiles creates three files, config.json (correct), dev.config.json (correct) and an incorrect json file fail.config.json.
func mockFiles() {
	// config.json - correct json file
	f1 := []byte("{\"server\": {\"ip\": \"127.0.0.1\", \"port\": 8080}, \"user\": {\"name\": \"root\", \"password\": \"toor\"}}")
	err := ioutil.WriteFile("config.json", f1, 0644)
	if err != nil {
		fmt.Println(err)
	}
	// dev.config.json - correct env json file
	f2 := []byte("{\"user\": {\"name\": \"dev\", \"password\": \"ved\"}}")
	err = ioutil.WriteFile("dev.config.json", f2, 0644)
	if err != nil {
		fmt.Println(err)
	}
	// fail.config.json - incorrect json format
	f3 := []byte("{\"user\": {\"name\": \"dev\", \"password\": ved\"}}")
	err = ioutil.WriteFile("fail.config.json", f3, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// removeMockFiles removes all test files
func removeMockFiles() {
	err := os.Remove("config.json")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove("dev.config.json")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Remove("fail.config.json")
	if err != nil {
		fmt.Println(err)
	}
}

// TestJson_Parse testing if the files are getting parsed correctly into the config struct
func TestJson_Parse(t *testing.T) {
	mockFiles()
	defer removeMockFiles()

	jc := json.New()

	// table driven tests
	var tests = []struct {
		Error  bool
		Env    string
		File   string
		Result *mockConfig
	}{
		{Error: false, File: "config.json", Env: "", Result: &mockConfig{Server: mockServerConfig{Ip: "127.0.0.1", Port: 8080}, User: mockUserConfig{Name: "root", Password: "toor"}}},
		{Error: false, File: "config.json", Env: "dev", Result: &mockConfig{Server: mockServerConfig{Ip: "127.0.0.1", Port: 8080}, User: mockUserConfig{Name: "dev", Password: "ved"}}},
		{Error: true, File: "404.config.json", Env: "404", Result: &mockConfig{}},
		{Error: true, File: "config.json", Env: "fail", Result: &mockConfig{Server: mockServerConfig{Ip: "127.0.0.1", Port: 8080}, User: mockUserConfig{Name: "root", Password: "toor"}}},
		{Error: true, File: "", Env: "", Result: &mockConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.Env, func(t *testing.T) {
			conf := &mockConfig{}
			err := jc.Parse(conf, tt.Env, json.Options{Filepath: tt.File})
			if tt.Error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.Result, conf)
		})
	}
}
