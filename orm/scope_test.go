package orm_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
)

func getModel() (orm.Interface, error) {
	car := &car{}
	return car, car.Init(car)
}

func TestScope_Builder(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.IsType(&sqlquery.Builder{}, model.Scope().Builder())
}

func TestScope_Model(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.IsType(&orm.Model{}, model.Scope().Model())
}

func TestScope_Parent(t *testing.T) {
	test := assert.New(t)

	car := &car{}
	err := car.Init(car)
	test.NoError(err)
	car.Owner = &owner{}

	// error because its already the root model
	_, err = car.Scope().Parent("")
	test.Error(err)

	// init owner
	err = car.Scope().InitRelation(car.Owner, "Owner")
	test.NoError(err)

	// check parent of owner
	c2, err := car.Owner.Scope().Parent(car.Scope().Name(true))
	test.NoError(err)
	test.Equal(car, c2.Scope().Caller())

	// check root parent of owner
	// TODO: create a model with more depth than car.owner
	c2, err = car.Owner.Scope().Parent("")
	test.NoError(err)
	test.Equal(car, c2.Scope().Caller())
}

func TestScope_SetBackReference(t *testing.T) {
	test := assert.New(t)

	car := &carBackRef{}
	err := car.Init(car)
	test.NoError(err)
	car.Owner = &ownerBackRef{}

	err = car.Scope().InitRelation(car.Owner, "Owner")
	test.NoError(err)

	r, err := car.Owner.Scope().Relation("Car", orm.Permission{})
	test.NoError(err)

	err = car.Owner.Scope().SetBackReference(r)
	test.NoError(err)

	// testing backref
	test.Equal(car, car.Owner.Car)

	// err - relation type name is not in parent model.
	relationFaker := r
	relationFaker.Type = reflect.ValueOf("").Type()
	err = car.Owner.Scope().SetBackReference(relationFaker)
	test.Error(err)
}

