// Copyright 2020 Patrick Ascher <pat@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package types

// Type interface
type Interface interface {
	Kind() string
	Raw() string
}

// NewInt returns a ptr to a Int.
// It also defines the name and raw.
func NewInt(raw string) *Int {
	return &Int{common: common{name: "Integer", raw: raw}}
}

// NewBool returns a ptr to a Bool.
// It also defines the name and raw.
func NewBool(raw string) *Bool {
	return &Bool{common: common{name: "Bool", raw: raw}}
}

// NewText returns a ptr to a Text.
// It also defines the name and raw.
func NewText(raw string) *Text {
	return &Text{common: common{name: "Text", raw: raw}}
}

// NewTextArea returns a ptr to a TextArea.
// It also defines the name and raw.
func NewTextArea(raw string) *TextArea {
	return &TextArea{common: common{name: "TextArea", raw: raw}}
}

// NewFloat returns a ptr to a NewFloat.
// It also defines the name and raw.
func NewFloat(raw string) *Float {
	return &Float{common: common{name: "Float", raw: raw}}
}

// NewTime returns a ptr to a NewTime.
// It also defines the name and raw.
func NewTime(raw string) *Time {
	return &Time{common: common{name: "Time", raw: raw}}
}

// NewDate returns a ptr to a NewDate.
// It also defines the name and raw.
func NewDate(raw string) *Date {
	return &Date{common: common{name: "Date", raw: raw}}
}

// NewDateTime returns a ptr to a NewDateTime.
// It also defines the name and raw.
func NewDateTime(raw string) *DateTime {
	return &DateTime{common: common{name: "DateTime", raw: raw}}
}

// NewInt returns a ptr to a Int.
// It also defines the name and raw.
func NewEnum(raw string) *Enum {
	return &Enum{common: common{name: "Select", raw: raw}}
}

// NewSet returns a ptr to a NewSet.
// It also defines the name and raw.
//func NewSet(raw string) *Set {
//	return &Set{common: common{name: "MultiSelect", raw: raw}}
//}

type Select interface {
	Items() []string
}

type common struct {
	raw  string
	name string
}

func (c *common) Raw() string {
	return c.raw
}

func (c *common) Kind() string {
	return c.name
}

// Int represents all kind of sql integers
type Int struct {
	Min int64
	Max uint64
	common
}

// Int represents all kind of sql integers
type Bool struct {
	common
}

// Text represents all kind of sql character
type Text struct {
	Size int
	common
}

// TextArea represents all kind of sql text
type TextArea struct {
	Size int
	common
}

// Time represents all kind of sql time
type Time struct {
	common
}

// Date represents all kind of sql dates
type Date struct {
	common
	//Timezone?
}

// DateTime represents all kind of sql dateTimes
type DateTime struct {
	common
}

// Float represents all kind of sql floats
type Float struct {
	common
	//precision
}

// Enum represents all kind of sql enums
type Enum struct {
	Values []string
	common
}

func (e *Enum) Items() []string {
	return e.Values
}

// Set represents all kind of sql sets
type Set struct {
	Values []string
	common
}
