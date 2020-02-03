package sqlquery_test

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// more tests are done in the specific driver
func TestInformation_Describe(t *testing.T) {
	builder, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		// table name only
		cols, err := builder.Information("robots").Describe()
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(cols))
		}

		// qualifier database.table
		cols, err = builder.Information("tests.robots").Describe()
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(cols))
		}

		// database does not exist
		cols2, err := builder.Information("test.robots").Describe()
		if assert.Error(t, err) {
			assert.Equal(t, 0, len(cols2))
		}

		// table does not exist
		cols, err = builder.Information("tests.robot").Describe()
		if assert.Error(t, err) {
			assert.Equal(t, 0, len(cols))
		}
	}
}

func TestInformation_DescribeNoDatabaseInConfig(t *testing.T) {
	var cfg sqlquery.Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/db.psql.json"})
	}
	if assert.NoError(t, err) {

		cfg.Database = ""
		builder, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)

		// table name only
		_, err = builder.Information("robots").Describe()
		assert.Error(t, err)

		// self defined database //TODO this test is disabled because of Postgres + no Database definition
		//cols, err := builder.Information("tests.robots").Describe()
		//assert.NoError(t, err)
		//assert.Equal(t, 2, len(cols))
	}
}

// more tests are done in the specific driver
func TestInformation_ForeignKeys(t *testing.T) {
	var cfg sqlquery.Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/db.json"})
	}

	if assert.NoError(t, err) {

		builder, err := sqlquery.NewBuilderFromConfig(&cfg)
		assert.NoError(t, err)

		// table name only
		fk, err := builder.Information("robots").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(fk))

		// self defined database
		fk, err = builder.Information("tests.robots").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(fk))

		// self defined database
		fk, err = builder.Information("user").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(fk))

		// self defined database
		fk, err = builder.Information("user_posts").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(fk))

		// self defined database
		fk, err = builder.Information("histories").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(fk))

		// self defined database
		fk, err = builder.Information("addresses").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(fk))

		// self defined database
		fk, err = builder.Information("posts").ForeignKeys()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(fk))
	}
}
