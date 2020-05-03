package orm_test

import (
	"testing"
	"time"

	"github.com/patrickascher/gofw/orm"
	"github.com/stretchr/testify/assert"
)

func TestNewNullInt(t *testing.T) {
	assert.IsType(t, orm.NullInt{}, orm.NewNullInt(1))
	assert.True(t, orm.NewNullInt(1).Valid)
	assert.Equal(t, int64(1), orm.NewNullInt(1).Int64)
}

func TestNewNullString(t *testing.T) {
	assert.IsType(t, orm.NullString{}, orm.NewNullString("John Doe"))
	assert.True(t, orm.NewNullString("John Doe").Valid)
	assert.Equal(t, "John Doe", orm.NewNullString("John Doe").String)
}

func TestNewNullTime(t *testing.T) {
	now := time.Now()
	assert.IsType(t, orm.NullTime{}, orm.NewNullTime(now))
	assert.True(t, orm.NewNullTime(now).Valid)
	assert.Equal(t, now, orm.NewNullTime(now).Time)
}

func TestInt(t *testing.T) {
	assert.Equal(t, 1, orm.Int(1))
	assert.Equal(t, 1, orm.Int(int64(1)))
	assert.Equal(t, 1, orm.Int(orm.NewNullInt(1)))

	nullInt := orm.NewNullInt(1)
	nullInt.Valid = false
	assert.Equal(t, 0, orm.Int(nullInt))

}
