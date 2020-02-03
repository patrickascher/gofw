package middleware_test

import (
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// LoggerJR test middleware
func LoggerJR(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Write([]byte("Logger-Before "))
		h(w, r, ps)
		w.Write([]byte("Logger-After"))
	}
}

// Logger test middleware
func Logger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Logger-Before "))
		h(w, r)
		w.Write([]byte("Logger-After"))
	}
}

// Auth test middleware
func AuthJR(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Write([]byte("Auth-Before "))
		h(w, r, ps)
		w.Write([]byte("Auth-After "))
	}
}

// Auth test middleware
func Auth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Auth-Before "))
		h(w, r)
		w.Write([]byte("Auth-After "))
	}
}

func TestMiddlewares(t *testing.T) {

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Handler "))
	}

	r, _ := http.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()

	mw := middleware.New(Logger)
	mw.Add(Auth)
	mw.Handle(handlerFunc)(w, r)

	assert.Equal(t, 2, len(mw.GetAll()))
	assert.Equal(t, "Logger-Before Auth-Before Handler Auth-After Logger-After", w.Body.String())
}

// TestMiddlewares_AddAndHandle testing if the middleware is getting added and called in the right order
func TestMiddlewares_JR(t *testing.T) {

	handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("Handler "))
	}

	r, _ := http.NewRequest("GET", "/one", nil)
	w := httptest.NewRecorder()
	var p []httprouter.Param

	mw := middleware.NewJR(LoggerJR)
	mw.Add(AuthJR)
	mw.Handle(handlerFunc)(w, r, p)

	assert.Equal(t, 2, len(mw.GetAll()))
	assert.Equal(t, "Logger-Before Auth-Before Handler Auth-After Logger-After", w.Body.String())
}
