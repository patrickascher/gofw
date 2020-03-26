package orm

import (
	"database/sql"
	"github.com/patrickascher/gofw/cache"
	"github.com/patrickascher/gofw/cache/memory"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Users struct {
	Model

	Id        int
	FirstName sql.NullString
	LastName  sql.NullString
}

type Person struct {
	Model
}

type VeryImportantPerson struct {
	Model
}

type Custom struct {
	Model
}

func (c *Custom) TableName() string {
	return "parts"
}

func (c *Custom) DatabaseName() string {
	return "robots"
}

func (c *Custom) Builder() (*sqlquery.Builder, error) {
	return HelperCreateBuilder()
}

func (c *Custom) DefaultCache() (cache.Interface, time.Duration, error) {

	ca, err := cache.New(cache.MEMORY, memory.Options{GCInterval: 3600})
	return ca, 3600, err
}

func Test_structName(t *testing.T) {
	u := Users{}
	assert.Equal(t, "Users", structName(u, false))
	assert.Equal(t, "Users", structName(&u, false))
}

func Test_structNameWithNamespace(t *testing.T) {
	u := Users{}
	assert.Equal(t, "orm.Users", structName(u, true))
	assert.Equal(t, "orm.Users", structName(&u, true))
}

func TestModel_TableName(t *testing.T) {
	u := &Users{}
	u.caller = u
	assert.Equal(t, "users", u.TableName())

	p := Person{}
	p.caller = &p
	assert.Equal(t, "people", p.TableName())

	vip := &VeryImportantPerson{}
	vip.caller = vip
	assert.Equal(t, "very_important_people", vip.TableName())

	c := &Custom{}
	c.caller = c
	assert.Equal(t, "parts", c.TableName())

}

func TestModel_DatabaseName(t *testing.T) {
	u := &Users{}
	assert.Equal(t, "", u.DatabaseName())
	c := &Custom{}
	assert.Equal(t, "robots", c.DatabaseName())
}

func TestModel_Cache(t *testing.T) {
	u := &Users{}
	u.caller = u
	_, ttl, err := u.Cache()
	if assert.NoError(t, err) {
		assert.Equal(t, time.Duration(21600000000000), ttl)
	}

	custom := &Custom{}
	custom.caller = custom
	_, ttl, err = custom.Cache()
	if assert.NoError(t, err) {
		assert.Equal(t, time.Duration(3600), ttl)
	}
}

func TestModel_Builder(t *testing.T) {

	GlobalBuilder = nil

	cust := &Customerfk{}
	b, err := cust.Builder()
	if assert.Error(t, err) {
		assert.Equal(t, (*sqlquery.Builder)(nil), b)
	}

	custom := &Custom{}
	cbuilder, err := custom.Builder()
	assert.NoError(t, err)
	assert.True(t, cbuilder != nil)

	builder, err := HelperCreateBuilder()
	if assert.NoError(t, err) {
		GlobalBuilder = builder
		b, err := cust.Builder()
		if assert.NoError(t, err) {
			assert.Equal(t, builder, b)
		}
	}
}
