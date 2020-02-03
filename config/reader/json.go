package reader

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/patrickascher/gofw/config"
	"os"
	"path/filepath"
	"sync"
)

// Json config reader
type Json struct {
	mux sync.Mutex
	env string
}

// JsonOptions for the json config reader
type JsonOptions struct {
	File string
}

// init registers the json reader
func init() {
	config.Register("json", &Json{})
}

// Env sets the environment for the json parser
func (j *Json) Env(env string) {
	j.env = env
}

// Parse the given file into the config struct. sync.mux is used
func (j *Json) Parse(config interface{}, options config.Options) error {

	opt := options.(*JsonOptions)

	if opt.File == "" {
		return errors.New("config: file is not given")
	}

	//only one routine can read the config at a time
	j.mux.Lock()
	defer j.mux.Unlock()

	err := fileOpen(opt.File, config)
	if err != nil {
		return err
	}

	// checking if a env file exists
	path, file := filepath.Dir(opt.File), filepath.Base(opt.File)
	envFile := fmt.Sprintf("%v%v%v%v", path, string(filepath.Separator), j.env+".", file)
	if _, err := os.Stat(envFile); err == nil {
		//ignore error because its only the env file
		fileOpen(envFile, config)
	}

	return nil
}

// fileOpen checks if a file exists and read it into the given config struct
func fileOpen(f string, c interface{}) error {
	//Open the config file and throw an error
	configFile, err := os.Open(f)
	defer configFile.Close()
	if err != nil {
		return err
	}

	//Read and convert the Json to the given struct
	//Throw all decode errors
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&c)
	if err != nil {
		return err
	}

	return nil
}
