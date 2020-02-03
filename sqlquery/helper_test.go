package sqlquery

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"os"
)

// This package is an exact copy of the helper_internal_test.go.
// I need this functions in both tests (internal and public).
// At the moment this is the only solution i have for it.

// TABLE is the db table name which is used in the tests
const TABLE = "robots"

type Robot struct {
	ID   int
	Name NullString
}

func HelperDeleteEntries(b *Builder) error {
	_, err := b.Delete(TABLE).Exec()
	return err
}

func HelperInsertEntries(b *Builder) error {
	var valueSets []map[string]interface{}

	valueSets = append(valueSets, map[string]interface{}{"id": 1, "name": "Cozmo"})
	valueSets = append(valueSets, map[string]interface{}{"id": 2, "name": "Wall-E"})
	valueSets = append(valueSets, map[string]interface{}{"id": 3, "name": "Spark"})
	valueSets = append(valueSets, map[string]interface{}{"id": 4, "name": "Ubimator"})

	_, err := b.Insert(TABLE).Values(valueSets).Exec()
	return err
}

func HelperParseConfig() (*Config, error) {
	var cfg Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/db.psql.json"})
	}

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func HelperCreateBuilder() (*Builder, error) {
	cfg, err := HelperParseConfig()

	if err != nil {
		return nil, err
	}
	return NewBuilderFromConfig(cfg)
}
