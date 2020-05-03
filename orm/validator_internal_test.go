package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_appendConfig(t *testing.T) {
	test := assert.New(t)

	// validator with no config
	v := &validator{}
	v.appendConfig("required")
	test.Equal("required", v.Config)
	v.appendConfig("numeric")
	test.Equal("required"+validatorSeparator+"numeric", v.Config)

	// same key is not added twice
	v.appendConfig("numeric")
	test.Equal("required"+validatorSeparator+"numeric", v.Config)

	// TODO this is wrong, numeric is only allowed once.
	v.appendConfig("numeric,listof=a b")
	test.Equal("required"+validatorSeparator+"numeric"+validatorSeparator+"listof=a b", v.Config)
}

func Test_validationKeys(t *testing.T) {
	test := assert.New(t)
	v := &validator{}
	test.Equal(4, len(v.validationKeys("required, omitempty,min=1,max=5")))
}

func Test_split(t *testing.T) {
	test := assert.New(t)
	test.True(split(rune(validatorSeparator[0])))
	test.True(split(rune(validatorOr[0])))
	test.False(split(rune(string("x")[0])))
}

func Test_skipByTag(t *testing.T) {
	test := assert.New(t)
	v := &validator{Config: validatorSkip}
	test.True(v.skipByTag())
	v = &validator{Config: ""}
	test.False(v.skipByTag())
}

func Test_addDBValidation(t *testing.T) {
	test := assert.New(t)
	car := car{}
	err := car.Init(&car)

	if test.NoError(err) {

		ID, err := car.scope.Field("ID")
		test.NoError(err)
		test.Equal("omitempty,numeric,min=0,max=4294967295", ID.Validator.Config)

		OwnerID, err := car.scope.Field("OwnerID")
		test.NoError(err)
		test.Equal("omitempty,numeric,min=0,max=4294967295", OwnerID.Validator.Config)

		Brand, err := car.scope.Field("Brand")
		test.NoError(err)
		test.Equal("oneof=BMW BMW2,required,max=100", Brand.Validator.Config)

		YearCheck, err := car.scope.Field("YearCheck")
		test.NoError(err)
		test.Equal("omitempty,numeric,min=-32768,max=32767", YearCheck.Validator.Config)

		CreatedAt, err := car.scope.Field("CreatedAt")
		test.NoError(err)
		test.Equal("", CreatedAt.Validator.Config) // no validation is set

		UpdatedAt, err := car.scope.Field("CreatedAt")
		test.NoError(err)
		test.Equal("", UpdatedAt.Validator.Config) // no validation is set
	}
}

func Test_isValid(t *testing.T) {
	test := assert.New(t)
	car := car{}
	err := car.Init(&car)
	test.NoError(err)

	car.Brand = "BMW"
	test.Nil(car.isValid())

	// max size 100
	car.Brand = "requiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequiredrequired"
	test.Error(car.isValid())

	// oneof BMW BMW2
	car.Brand = "BMW3"
	test.Error(car.isValid())
}
