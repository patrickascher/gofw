// Copyright by https://github.com/benbjohnson/testing
// Modified by Patrick Ascher (pat@fullhouse-productions.com)
// Thanks @ https://github.com/benbjohnson/testing for the code which was the author
// of the original code here. It just got modified by me.
// TODO check license styles
package test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"database/sql"
	_ "github.com/lib/pq"
	"os"
)

// assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// Err fails the test if an err is not nil.
func Err(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

func Zero(tb testing.TB, act interface{}) {
	if !reflect.ValueOf(act).IsNil() {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: value is not zero \033[39m\n\n", filepath.Base(file), line)
		tb.FailNow()
	}
}

// Err fails the test if an err is not nil.
func Type(tb testing.TB, exp string, act interface{}) {
	act = reflect.TypeOf(reflect.Indirect(reflect.ValueOf(act)).Interface()).String()
	if exp != act {
		pc, file, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		funcName := strings.Split(details.Name(), ".")
		fmt.Printf("\033[31m%s:%d:%s\texp: %#v %#v\tgot: %#v %#v\033[39m\n\n", filepath.Base(file), line, funcName[len(funcName)-1], exp, reflect.ValueOf(exp).Type().String(), act, reflect.ValueOf(act).Type().String())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		pc, file, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)

		funcName := strings.Split(details.Name(), ".")

		expType := ""
		if reflect.ValueOf(exp).IsValid() {
			expType = reflect.ValueOf(exp).Type().String()
		}
		actType := ""
		if reflect.ValueOf(act).IsValid() {
			actType = reflect.ValueOf(act).Type().String()
		}

		fmt.Printf("\033[31m%s:%d:%s\texp: %#v %#v\tgot: %#v %#v\033[39m\n\n", filepath.Base(file), line, funcName[len(funcName)-1], exp, expType, act, actType)
		tb.FailNow()
	}
}

func ConnectionPostgres() *sql.DB {

	host := "localhost"
	port := 5432
	user := "root"
	password := "root"
	database := "fhp"
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return db
}

func ConnectionMysql() *sql.DB {

	host := "localhost"
	port := 3319
	user := "root"
	password := "root"
	database := "fhp"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println(err)
	}

	return db
}

//HTTP
type FakeResponse struct {
	headers http.Header
	body    []byte
	status  int
}

func (r *FakeResponse) Status() int {
	return r.status
}
func (r *FakeResponse) Body() string {
	return string(r.body)
}

func (r *FakeResponse) BodyRaw() []byte {
	return r.body
}

func (r *FakeResponse) Header() http.Header {

	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

func (r *FakeResponse) Write(body []byte) (int, error) {
	r.body = body
	return len(body), nil
}

func (r *FakeResponse) WriteHeader(status int) {
	r.status = status
}
