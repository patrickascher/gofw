package grid

import (
	"reflect"
)

// value is holding values for different grid modes.
// the grid pointer is needed to get the right value by mode.
type value struct {
	grid *Grid

	table   interface{}
	details interface{}
	create  interface{}
	update  interface{}
}

// NewValue creates a new *value with the given value for all element.
// the grid ptr is added.
func (g *Grid) NewValue(val interface{}) *value {
	v := value{grid: g}
	v.set(val)
	return &v
}

// SetTable sets the value only for the table view
func (v *value) SetTable(val interface{}) *value {
	v.table = val
	return v
}

// SetDetails sets the value only for the details view
func (v *value) SetDetails(val interface{}) *value {
	v.details = val
	return v
}

// SetCreate sets the value only for the create view
func (v *value) SetCreate(val interface{}) *value {
	v.create = val
	return v
}

// SetUpdate sets the value only for the update view
func (v *value) SetUpdate(val interface{}) *value {
	v.update = val
	return v
}

// set is a internal helper to set the value to all mode variables
func (v *value) set(val interface{}) {
	v.table = val
	v.details = val
	v.create = val
	v.update = val
}

// setByValue updates the mode variables by the given value struct
func (v *value) setValueStruct(val *value) {
	v.grid = val.grid
	v.table = val.table
	v.details = val.details
	v.create = val.create
	v.update = val.update
}

// value is a internal helper to get the value by grid mode.
// nil will return if the mode was not defined.
func (v value) value() interface{} {

	if v.grid == nil {
		return nil
	}

	switch v.grid.Mode() {
	case VTable:
		return v.table
	case VDetails:
		return v.details
	case VCreate, CREATE:
		return v.create
	case VUpdate, UPDATE:
		return v.update
	}

	return nil
}

// getInterface converts the result to a interface.
// nil will return as default value.
func (v value) getInterface() interface{} {
	val := v.value()
	if val != nil {
		return val.(interface{})
	}
	return nil
}

// getBool converts the result to a boolean.
// false will return as default value.
func (v value) getBool() bool {
	val := v.value()
	if val != nil {
		return val.(bool)
	}
	return false
}

// getInt converts the result to an integer.
// 0 will return as default value.
func (v value) getInt() int {
	val := v.value()
	if val != nil {
		return val.(int)
	}
	return 0
}

// getString converts the result to a string.
// "" will return as default value.
func (v value) getString() string {
	val := v.value()
	if val != nil {
		return val.(string)
	}
	return ""
}

// setValueHelper is a helper to set the value of a *value struct by string or by another *value struct.
// this is used, that the user can enter the same string for all grid modes or define different ones per mode.
func setValueHelper(field *value, v interface{}) {
	tpe := reflect.TypeOf(v).String()
	switch tpe {
	case reflect.TypeOf(field).String():
		field.setValueStruct(v.(*value))
		return
	default:
		field.set(v)
		return
	}
}
