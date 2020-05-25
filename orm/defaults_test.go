package orm_test

import (
	"testing"
	"time"

	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/logger"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

const testSchema = "public"
const testDatabase = "orm_test"
const testStrategy = "eager"

// test orm model with custom default values
type mockModel struct {
	orm.Model
}

func (m mockModel) DefaultBuilder() sqlquery.Builder {
	return sqlquery.Builder{}
}

func (m mockModel) DefaultCache() (cache.Interface, time.Duration, error) {
	return nil, 5 * time.Hour, nil
}

func (m mockModel) DefaultLogger() *logger.Logger {
	return nil
}

func (m mockModel) DefaultSchemaName() string {
	return "mock_schema"
}

func (m mockModel) DefaultTableName() string {
	return "mock_table"
}

func (m mockModel) DefaultDatabaseName() string {
	return "mock_db"
}

func (m mockModel) DefaultStrategy() string {
	return "mock_strategy"
}

// Test default builder.
// todo: create a test which tests more than just the type?
func TestModel_DefaultBuilder(t *testing.T) {
	r := car{}
	assert.IsType(t, sqlquery.Builder{}, r.DefaultBuilder())

	rc := mockModel{}
	assert.IsType(t, sqlquery.Builder{}, rc.DefaultBuilder())
}

// test default cache
func TestModel_DefaultCache(t *testing.T) {
	r := car{}
	c, ttl, err := r.DefaultCache()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(cache.INFINITY), ttl)
	assert.NotNil(t, c)

	rc := mockModel{}
	c, ttl, err = rc.DefaultCache()
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Hour, ttl)
	assert.Nil(t, c)
}

// test default logger
func TestModel_DefaultLogger(t *testing.T) {
	r := car{}
	assert.IsType(t, &logger.Logger{}, r.DefaultLogger())
	assert.NotNil(t, r.DefaultLogger())

	rc := mockModel{}
	assert.Nil(t, rc.DefaultLogger())
}

func TestModel_DefaultSchemaName(t *testing.T) {
	r := car{}
	assert.Equal(t, testSchema, r.DefaultSchemaName())

	rc := mockModel{}
	assert.Equal(t, "mock_schema", rc.DefaultSchemaName())
}

func TestModel_DefaultTableName(t *testing.T) {
	r := car{}
	assert.Equal(t, "", r.DefaultTableName())

	// robot must be initialized because its saved as model.name as performance reasons.
	err := r.Init(&r)
	if assert.NoError(t, err) {
		assert.Equal(t, "cars", r.DefaultTableName())
	}

	rc := mockModel{}
	assert.Equal(t, "mock_table", rc.DefaultTableName())
}

func TestModel_DefaultDatabaseName(t *testing.T) {
	r := car{}
	assert.Equal(t, testDatabase, r.DefaultDatabaseName())

	rc := mockModel{}
	assert.Equal(t, "mock_db", rc.DefaultDatabaseName())
}

func TestModel_DefaultStrategy(t *testing.T) {
	r := car{}
	assert.Equal(t, testStrategy, r.DefaultStrategy())

	rc := mockModel{}
	assert.Equal(t, "mock_strategy", rc.DefaultStrategy())
}
