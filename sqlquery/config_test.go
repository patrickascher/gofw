package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestConfig_Debugger(t *testing.T) {
	cfg := sqlquery.Config{Debug: true}
	assert.Equal(t, "*logger.logger", reflect.TypeOf(cfg.Debugger()).String())

	cfg = sqlquery.Config{Debug: false}
	assert.IsType(t, nil, cfg.Debugger())
}

func TestConfig_Driver(t *testing.T) {
	cfg := sqlquery.Config{Adapter: "mysql"}
	assert.Equal(t, "mysql", cfg.Driver())
}

func TestConfig_DSN(t *testing.T) {
	cfg := sqlquery.Config{Adapter: "mysql", Username: "root", Password: "toor", Host: "localhost", Port: 1234, Database: "tests"}
	assert.Equal(t, "root:toor@tcp(localhost:1234)/tests?charset=utf8&parseTime=true", cfg.DSN())

	cfg = sqlquery.Config{Adapter: "postgres", Username: "root", Password: "toor", Host: "localhost", Port: 1234, Database: "tests"}
	assert.Equal(t, "host=localhost port=1234 user=root password=toor dbname=tests sslmode=disable", cfg.DSN())

	cfg = sqlquery.Config{Adapter: "postgres", Username: "root", Password: "", Host: "localhost", Port: 1234, Database: "tests"}
	assert.Equal(t, "host=localhost port=1234 user=root  dbname=tests sslmode=disable", cfg.DSN())
}

func TestConfig_Placeholder(t *testing.T) {
	cfg, err := HelperParseConfig()
	assert.NoError(t, err)
	assert.IsType(t, &sqlquery.Placeholder{}, cfg.Placeholder())
}

func TestConfig_QuoteCharacter(t *testing.T) {
	cfg := sqlquery.Config{QuoteChar: "`"}
	assert.Equal(t, "`", cfg.QuoteCharacter())
}
