package server

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoRouterConfig = errors.New("server: no router is configured")
)

const (
	LOGGER = iota + 1
	BUILDER
	ROUTER
	CACHE
)

// fancy fancy o.O
func header() {
	fmt.Println("   __      _ _ _                                                     _            _   _")
	fmt.Println("/  _|     | | | |                                                   | |          | | (_)")
	fmt.Println("| |_ _   _| | | |__   ___  _   _ ___  ___ ______ _ __  _ __ ___   __| |_   _  ___| |_ _  ___  _ __  ___   ___ ___  _ __ ___")
	fmt.Println("|  _| | | | | | '_ \\ / _ \\| | | / __|/ _ \\______| '_ \\| '__/ _ \\ / _` | | | |/ __| __| |/ _ \\| '_ \\/ __| / __/ _ \\| '_ ` _ \\")
	fmt.Println("| | | |_| | | | | | | (_) | |_| \\__ \\  __/      | |_) | | | (_) | (_| | |_| | (__| |_| | (_) | | | \\__ \\| (_| (_) | | | | | |")
	fmt.Println("|_|  \\__,_|_|_|_| |_|\\___/ \\__,_|___/\\___|      | .__/|_|  \\___/ \\__,_|\\__,_|\\___|\\__|_|\\___/|_| |_|___(_)___\\___/|_| |_| |_|")
	fmt.Println("___  ___ _ ____   _____ _ __                    | |")
	fmt.Println("/ __|/ _ \\ '__\\ \\ / / _ \\ '__|                  |_|")
	fmt.Println("\\__ \\  __/ |   \\ V /  __/ |")
	fmt.Println("|___/\\___|_|    \\_/ \\___|_|    ")
}

// Initialize is init the log, builder, router and cache by config.
func Initialize(config interface{}, hooks ...int) error {
	// setting the internal config
	cfg = loadConfig(config)

	var err error
	for _, hook := range hooks {

		switch hook {
		case LOGGER:
			err = initLogger()
			if err != nil {
				return err
			}
		case BUILDER:
			err = initBuilder()
			if err != nil {
				return err
			}
		case ROUTER:
			err = initRouter()
			if err != nil {
				return err
			}
		case CACHE:
			err = initCache()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Run the web-server
func Run() error {
	header()
	err := startServer()
	if err != nil {
		return err
	}

	return nil
}

// startServer starts the server.
// If the force redirect is set to true, all http request will be redirected to https
func startServer() error {

	c, err := config()
	if err != nil {
		return err
	}

	// checking if a router is defined
	if cfgRouter == nil {
		return ErrNoRouterConfig
	}

	// HTTPS Server
	cfgServer := http.Server{}
	cfgServer.Addr = fmt.Sprint(":", c.Server.HTTPPort)

	//TODO write own cors middleware
	//corsManager := cors.New(cors.Options{
	//	AllowCredentials: true,
	//	AllowedOrigins:   []string{"http://localhost:8080"},
	//	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	//	AllowedHeaders:   []string{"Authorization", "Origin", "Cache-Control", "Accept", "Content-Type", "X-Requested-With"},
	//	Debug:            true,
	//})
	//corsManager.Handler(..)

	cfgServer.Handler = cfgRouter.Handler()
	err = cfgServer.ListenAndServe()
	if err != nil {
		return err
	}

	defer cfgServer.Close()

	return nil
}
