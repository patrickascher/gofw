package orm

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestEagerLoading_setValue(t *testing.T) {

	type Profile struct {
		Name string
	}

	type User struct {
		Profile     Profile
		PtrProfile  *Profile
		Profiles    []Profile
		PtrProfiles []*Profile
	}

	p1 := Profile{Name: "Struct"}
	p2 := &Profile{Name: "Ptr"}
	p3 := Profile{Name: "Slice"}
	p4 := &Profile{Name: "SlicePtr"}

	user := &User{}
	setValue(reflect.ValueOf(user).Elem().FieldByName("Profile"), reflect.ValueOf(p1))
	setValue(reflect.ValueOf(user).Elem().FieldByName("PtrProfile"), reflect.ValueOf(p2).Elem())
	setValue(reflect.ValueOf(user).Elem().FieldByName("Profiles"), reflect.ValueOf(p3))
	setValue(reflect.ValueOf(user).Elem().FieldByName("PtrProfiles"), reflect.ValueOf(p4).Elem())

	assert.Equal(t, "Struct", user.Profile.Name)
	assert.Equal(t, "Ptr", user.PtrProfile.Name)
	assert.Equal(t, 1, len(user.Profiles))
	assert.Equal(t, "Slice", user.Profiles[0].Name)
	assert.Equal(t, 1, len(user.PtrProfiles))
	assert.Equal(t, "SlicePtr", user.PtrProfiles[0].Name)

}

func TestEagerLoading_inSlice(t *testing.T) {

	slice := []int{1, 2, 3, 4}

	assert.True(t, inSlice(1, slice))
	assert.True(t, inSlice(2, slice))
	assert.True(t, inSlice(3, slice))
	assert.True(t, inSlice(4, slice))
	assert.False(t, inSlice(5, slice))

}

func TestEagerLoading_isZeroOfUnderlyingType(t *testing.T) {
	assert.True(t, isZeroOfUnderlyingType(""))
	assert.False(t, isZeroOfUnderlyingType("test"))

	assert.True(t, isZeroOfUnderlyingType(0))
	assert.False(t, isZeroOfUnderlyingType(1))
}

func TestEagerLoading_initRelation(t *testing.T) {
	//GlobalBuilder, err := HelperCreateBuilder()
	cust := Customerfk{}
	err := cust.Initialize(&cust)
	assert.NoError(t, err)

	assert.True(t, cust.Info.Table() == nil)
	rel, err := initRelation(&cust, "Info")
	assert.NoError(t, err)
	assert.True(t, rel.Table() != nil)

	assert.True(t, cust.Orders == nil)
	rel, err = initRelation(&cust, "Orders")
	assert.NoError(t, err)
	assert.True(t, rel.Table() != nil)
	assert.Equal(t, 0, len(cust.Orders))

	// tabled does not exist
	custPtr := Customerptr{}
	err = custPtr.Initialize(&custPtr)
	assert.Error(t, err)

	assert.True(t, custPtr.Info == nil)
	rel, err = initRelation(&custPtr, "Info")
	assert.NoError(t, err)
	assert.True(t, rel.Table() != nil)

	custPtr.Info = &Contactfk{ID: 1}
	rel, err = initRelation(&custPtr, "Info")
	assert.NoError(t, err)
	assert.True(t, rel.Table() != nil)
	assert.True(t, custPtr.Info.Table() != nil)
	assert.True(t, custPtr.Info.ID == 1)

	assert.True(t, custPtr.Orders == nil)
	rel, err = initRelation(&custPtr, "Orders")
	assert.NoError(t, err)
	assert.True(t, rel.Table() != nil)
	assert.Equal(t, 0, len(cust.Orders))

}
