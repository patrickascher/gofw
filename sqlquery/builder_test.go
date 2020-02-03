package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"

	"database/sql"
	_ "github.com/patrickascher/gofw/sqlquery/mysql"
	_ "github.com/patrickascher/gofw/sqlquery/postgres"
	"reflect"
)

func TestNewBuilderFromConfig(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()

	assert.IsType(t, &sqlquery.Config{}, b.Config())
	if assert.NoError(t, err) {
		assert.IsType(t, &sqlquery.Builder{}, b)
	}
}

// TestNewBuilderFromConfig2 should fail because of a wrong adapter config
func TestNewBuilderFromConfig2(t *testing.T) {
	conf := sqlquery.Config{}
	_, err := sqlquery.NewBuilderFromConfig(&conf)
	assert.Error(t, err)
}

func TestNewBuilderFromAdapter(t *testing.T) {
	cfg, err := sqlquery.HelperParseConfig()
	if assert.NoError(t, err) {
		db, errDb := sql.Open("mysql", cfg.DSN())
		if assert.NoError(t, errDb) {
			assert.IsType(t, &sql.DB{}, db)

			b := sqlquery.NewBuilderFromAdapter(db, cfg)
			assert.IsType(t, &sqlquery.Builder{}, b)
		}
	}
}

func TestBuilder_Select(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, "*sqlquery.Select", reflect.TypeOf(b.Select(sqlquery.TABLE)).String())
	}
}

func TestBuilder_Insert(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, "*sqlquery.Insert", reflect.TypeOf(b.Insert(sqlquery.TABLE)).String())
	}
}

func TestBuilder_Update(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, "*sqlquery.Update", reflect.TypeOf(b.Update(sqlquery.TABLE)).String())
	}
}

func TestBuilder_Delete(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()
	if assert.NoError(t, err) {
		assert.Equal(t, "*sqlquery.Delete", reflect.TypeOf(b.Delete(sqlquery.TABLE)).String())
	}
}

// more test for the transaction is handled in the builder_internal_tests.go
func TestBuilder_BeginnTx_CommitTx(t *testing.T) {
	b, err := sqlquery.HelperCreateBuilder()
	if assert.NoError(t, err) {
		tx, err := b.NewTx()
		if assert.NoError(t, err) {
			assert.IsType(t, &sql.Tx{}, tx)
			err = b.CommitTx(tx)
			assert.NoError(t, err)
		}
	}
}