func Test_SetReflectValue(t *testing.T) {
	test := assert.New(t)

	type main struct {
		ID    int64
		IDint int

		Name     string
		NameNull orm.NullString
		IDNull   orm.NullInt
		Slice    []string
		SlicePtr []*string
		Owner    *owner
	}

	m := &main{}
	reflectM := reflect.ValueOf(m).Elem()

	// assigning String to String
	err := orm.SetReflectValue(reflectM.FieldByName("Name"), reflect.ValueOf("John Doe"))
	test.NoError(err)
	test.Equal("John Doe", m.Name)
	// assigning NullString to NullString
	err = orm.SetReflectValue(reflectM.FieldByName("Name"), reflect.ValueOf(orm.NewNullString("John Doe")))
	test.NoError(err)
	test.Equal("John Doe", m.Name)
	// assigning NullString to NullString
	err = orm.SetReflectValue(reflectM.FieldByName("NameNull"), reflect.ValueOf(orm.NewNullString("John Doe")))
	test.NoError(err)
	test.Equal("John Doe", m.NameNull.String)
	test.Equal(true, m.NameNull.Valid)
	// assigning NullString to NullString
	err = orm.SetReflectValue(reflectM.FieldByName("NameNull"), reflect.ValueOf("John Doe"))
	test.NoError(err)
	test.Equal("John Doe", m.NameNull.String)
	test.Equal(true, m.NameNull.Valid)

	// assigning INT to INT64
	err = orm.SetReflectValue(reflectM.FieldByName("ID"), reflect.ValueOf(1))
	test.NoError(err)
	test.Equal(1, orm.Int(m.ID))
	// assigning NullInt to INT64
	err = orm.SetReflectValue(reflectM.FieldByName("ID"), reflect.ValueOf(orm.NewNullInt(1)))
	test.NoError(err)
	test.Equal(1, orm.Int(m.ID))
	// assigning NullInt to NullInt
	err = orm.SetReflectValue(reflectM.FieldByName("IDNull"), reflect.ValueOf(orm.NewNullInt(1)))
	test.NoError(err)
	test.Equal(true, m.IDNull.Valid)
	test.Equal(1, orm.Int(m.IDNull.Int64))
	// assigning INT to NullInt
	err = orm.SetReflectValue(reflectM.FieldByName("IDNull"), reflect.ValueOf(1))
	test.NoError(err)
	test.Equal(true, m.IDNull.Valid)
	test.Equal(1, orm.Int(m.IDNull.Int64))
	// assigning INT to INT32
	err = orm.SetReflectValue(reflectM.FieldByName("IDint"), reflect.ValueOf(1))
	test.NoError(err)
	test.Equal(1, orm.Int(m.ID))
	// assigning NullInt to INT32
	err = orm.SetReflectValue(reflectM.FieldByName("IDint"), reflect.ValueOf(orm.NewNullInt(1)))
	test.NoError(err)
	test.Equal(1, orm.Int(m.ID))

	// assigning String to Slice
	err = orm.SetReflectValue(reflectM.FieldByName("Slice"), reflect.ValueOf("John"))
	test.NoError(err)
	test.Equal([]string{"John"}, m.Slice)
	// assigning *String to Slice*String
	ptrString := "John"
	err = orm.SetReflectValue(reflectM.FieldByName("SlicePtr"), reflect.ValueOf(&ptrString))
	test.NoError(err)
	test.Equal([]*string{&ptrString}, m.SlicePtr)

	// assigning ptr to ptr
	err = orm.SetReflectValue(reflectM.FieldByName("Owner"), reflect.ValueOf(&owner{Name: orm.NewNullString("John Doe")}))
	test.NoError(err)
	test.Equal(true, m.Owner.Name.Valid)
	test.Equal("John Doe", m.Owner.Name.String)
	// assigning addressable ptr
	owners := []owner{{Name: orm.NewNullString("John Doe")}}
	err = orm.SetReflectValue(reflectM.FieldByName("Owner"), reflect.ValueOf(owners).Index(0))
	test.NoError(err)
	test.Equal(true, m.Owner.Name.Valid)
	test.Equal("John Doe", m.Owner.Name.String)
}

func TestScope_Caller(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.IsType(&car{}, model.Scope().Caller())
}

func TestScope_Name(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal("orm_test.car", model.Scope().Name(true))
	test.Equal("Car", model.Scope().Name(false))
}

func TestScope_TableName(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal("orm_test.cars", model.Scope().TableName())
}

func TestScope_Columns(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal([]string{"id", "owner_id", "brand", "created_at", "updated_at"}, model.Scope().Columns(orm.Permission{}, false))
	test.Equal([]string{"!(id) AS `id`", "owner_id", "brand", "created_at", "updated_at"}, model.Scope().Columns(orm.Permission{}, true))

}

func TestScope_PrimaryKeys(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal(1, len(model.Scope().PrimaryKeys()))
	test.Equal("ID", model.Scope().PrimaryKeys()[0].Name)

}

func TestScope_PrimaryKeysFieldName(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal(1, len(model.Scope().PrimaryKeysFieldName()))
	test.Equal([]string{"ID"}, model.Scope().PrimaryKeysFieldName())
}

func TestScope_PrimariesSet(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal(false, model.Scope().PrimariesSet())
	x := model.Scope().Caller().(*car)
	x.ID = orm.NewNullInt(1)
	test.Equal(true, model.Scope().PrimariesSet())
}

func TestScope_IsEmpty(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal(true, model.Scope().IsEmpty(orm.Permission{Write: true}))
	x := model.(*car)
	x.ID = orm.NewNullInt(1)
	test.Equal(false, model.Scope().IsEmpty(orm.Permission{Write: true}))

	model, err = getModel()
	test.NoError(err)
	x = model.(*car)
	x.Owner = &owner{}
	test.Equal(true, model.Scope().IsEmpty(orm.Permission{}))

	x.Owner = &owner{Name: orm.NewNullString("John Doe")}
	test.Equal(false, model.Scope().IsEmpty(orm.Permission{}))
}

