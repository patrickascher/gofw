package sqlquery

import (
	"strconv"
)

// PLACEHOLDER character used in this package
const PLACEHOLDER = "?"

// Placeholder is used to ensure an unique placeholder with different database adapters.
type Placeholder struct {
	Numeric bool   `json:"numeric"` //must be true if the database uses something like $1,$2,...
	counter int    //internal counter for numeric placeholder
	Char    string `json:"char"` //database placeholder character
}

// reset the placeholder counter
func (p *Placeholder) reset() {
	p.counter = 0
}

// hasCounter returns true if the counter is numeric
func (p *Placeholder) hasCounter() bool {
	return p.Numeric
}

// placeholder returns the placeholder character.
// If the placeholder is numeric, the counter will be added as well.
func (p *Placeholder) placeholder() string {
	if p.hasCounter() {
		p.counter++
		return p.Char + strconv.Itoa(p.counter)
	}
	return p.Char
}
