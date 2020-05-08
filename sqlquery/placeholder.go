// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sqlquery

import (
	"strconv"
)

// PLACEHOLDER character.
const PLACEHOLDER_APPEND = "§§"
const PLACEHOLDER = "?"

// Placeholder is used to ensure an unique placeholder for different database adapters.
type Placeholder struct {
	Numeric bool   // must be true if the database uses something like $1,$2,...
	counter int    // internal counter for numeric placeholder
	Char    string // database placeholder character
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
