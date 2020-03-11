// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing an undefined log level
func TestLevel_String(t *testing.T) {
	assert.Equal(t, errUnknownLogLevel, level(7).String())
}
