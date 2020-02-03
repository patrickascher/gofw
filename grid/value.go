package grid

// value is holding values for different modes.
type value struct {
	grid    interface{}
	details interface{}
	create  interface{}
	update  interface{}

	g *Grid
}

// Value sets the given value to all mode variables.
func Value(val interface{}) *value {
	v := value{}
	v.set(val)
	return &v
}

// Grid sets the value only for the grid view
func (v *value) Grid(val interface{}) *value {
	v.grid = val
	return v
}

// Details sets the value only for the details view
func (v *value) Details(val interface{}) *value {
	v.details = val
	return v
}

// Create sets the value only for the create view
func (v *value) Create(val interface{}) *value {
	v.create = val
	return v
}

// Update sets the value only for the update view
func (v *value) Edit(val interface{}) *value {
	v.update = val
	return v
}

// valueWithGrid sets the given value to all mode variables and a ptr to grid
func valueWithGrid(val interface{}, g *Grid) *value {
	v := value{}
	v.set(val)
	v.g = g

	return &v
}

// set is a internal helper to set the value to all mode variables
func (v *value) set(val interface{}) {
	v.grid = val
	v.details = val
	v.create = val
	v.update = val
}

// setByValue updates the mode variables by the given value struct
func (v *value) setByValue(val *value) {
	v.grid = val.grid
	v.details = val.details
	v.create = val.create
	v.update = val.update
}

// get is a internal helper to get the value of the active grid mode
func (v *value) get() interface{} {

	switch v.g.Mode() {
	case ViewGrid:
		return v.grid
	case ViewDetails:
		return v.details
	case ViewCreate:
		return v.create
	case ViewEdit:
		return v.update
	}

	return nil
}

// getBool converts the result to a boolean
func (v *value) getInterface() interface{} {
	val := v.get()
	if val != nil {
		return val.(interface{})
	}
	return nil
}

// getBool converts the result to a boolean
func (v *value) getBool() bool {
	val := v.get()
	if val != nil {
		return val.(bool)
	}
	return false
}

// getInt converts the result to an integer
func (v *value) getInt() int {
	val := v.get()
	if val != nil {
		return val.(int)
	}
	return 0
}

// getString converts the result to a string
func (v *value) getString() string {
	val := v.get()
	if val != nil {
		return val.(string)
	}
	return ""
}
