package controller

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestController struct {
	Controller
}

func (c *TestController) Get() {

}

// Test SetContext and Context
func TestController_SetContext(t *testing.T) {
	c := TestController{}
	ctx := &Context{}

	c.SetContext(ctx)

	assert.Equal(t, ctx, c.Context())
}

func TestController_isInitialize(t *testing.T) {
	c := TestController{}

	assert.Equal(t, false, c.isInitialized())
	c.name = "Test"
	assert.Equal(t, true, c.isInitialized())
}

func TestController_setSkipFuncChecks(t *testing.T) {
	c := TestController{}

	assert.Equal(t, false, c.skipMethodChecks)
	c.setSkipFuncChecks(true)
	assert.Equal(t, true, c.skipMethodChecks)
}

func TestController_getFunc(t *testing.T) {
	c := &TestController{}
	_, err := getFunc(c, "test")
	assert.Error(t, err)

	fn, err := getFunc(c, "Get")
	assert.NoError(t, err)
	assert.IsType(t, func() {}, fn)
}

func TestController_functionByPatternAndHTTPMethod(t *testing.T) {
	c := TestController{}
	c.caller = &c
	c.patternHTTPMethodStructFuncMapping = map[string]map[string]string{"/": {"GET": "Get"}}

	_, err := c.functionByPatternAndHTTPMethod("/", "POST")
	assert.Error(t, err)

	fn, err := c.functionByPatternAndHTTPMethod("/", "GET")
	assert.NoError(t, err)
	assert.IsType(t, func() {}, fn)
}

func TestController_copyController(t *testing.T) {

	c := Controller{}
	cc := &TestController{}
	c.patternHTTPMethodStructFuncMapping = map[string]map[string]string{"/": {"GET": "Get"}}
	c.caller = cc

	fn := copyController(&c)

	newC := fn()
	assert.Equal(t, map[string]string{"GET": "Get"}, newC.HTTPMethodsByPattern("/"))
}