func TestScope_CallerField(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	car := model.(*car)
	car.ID = orm.NewNullInt(1)

	test.Equal(orm.NewNullInt(1), model.Scope().CallerField("ID").Interface())
}

func TestScope_Field(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	car := model.(*car)
	car.ID = orm.NewNullInt(1)

	f, err := model.Scope().Field("ID")
	test.NoError(err)
	test.Equal("ID", f.Name)

	f, err = model.Scope().Field("IDx")
	test.Error(err)
}

func TestScope_Fields(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	f := model.Scope().Fields(orm.Permission{})
	test.Equal(5, len(f))

	f = model.Scope().Fields(orm.Permission{Read: true})
	test.Equal(2, len(f)) // Brand has no read permission
}

func TestScope_Relation(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	car := model.(*car)
	car.ID = orm.NewNullInt(1)

	r, err := model.Scope().Relation("Owner", orm.Permission{})
	test.NoError(err)
	test.Equal("Owner", r.Field)

	r, err = model.Scope().Relation("Wheel", orm.Permission{Read: true})
	test.Error(err)
}

func TestScope_Relations(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	f := model.Scope().Relations(orm.Permission{})
	test.Equal(5, len(f))

	f = model.Scope().Relations(orm.Permission{Read: true})
	test.Equal(4, len(f)) // Wheel has no read permission
}

func TestScope_IsPolymorphic(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	r, err := model.Scope().Relation("Owner", orm.Permission{})
	test.NoError(err)
	test.Equal("Owner", r.Field)
	test.False(model.Scope().IsPolymorphic(r))

	r, err = model.Scope().Relation("Radio", orm.Permission{})
	test.NoError(err)
	test.Equal("Radio", r.Field)
	test.True(model.Scope().IsPolymorphic(r))
}

func TestScope_CachedModel(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	m, err := model.Scope().CachedModel("orm_test.car")
	test.NoError(err)
	test.Equal("Car", m.Scope().Name(false))

	m, err = model.Scope().CachedModel("orm_test.car2")
	test.Error(err)
	test.Equal(nil, m)
}

func TestScope_ScanValues(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	// ID, OwnerID, Brand, CreatedAt, UpdatedAt
	test.Equal(5, len(model.Scope().ScanValues(orm.Permission{Write: true})))
}

func TestScope_NewScopeFromType(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)
	f, err := model.Scope().Field("Brand")
	test.NoError(err)
	f.Permission.Write = false
	f.Permission.Read = false

	s, err := model.Scope().NewScopeFromType(reflect.TypeOf(car{}))
	test.NoError(err)
	test.Equal("Car", s.Name(false))
	fcopy, err := model.Scope().Field("Brand")
	test.NoError(err)
	fcopy.Permission.Write = false
	fcopy.Permission.Read = false

	test.Equal(f.Permission, fcopy.Permission)

	//s, err = model.Scope().NewScopeFromType(reflect.TypeOf(owner{}))
	//test.Error(err)
}

func TestScope_TimeFields(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	test.Equal([]string{"CreatedAt", "UpdatedAt"}, model.Scope().TimeFields(orm.Permission{}))
}

func TestScope_InitCallerRelation(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	i, err := model.Scope().InitCallerRelation("Owner", false) //* nil
	test.NoError(err)
	test.IsType(&owner{}, i)

	i, err = model.Scope().InitCallerRelation("Radio", false) // struct
	test.NoError(err)
	test.IsType(&radio{}, i)

	i, err = model.Scope().InitCallerRelation("Wheels", false) // slice
	test.NoError(err)
	test.IsType(&wheel{}, i)

	i, err = model.Scope().InitCallerRelation("Driver", false) // slice*
	test.NoError(err)
	test.IsType(&driver{}, i)

	i, err = model.Scope().InitCallerRelation("notExisting", false)
	test.Error(err)
	test.IsType(nil, i)
}

