package orm

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrParent              = errors.New("orm: no parent exists")
	ErrCallbackArguments   = errors.New("orm: the callback %#v method must have one argument of type *Callback")
	ErrCallbackReturnValue = errors.New("orm: the callback %#v method must return an error type")
)

type Parent struct {
	caller Interface
}

func (p *Parent) Model() Interface {
	return p.caller
}

func (p *Parent) ResultSet() interface{} {
	return p.caller.resultSet()
}

// TODO create Pkeys
type Relation struct {
	field reflect.StructField // set on call (setParent)
}

func (r *Relation) Type() string {
	return "" //r.caller
}

func (r *Relation) Field() reflect.StructField {
	return r.field
}

type Callback struct {
	caller Interface
	mode   string // set on call, if not set use default

	callbacks []string // check if callbacks exist
	rel       *Relation
}

func (c *Callback) setRelField(r reflect.StructField) {
	c.rel.field = r
}
func (c *Callback) ResultSet() interface{} {
	return c.caller.resultSet()
}

func (c *Callback) Parent() (*Parent, error) {
	if c.caller.parent() != nil {
		p := &Parent{caller: c.caller.parent()}
		return p, nil
	}
	return nil, ErrParent
}

func (c *Callback) Mode() string {
	return c.mode
}

func (c *Callback) Relation() *Relation {
	return c.rel
}

func (c *Callback) setCaller(m Interface) {
	c.caller = m
}

func (c *Callback) setMode(m string) {
	c.mode = m
}

func (c *Callback) call(method string) error {

	r := reflect.ValueOf(c.caller)
	m := r.MethodByName(method)

	if m.IsValid() {
		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(c)

		e := m.Call(in)
		if !e[0].IsNil() {
			return e[0].Interface().(error)
		}

		return nil
	}
	return nil
}

func (c *Callback) callIfExists(mode string, before bool) error {

	// exit if no callbacks exist
	if len(c.callbacks) == 0 || c.caller.disableCallback() {
		return nil
	}

	// setting custom mode if exists
	if c.mode == "" {
		c.mode = mode
	}

	// define callbacks
	globalCallback := CallbackBefore
	specificCallback := globalCallback + c.mode
	if before == false {
		globalCallback = CallbackAfter
		specificCallback = globalCallback + c.mode
	}

	// calling specific callback if exists
	if stringInSlice(specificCallback, c.callbacks) {
		return c.call(specificCallback)
	}

	// calling global callback if exists
	if stringInSlice(globalCallback, c.callbacks) {
		return c.call(globalCallback)
	}

	// callback does not exist
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (c *Callback) addCallbacks() error {
	ormCallbacks := []string{CallbackBefore, CallbackAfter, CallbackBefore + "First", CallbackAfter + "First", CallbackBefore + "All", CallbackAfter + "All", CallbackBefore + "Create", CallbackAfter + "Create", CallbackBefore + "Update", CallbackAfter + "Update", CallbackBefore + "Delete", CallbackAfter + "Delete"}
	for _, cbk := range ormCallbacks {
		method := reflect.ValueOf(c.caller).MethodByName(cbk)
		if method.IsValid() {

			methodName := structName(c.caller, true) + ":" + cbk

			// checking the method arguments
			numIn := method.Type().NumIn()
			if numIn != 1 {
				return fmt.Errorf(fmt.Sprintf(ErrCallbackArguments.Error(), methodName))
			}
			for i := 0; i < numIn; i++ {
				argType := method.Type().In(i)
				if argType != reflect.TypeOf(&Callback{}) {
					return fmt.Errorf(fmt.Sprintf(ErrCallbackArguments.Error(), methodName))
				}
			}

			// checking the method return value
			numOut := method.Type().NumOut()
			if numOut != 1 {
				return fmt.Errorf(fmt.Sprintf(ErrCallbackReturnValue.Error(), methodName))
			}
			for i := 0; i < numOut; i++ {
				rvType := method.Type().Out(i)
				if rvType.String() != "error" {
					return fmt.Errorf(fmt.Sprintf(ErrCallbackReturnValue.Error(), methodName))
				}
			}

			c.callbacks = append(c.callbacks, cbk)
		}
	}

	return nil
}

func NewCallback(caller Interface) (*Callback, error) {
	cbk := &Callback{}
	cbk.setCaller(caller)
	cbk.rel = &Relation{}
	err := cbk.addCallbacks()
	return cbk, err
}
