package reader_test

import (
	"fmt"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

// HelperCreateFiles creating some dummy test files (config.json, dev.config.json)
func HelperCreateFiles() {
	f1 := []byte("{\"server\": {\"ip\": \"127.0.0.1\", \"port\": 8080}, \"user\": {\"name\": \"root\", \"password\": \"toor\"}}")
	err := ioutil.WriteFile("config.json", f1, 0644)
	if err != nil {
		fmt.Println(err)
	}
	f2 := []byte("{\"user\": {\"name\": \"dev\", \"password\": \"ved\"}}")
	err = ioutil.WriteFile("dev.config.json", f2, 0644)
	if err != nil {
		fmt.Println(err)
	}

	f3 := []byte("{\"user\": {\"name\": \"dev\", \"password\": ved\"}}")
	err = ioutil.WriteFile("fail.config.json", f3, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

// removeTestFiles removing test.json and development.test.json
func HelperRemoveFiles() {
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

// TestJsonConfig_Parse testing if the file are getting parsed correctly into the config struct
func TestJsonConfig_Parse(t *testing.T) {
	HelperCreateFiles()

	type ServerConfig struct {
		Ip   string
		Port int
	}

	type UserConfig struct {
		Name     string
		Password string
	}

	type Config struct {
		Server ServerConfig
		User   UserConfig
	}

	conf := Config{}
	jc := reader.Json{}

	// test file
	err := jc.Parse(&conf, &reader.JsonOptions{File: "config.json"})
	assert.NoError(t, err)
	assert.Equal(t, Config{Server: ServerConfig{Ip: "127.0.0.1", Port: 8080}, User: UserConfig{Name: "root", Password: "toor"}}, conf)

	// test file without file information
	err = jc.Parse(&conf, &reader.JsonOptions{})
	assert.Error(t, err)

	// env test file
	jc.Env("dev")
	err = jc.Parse(&conf, &reader.JsonOptions{File: "config.json"})
	assert.NoError(t, err)
	assert.Equal(t, Config{Server: ServerConfig{Ip: "127.0.0.1", Port: 8080}, User: UserConfig{Name: "dev", Password: "ved"}}, conf)

	// file does not exist
	err = jc.Parse(&conf, &reader.JsonOptions{File: "404.config.json"})
	assert.Error(t, err)

	// not a correct json file
	err = jc.Parse(&conf, &reader.JsonOptions{File: "fail.config.json"})
	assert.Error(t, err)

	HelperRemoveFiles()
}
