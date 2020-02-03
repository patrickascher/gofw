package middleware

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/logger"
	"reflect"
)

//LoggerJR for julienschmidt-router prints an info to the console
func LoggerJR(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		loggerLogic(h, w, r, ps)
	}
}

//Logger prints an info to the console
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggerLogic(h, w, r)
	}
}

func loggerLogic(handler interface{}, args ...interface{}) {

	var w http.ResponseWriter
	var r *http.Request
	var ps httprouter.Params
	var hJR httprouter.Handle
	var h http.HandlerFunc

	for k, arg := range args {
		switch k {
		case 0:
			w = arg.(http.ResponseWriter)
		case 1:
			r = arg.(*http.Request)
		case 2:
			ps = arg.(httprouter.Params)
		}

	}

	l, _ := logger.Get(logger.CONSOLE)
	start := time.Now()

	crw := newResponseWriter(w)

	if reflect.TypeOf(handler).String() == "httprouter.Handle" {
		hJR = handler.(httprouter.Handle)
		hJR(crw, r, ps)
	} else {
		h = handler.(http.HandlerFunc)
		h(crw, r)
	}

	l.Info("(%s) \x1b[33m%s\x1b[39m %s %s\" %d %d \x1b[33m%s\x1b[39m\n", r.RemoteAddr, r.Method, r.URL.Path, r.Proto, crw.status, crw.size, time.Since(start))
}

// responseWriter is a custom response writer so that we can read the response size and HTTP code
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader is getting the HTTP status of the response
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write is getting the size of the response
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// newResponseWriter returns a new rw struct with some default values
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         200,
	}
}
