package sqlquery

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlaceholderChar(t *testing.T) {
	p := Placeholder{Char: "?"}
	assert.Equal(t, "?", p.placeholder())

	p = Placeholder{Char: "$"}
	assert.Equal(t, "$", p.placeholder())
}

func TestPlaceholderNumeric(t *testing.T) {
	p := Placeholder{Numeric: true, Char: "?"}
	assert.Equal(t, "?1", p.placeholder())

	p = Placeholder{Numeric: true, Char: "$"}
	assert.Equal(t, "$1", p.placeholder())
}

func TestPlaceholderCounter(t *testing.T) {
	p := Placeholder{Numeric: true, Char: "?"}
	assert.Equal(t, "?1", p.placeholder())
	assert.Equal(t, "?2", p.placeholder())
	assert.Equal(t, "?3", p.placeholder())

	p.reset()

	assert.Equal(t, "?1", p.placeholder())
}