func TestScope_InitRelation(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	car := model.(*car)

	// err car.Owner is nil.
	err = model.Scope().InitRelation(car.Owner, "Owner")
	test.Error(err)

	// init the owner model completely.
	owner := &owner{}
	err = model.Scope().InitRelation(owner, "Owner")
	test.NoError(err)
	test.Equal("Owner", owner.Scope().Name(false))

	// already init, only change the caller.
	err = model.Scope().InitRelation(owner, "Owner")
	test.NoError(err)
	test.Equal(owner, owner.Scope().Caller())
}

func TestScope_checkLoopMap(t *testing.T) {
	test := assert.New(t)
	r := &Role{}
	err := r.Init(r)
	test.NoError(err)

	r.Name = orm.NewNullString("Admin")
	r.Roles = append(r.Roles, &Role{Name: orm.NewNullString("Writer"), Roles: []*Role{r}})
	err = r.Create()
	test.NoError(err)

	x := &Role{}
	err = x.Init(x)
	test.NoError(err)
	err = x.First(sqlquery.NewCondition().Where("id=?", r.ID))
	test.Error(err)
	test.Equal("orm: 🎉 congratulation you created an infinity loop", err.Error())
}

func TestScope_addParentWbList(t *testing.T) {

	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	car := model.(*car)
	car.Owner = &owner{}
	car.SetWBList(orm.WHITELIST, "ID", "Owner.Name")
	err = car.Scope().InitRelation(car.Owner, "Owner")
	test.NoError(err)

	p, f := car.Owner.WBList()
	test.Equal(orm.WHITELIST, p)
	test.Equal([]string{"Name"}, f)

	// self reference
	r := &Role{}
	err = r.Init(r)
	test.NoError(err)
	r.SetWBList(orm.BLACKLIST, "Name")
	r.Roles = append(r.Roles, &Role{})
	err = r.Scope().InitRelation(r.Roles[0], "Roles")
	test.NoError(err)

	p, f = r.Roles[0].WBList()
	test.Equal(orm.BLACKLIST, p)
	test.Equal([]string{"Name"}, f)

}

func TestScope_ChangedValue(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	// append changed value
	cv := orm.ChangedValue{Field: "Name"}
	model.Scope().AppendChangedValue(cv)

	// get named changed value
	c := model.Scope().ChangedValueByFieldName("Name")
	test.NotNil(c)

	// set a changed value list
	cvs := []orm.ChangedValue{{Field: "ID"}}
	model.Scope().SetChangedValues(cvs)
	c = model.Scope().ChangedValueByFieldName("ID")
	test.NotNil(c)
}

