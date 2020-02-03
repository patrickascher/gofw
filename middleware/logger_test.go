package middleware_test

import (
	"github.com/julienschmidt/httprouter"
	"github.com/patrickascher/gofw/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMiddlewares_LoggerJR(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.WriteHeader(200)
		w.Write([]byte("new Writer JR"))
	}

	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	w := httptest.NewRecorder()
	var p []httprouter.Param

	mw := middleware.NewJR(middleware.LoggerJR)
	mw.Handle(handlerFunc)(w, r, p)

	assert.Equal(t, "new Writer JR", w.Body.String())
}

func TestMiddlewares_Logger(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("new Writer"))
	}

	r, _ := http.NewRequest("GET", "https://example.org/path?foo=bar", nil)
	w := httptest.NewRecorder()

	mw := middleware.New(middleware.Logger)
	mw.Handle(handlerFunc)(w, r)

	assert.Equal(t, "new Writer", w.Body.String())
}

func TestMiddlewares_LoggerConsole(t *testing.T) {

	if os.Getenv("cmdRun") == "1" {
		handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		}

		r, _ := http.NewRequest("GET", "/one", nil)
		w := httptest.NewRecorder()
		var p []httprouter.Param

		mw := middleware.NewJR(middleware.LoggerJR)
		mw.Handle(handlerFunc)(w, r, p)

		assert.Equal(t, "", w.Body.String())
		return
	}

	//Calling the Test in a different GO routine
	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), "cmdRun=1")
	o, _ := cmd.Output()

	assert.Equal(t, true, cmd.ProcessState.Success())
	lines := strings.Split(string(o), "\n")

	//skipping empty lines, PASS text and coverage text from gotest
	c := 0
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) == 0 || lines[i] == "PASS" || strings.HasPrefix(lines[i], "coverage:") {
			continue
		}
		c++
	}
	assert.Equal(t, 1, c)
	assert.Contains(t, lines[0], "INFO")

}
