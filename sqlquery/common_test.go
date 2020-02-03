package sqlquery_test

import (
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestBuilder_QuoteIdentifier(t *testing.T) {
	cfg, err := sqlquery.HelperParseConfig()
	if assert.NoError(t, err) {
		b, err := sqlquery.HelperCreateBuilder()
		if assert.NoError(t, err) {
			assert.Equal(t, strings.Replace("$robots$", "$", cfg.QuoteChar, -1), b.QuoteIdentifier("robots"))
			assert.Equal(t, strings.Replace("$robots$.$name$", "$", cfg.QuoteChar, -1), b.QuoteIdentifier("robots.name"))
		}
	}
}
