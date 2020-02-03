package util

import (
	"github.com/patrickascher/gofw/util/test"
	"testing"
)

func TestDatabase_DataSourceNameConfigPostgres(t *testing.T) {
	dsn := DataSourceNameConfig("host=localhost port=4567 user=username password=password dbname=fhp sslmode=disable")

	test.Equals(t, dsn.User, "username")
	test.Equals(t, dsn.Password, "password")
	test.Equals(t, dsn.Host, "localhost")
	test.Equals(t, dsn.Port, 4567)
	test.Equals(t, dsn.Dbname, "fhp")
}

func TestDatabase_DataSourceNameConfigMysql(t *testing.T) {
	dsn := DataSourceNameConfig("username:password@tcp(localhost:3319)/fhp?charset=utf8")
	test.Equals(t, dsn.User, "username")
	test.Equals(t, dsn.Password, "password")
	test.Equals(t, dsn.Host, "localhost")
	test.Equals(t, dsn.Port, 3319)
	test.Equals(t, dsn.Dbname, "fhp")

	dsn = DataSourceNameConfig("&{%!s(*mysql.MySQLDriver=&{}) root:root@tcp(localhost:3319)/fhp?charset=utf8 %!s(uint64=0) {%!s(int32=0) %!s(uint32=0)} [] map[] %!s(uint64=0) %!s(int=0) %!s(chan struct {}=0xc420162180) %!s(bool=false) map[] map[] %!s(int=0) %!s(int=0) %!s(time.Duration=0) %!s(chan struct {}=<nil>)}")
	test.Equals(t, dsn.User, "root")
	test.Equals(t, dsn.Password, "root")
	test.Equals(t, dsn.Host, "localhost")
	test.Equals(t, dsn.Port, 3319)
	test.Equals(t, dsn.Dbname, "fhp")
}
