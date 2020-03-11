package sqlquery_test

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/patrickascher/gofw/sqlquery"
	"os"
)

// This package is an exact copy of the helper_test.go.
// I need this functions in both tests (internal and public).
// At the moment this is the only solution i have for it.

// TABLE is the db table name which is used in the tests
const TABLE = "robots"

type Robot struct {
	ID   int
	Name sqlquery.NullString
}

func HelperDeleteEntries(b *sqlquery.Builder) error {
	_, err := b.Delete(TABLE).Exec()
	return err
}

func HelperInsertEntries(b *sqlquery.Builder) error {
	var valueSets []map[string]interface{}

	valueSets = append(valueSets, map[string]interface{}{"id": 1, "name": "Cozmo"})
	valueSets = append(valueSets, map[string]interface{}{"id": 2, "name": "Wall-E"})
	valueSets = append(valueSets, map[string]interface{}{"id": 3, "name": "Spark"})
	valueSets = append(valueSets, map[string]interface{}{"id": 4, "name": "Ubimator"})

	_, err := b.Insert(TABLE).Values(valueSets).Exec()
	return err
}

func HelperParseConfig() (*sqlquery.Config, error) {
	var cfg sqlquery.Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &json.JsonOptions{File: "tests/db.psql.json"})
	}

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func HelperCreateBuilder() (*sqlquery.Builder, error) {
	cfg, err := HelperParseConfig()
	if err != nil {
		return nil, err
	}
	return sqlquery.NewBuilderFromConfig(cfg)
}
