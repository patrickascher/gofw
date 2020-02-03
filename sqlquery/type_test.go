package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewInt(t *testing.T) {
	i := sqlquery.NewInt("int(11) unsigned")
	assert.IsType(t, &sqlquery.Int{}, i)
	assert.Equal(t, "int(11) unsigned", i.Raw())
	assert.Equal(t, "Integer", i.Kind())
	assert.Equal(t, int64(0), i.Min)
	assert.Equal(t, uint64(0), i.Max)
}

func TestNewFloat(t *testing.T) {
	i := sqlquery.NewFloat("float")
	assert.IsType(t, &sqlquery.Float{}, i)
	assert.Equal(t, "float", i.Raw())
	assert.Equal(t, "Float", i.Kind())
}

func TestNewText(t *testing.T) {
	i := sqlquery.NewText("varchar")
	assert.IsType(t, &sqlquery.Text{}, i)
	assert.Equal(t, "varchar", i.Raw())
	assert.Equal(t, "Text", i.Kind())
	assert.Equal(t, 0, i.Size)

}

func TestNewTextArea(t *testing.T) {
	i := sqlquery.NewTextArea("text")
	assert.IsType(t, &sqlquery.TextArea{}, i)
	assert.Equal(t, "text", i.Raw())
	assert.Equal(t, "TextArea", i.Kind())
}

func TestNewTime(t *testing.T) {
	i := sqlquery.NewTime("time")
	assert.IsType(t, &sqlquery.Time{}, i)
	assert.Equal(t, "time", i.Raw())
	assert.Equal(t, "Time", i.Kind())
}

func TestNewDate(t *testing.T) {
	i := sqlquery.NewDate("date")
	assert.IsType(t, &sqlquery.Date{}, i)
	assert.Equal(t, "date", i.Raw())
	assert.Equal(t, "Date", i.Kind())
}

func TestNewDateTime(t *testing.T) {
	i := sqlquery.NewDateTime("datetime")
	assert.IsType(t, &sqlquery.DateTime{}, i)
	assert.Equal(t, "datetime", i.Raw())
	assert.Equal(t, "DateTime", i.Kind())
}

func TestNewEnum(t *testing.T) {
	i := sqlquery.NewEnum("enum('A','B')")
	assert.IsType(t, &sqlquery.Enum{}, i)
	assert.Equal(t, "enum('A','B')", i.Raw())
	assert.Equal(t, "Select", i.Kind())
	assert.Equal(t, 0, len(i.Values))
}

func TestNewSet(t *testing.T) {
	i := sqlquery.NewSet("set('A','B')")
	assert.IsType(t, &sqlquery.Set{}, i)
	assert.Equal(t, "set('A','B')", i.Raw())
	assert.Equal(t, "MultiSelect", i.Kind())
	assert.Equal(t, 0, len(i.Values))
}
