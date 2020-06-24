package grid

import (
	"reflect"
)

// value is holding values for different grid modes.
// the grid pointer is needed to get the right value by mode.
type value struct {
	table   interface{}
	details interface{}
	create  interface{}
	update  interface{}
	export  interface{}
}

// NewValue creates a new *value with the given value for all element.
// the grid ptr is added.
func NewValue(val interface{}) *value {
	v := value{}
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

// SetExport sets the value only for the export view
func (v *value) SetExport(val interface{}) *value {
	v.export = val
	return v
}

// set is a internal helper to set the value to all mode variables
func (v *value) set(val interface{}) {
	v.table = val
	v.details = val
	v.create = val
	v.update = val
	v.export = val
}

func setValueString(s interface{}) map[int]string {
	f := make(map[int]string, 5)
	tpe := reflect.TypeOf(s).String()
	switch tpe {
	case "*grid.value":
		s := s.(*value)
		f[VTable] = s.table.(string)
		f[VCreate] = s.create.(string)
		f[VDetails] = s.details.(string)
		f[VUpdate] = s.update.(string)
		f[Export] = s.export.(string)
	default:
		f[VTable] = s.(string)
		f[VCreate] = s.(string)
		f[VDetails] = s.(string)
		f[VUpdate] = s.(string)
		f[Export] = s.(string)
	}
	return f
}

func setValueInt(s interface{}) map[int]int {
	f := make(map[int]int, 5)
	tpe := reflect.TypeOf(s).String()
	switch tpe {
	case "*grid.value":
		s := s.(*value)
		f[VTable] = s.table.(int)
		f[VCreate] = s.create.(int)
		f[VDetails] = s.details.(int)
		f[VUpdate] = s.update.(int)
		f[Export] = s.export.(int)
	default:
		f[VTable] = s.(int)
		f[VCreate] = s.(int)
		f[VDetails] = s.(int)
		f[VUpdate] = s.(int)
		f[Export] = s.(int)
	}
	return f
}

func setValueBool(s interface{}) map[int]bool {
	f := make(map[int]bool, 5)
	tpe := reflect.TypeOf(s).String()
	switch tpe {
	case "*grid.value":
		s := s.(*value)
		f[VTable] = s.table.(bool)
		f[VCreate] = s.create.(bool)
		f[VDetails] = s.details.(bool)
		f[VUpdate] = s.update.(bool)
		f[Export] = s.export.(bool)
	default:
		f[VTable] = s.(bool)
		f[VCreate] = s.(bool)
		f[VDetails] = s.(bool)
		f[VUpdate] = s.(bool)
		f[Export] = s.(bool)
	}
	return f
}
