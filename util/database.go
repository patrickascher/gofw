package util

import (
	"strconv"
	"strings"
)

type Dsn struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Sslmode  string //Params?
	Char     string //Params?
}

//DataSourceName from connection
func DataSourceNameConfig(db string) *Dsn {

	dsn := &Dsn{}

	//TODO: find a better solution
	if !strings.Contains(db, "@") {
		tmp := strings.Split(db, "host=")
		dsn.Host = strings.Split(tmp[1], " ")[0]

		tmp = strings.Split(db, "port=")
		dsn.Port, _ = strconv.Atoi(strings.Split(tmp[1], " ")[0])

		tmp = strings.Split(db, "user=")
		dsn.User = strings.Split(tmp[1], " ")[0]

		tmp = strings.Split(db, "password=")
		dsn.Password = strings.Split(tmp[1], " ")[0]

		tmp = strings.Split(db, "dbname=")
		dsn.Dbname = strings.Split(tmp[1], " ")[0]

		tmp = strings.Split(db, "sslmode=")
		dsn.Sslmode = strings.Split(tmp[1], " ")[0]
	} else {
		//username:password@protocol(address:port)/dbname?param=value
		//&{%!s(*mysql.MySQLDriver=&{}) root:root@tcp(localhost:3319)/fhp?charset=utf8 %!s(uint64=0)
		var tmp = []string{}
		if strings.Contains(db, " ") {
			tmp = strings.Split(db, " ")
			tmp = strings.Split(tmp[1], ":")
		} else {
			tmp = strings.Split(db, ":")
		}
		dsn.User = tmp[0]

		tmp = strings.Split(db, ":")
		dsn.Password = strings.Split(tmp[1], "@")[0]

		tmp = strings.Split(db, "@")
		dsn.Host = strings.Split(tmp[1], ")")[0]
		hostpart := strings.Split(strings.Split(tmp[1], ")")[0], "(")
		dsn.Host = strings.Split(hostpart[1], ":")[0]
		dsn.Port, _ = strconv.Atoi(strings.Split(hostpart[1], ":")[1])

		tmp = strings.Split(db, "/")
		dsn.Dbname = strings.Split(tmp[1], "?")[0]
	}

	return dsn
}