func TestScope_EqualWith(t *testing.T) {
	test := assert.New(t)

	model, err := getModel()
	test.NoError(err)

	model2, err := getModel()
	test.NoError(err)

	m1 := model.(*car)
	m2 := model2.(*car)

	err = m1.Init(m1)
	test.NoError(err)
	err = m2.Init(m2)
	test.NoError(err)

	// no changes
	cv, err := m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Nil(cv)

	// time fields are excluded
	cratedAt := orm.NewNullTime(time.Now())
	updatedAt := orm.NewNullTime(time.Now())
	m1.CreatedAt = &cratedAt
	m1.UpdatedAt = &updatedAt
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Nil(cv)

	// normal fields test
	m1.ID = orm.NewNullInt(1)
	m1.Brand = "BMW"
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	// ID
	test.Equal(3, len(cv))
	test.Equal("ID", cv[0].Field)
	test.Equal(orm.UPDATE, cv[0].Operation)
	test.Equal(orm.NewNullInt(1), cv[0].NewV)
	test.Equal(orm.NullInt{}, cv[0].OldV)
	// Brand
	test.Equal("Brand", cv[1].Field)
	test.Equal(orm.UPDATE, cv[1].Operation)
	test.Equal("BMW", cv[1].NewV)
	test.Equal("", cv[1].OldV)
	// UpdatedAt - is added as soon as one field gets added and the field exists in the db
	test.Equal("UpdatedAt", cv[2].Field)
	test.Equal(orm.UPDATE, cv[2].Operation)
	test.Nil(cv[2].NewV)
	test.Nil(cv[2].OldV)

	// relation owner belongsTo & hasOne
	// new Created
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Owner = &owner{Name: orm.NewNullString("John Doe")}
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.CREATE, cv[0].Operation)
	test.Equal("Owner", cv[0].Field)
	test.Equal(2, len(cv[0].ChangedValue))
	test.Equal("Name", cv[0].ChangedValue[0].Field)
	test.Equal("UpdatedAt", cv[0].ChangedValue[1].Field)
	// updated (IDs are the same)
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Owner = &owner{ID: 1, Name: orm.NewNullString("John Doe2")}
	m2.Owner = &owner{ID: 1, Name: orm.NewNullString("John Doe")}
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.UPDATE, cv[0].Operation)
	test.Equal("Owner", cv[0].Field)
	test.Equal(2, len(cv[0].ChangedValue))
	test.Equal("Name", cv[0].ChangedValue[0].Field)
	test.Equal(orm.NewNullString("John Doe2"), cv[0].ChangedValue[0].NewV)
	test.Equal(orm.NewNullString("John Doe"), cv[0].ChangedValue[0].OldV)
	test.Equal("UpdatedAt", cv[0].ChangedValue[1].Field)
	// create (IDs are the not the same)
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Owner = &owner{ID: 2, Name: orm.NewNullString("John Doe2")}
	m2.Owner = &owner{ID: 1, Name: orm.NewNullString("John Doe")}
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.CREATE, cv[0].Operation)
	test.Equal("Owner", cv[0].Field)
	test.Equal(3, len(cv[0].ChangedValue))
	test.Equal("ID", cv[0].ChangedValue[0].Field)
	test.Equal(2, cv[0].ChangedValue[0].NewV)
	test.Equal(1, cv[0].ChangedValue[0].OldV)
	test.Equal("Name", cv[0].ChangedValue[1].Field)
	test.Equal(orm.NewNullString("John Doe2"), cv[0].ChangedValue[1].NewV)
	test.Equal(orm.NewNullString("John Doe"), cv[0].ChangedValue[1].OldV)
	test.Equal("UpdatedAt", cv[0].ChangedValue[2].Field)
	// delete (object is empty)
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Owner = &owner{}
	m2.Owner = &owner{ID: 1, Name: orm.NewNullString("John Doe")}
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.DELETE, cv[0].Operation)
	// ID, Name, UpdatedAt
	test.Equal(3, len(cv[0].ChangedValue))

	// relation hasMany/many2many
	// no wheels added yet len = 0
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)

	// added new wheels, non were existing
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Wheels = append(m1.Wheels, wheel{Brand: orm.NewNullString("Goodyear")})
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.CREATE, cv[0].Operation)
	test.Equal("Wheels", cv[0].Field)
	test.Equal(0, len(cv[0].ChangedValue))

	// there were wheels before but not anymore
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m2.Wheels = append(m2.Wheels, wheel{Brand: orm.NewNullString("Goodyear")})
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.DELETE, cv[0].Operation)
	test.Equal("Wheels", cv[0].Field)
	test.Equal(0, len(cv[0].ChangedValue))

	// There were wheels before, but user added a new list
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Wheels = append(m1.Wheels, wheel{ID: 88, Brand: orm.NewNullString("Pirelli")}, wheel{ID: 88, Brand: orm.NewNullString("Pirelli2")})
	m2.Wheels = append(m2.Wheels, wheel{ID: 199, Brand: orm.NewNullString("Goodyear")}, wheel{ID: 200, Brand: orm.NewNullString("Goodyear2")})
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)
	test.Equal(1, len(cv))
	test.Equal(orm.UPDATE, cv[0].Operation)
	test.Equal("Wheels", cv[0].Field)
	test.Equal(4, len(cv[0].ChangedValue))
	test.Equal("Wheels", cv[0].ChangedValue[0].Field)
	test.Equal(orm.CREATE, cv[0].ChangedValue[0].Operation)
	test.Equal(0, cv[0].ChangedValue[0].Index) // add index of new slice

	test.Equal("Wheels", cv[0].ChangedValue[1].Field)
	test.Equal(orm.CREATE, cv[0].ChangedValue[1].Operation)
	test.Equal(1, cv[0].ChangedValue[1].Index) // add index of new slice

	test.Equal("Wheels", cv[0].ChangedValue[2].Field)
	test.Equal(orm.DELETE, cv[0].ChangedValue[2].Operation)
	test.Equal(199, cv[0].ChangedValue[2].Index) // index is representing the ID

	test.Equal("Wheels", cv[0].ChangedValue[3].Field)
	test.Equal(orm.DELETE, cv[0].ChangedValue[3].Operation)
	test.Equal(200, cv[0].ChangedValue[3].Index) // index is representing the ID

	// The same ID exists in the new list and old list, the rest is getting created or deleted.
	m1, m2 = &car{}, &car{}
	err, err = m1.Init(m1), m2.Init(m2)
	test.NoError(err)
	m1.Wheels = append(m1.Wheels, wheel{ID: 1, Brand: orm.NewNullString("Pirelli")}, wheel{Brand: orm.NewNullString("Pirelli3")}, wheel{ID: 88, Brand: orm.NewNullString("Pirelli2")})
	m2.Wheels = append(m2.Wheels, wheel{ID: 1, Brand: orm.NewNullString("Goodyear")}, wheel{Brand: orm.NewNullString("Goodyear2")}, wheel{ID: 200, Brand: orm.NewNullString("Goodyear2")})
	cv, err = m1.Scope().EqualWith(m2)
	test.NoError(err)

	test.Equal(1, len(cv))
	test.Equal(orm.UPDATE, cv[0].Operation)
	test.Equal("Wheels", cv[0].Field)
	test.Equal(5, len(cv[0].ChangedValue))

	test.Equal(orm.UPDATE, cv[0].ChangedValue[0].Operation)
	test.Equal("Wheels", cv[0].ChangedValue[0].Field)
	test.Equal(2, len(cv[0].ChangedValue[0].ChangedValue))
	test.Equal("Brand", cv[0].ChangedValue[0].ChangedValue[0].Field)
	test.Equal(orm.NewNullString("Pirelli"), cv[0].ChangedValue[0].ChangedValue[0].NewV)
	test.Equal(orm.NewNullString("Goodyear"), cv[0].ChangedValue[0].ChangedValue[0].OldV)
	test.Equal("UpdatedAt", cv[0].ChangedValue[0].ChangedValue[1].Field)

	test.Equal(orm.CREATE, cv[0].ChangedValue[1].Operation)
	test.Equal("Wheels", cv[0].ChangedValue[1].Field)
	test.Equal(1, cv[0].ChangedValue[1].Index)
	test.Equal(orm.CREATE, cv[0].ChangedValue[2].Operation)
	test.Equal("Wheels", cv[0].ChangedValue[2].Field)
	test.Equal(2, cv[0].ChangedValue[2].Index)

	test.Equal(orm.DELETE, cv[0].ChangedValue[3].Operation)
	test.Equal("Wheels", cv[0].ChangedValue[3].Field)
	test.Equal(0, cv[0].ChangedValue[3].Index) // TODO index - no ID was found an set? throw error?
	test.Equal(orm.DELETE, cv[0].ChangedValue[4].Operation)
	test.Equal("Wheels", cv[0].ChangedValue[4].Field)
	test.Equal(200, cv[0].ChangedValue[4].Index)
}
