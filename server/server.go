package server

import (
	"errors"
	"fmt"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// HTTP server - redirect http to https if forceHTTPS is set to true
	redirectServer := http.Server{}
	redirectServer.Addr = fmt.Sprint(":", c.Server.HTTPPort)
	if cfg.Server.ForceHTTPS {
		redirectServer.Handler = http.HandlerFunc(redirect)
	} else {
		redirectServer.Handler = cfgRouter.Handler()
	}
	go redirectServer.ListenAndServe()

	// HTTPS Server
	cfgServer := http.Server{}
	cfgServer.Addr = fmt.Sprint(":", c.Server.HTTPSPort)

	//TODO write own cors middleware
	corsManager := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Origin", "Cache-Control", "Accept", "Content-Type", "X-Requested-With"},
		Debug:            true,
	})

	cfgServer.Handler = corsManager.Handler(cfgRouter.Handler())
	err = cfgServer.ListenAndServeTLS(c.Server.CertFile, c.Server.KeyFile)
	if err != nil {
		return err
	}

	defer redirectServer.Close()
	defer cfgServer.Close()

	return nil
}

// redirect handler for forcing http to https
func redirect(w http.ResponseWriter, req *http.Request) {
	c, _ := config()
	target := "https://" + strings.Replace(req.Host, strconv.Itoa(c.Server.HTTPPort), strconv.Itoa(c.Server.HTTPSPort), 1) + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
