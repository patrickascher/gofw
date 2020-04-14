package orm2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestModel_foreignKey(t *testing.T) {
	test := assert.New(t)
	c := &car{}
	err := c.Init(c)
	test.NoError(err)

	f, err := c.foreignKey("Brand")
	test.NoError(err)

	test.Equal("Brand", f.Name)
}

func TestModel_polymorphicRestriction(t *testing.T) {
	test := assert.New(t)

	// no err
	err := polymorphicRestriction(map[string]string{"a": "b"}, "Test", HasOne)
	test.NoError(err)

	// err tagPolymorphic is set
	err = polymorphicRestriction(map[string]string{tagPolymorphic: "foo"}, "Test", HasOne)
	test.Error(err)

	// err tagPolymorphicValue is set
	err = polymorphicRestriction(map[string]string{tagPolymorphicValue: "foo"}, "Test", HasOne)
	test.Error(err)
}

func TestModel_polymorphic(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c) // needed for the m.name
	test.NoError(err)

	r := &radio{}
	err = r.Init(r)
	test.NoError(err)

	// no err
	p, err := c.polymorphic(map[string]string{tagPolymorphic: "Car"}, r)
	test.NoError(err)
	test.Equal("Car", p.Value)
	test.Equal("CarID", p.Field.Name)
	test.Equal("CarType", p.Type.Name)

	// err {name}ID does not exist
	cNoId := &ComponentNoID{}
	err = cNoId.Init(cNoId)
	test.NoError(err)
	p, err = c.polymorphic(map[string]string{tagPolymorphic: "Car"}, cNoId)
	test.Error(err)
	test.Equal(Polymorphic{}, p)

	// err {name}Type does nit exist
	cNoType := &ComponentNoType{}
	err = cNoType.Init(cNoType)
	test.NoError(err)
	p, err = c.polymorphic(map[string]string{tagPolymorphic: "Car"}, cNoType)
	test.Error(err)
	test.Equal(Polymorphic{}, p)

}

func TestModel_associationForeignKey(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c) // needed for the m.name
	test.NoError(err)

	r := &radio{}
	err = r.Init(r)
	test.NoError(err)

	//ok no tag - CarID exists
	f, err := c.associationForeignKey("", r)
	test.NoError(err)
	test.Equal("CarID", f.Name)

	//ok tag - CarType exists
	f, err = c.associationForeignKey("CarType", r)
	test.NoError(err)
	test.Equal("CarType", f.Name)

	//err tag - CarType exists
	f, err = c.associationForeignKey("Foo", r)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "Foo", r.name), err.Error())
	test.Equal(Field{}, f)
}

func TestModel_findByName(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c) // needed for the m.name
	test.NoError(err)

	scope := NewScopeFromInterface(c)
	//ok no tag - CarID exists
	f, err := scope.Field("ID")
	test.NoError(err)
	test.Equal("ID", f.Name)

	//err tag - CarType exists
	f, err = scope.Field("Foo")
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "Foo", c.name), err.Error())

	test.Equal(Field{}, f)
}

func TestModel_joinTable(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	test.NoError(err)

	d := &driver{}
	err = d.Init(d)
	test.NoError(err)

	// ok
	j, err := joinTable(c, d, nil, false)
	test.NoError(err)
	test.Equal(JoinTable{Name: "car_drivers", ForeignKey: "car_id", AssociationForeignKey: "driver_id"}, j)

	// err both join table columns does not exist
	tags := map[string]string{tagJoinTable: "car_drivers", tagJoinForeignKey: "car_id2", tagJoinAssociationForeignKey: "driver_id2"}
	j, err = joinTable(c, d, tags, false)
	test.Error(err)

	// err one join table columns does not exist
	tags = map[string]string{tagJoinTable: "car_drivers", tagJoinForeignKey: "car_id", tagJoinAssociationForeignKey: "driver_id2"}
	j, err = joinTable(c, d, tags, false)
	test.Error(err)

	// err join table does not exist
	tags = map[string]string{tagJoinTable: "car_driver", tagJoinForeignKey: "car_id", tagJoinAssociationForeignKey: "driver_id"}
	j, err = joinTable(c, d, tags, false)
	test.Error(err)

	// err join table columns does not exist
	tags = map[string]string{tagJoinTable: "custom_car_drivers", tagJoinForeignKey: "custom_car_id", tagJoinAssociationForeignKey: "custom_driver_id"}
	j, err = joinTable(c, d, tags, false)
	test.NoError(err)
	test.Equal(JoinTable{Name: "custom_car_drivers", ForeignKey: "custom_car_id", AssociationForeignKey: "custom_driver_id"}, j)

	// self reference
	r := &Role{}
	err = r.Init(r)
	test.NoError(err)
	j, err = joinTable(r, r, nil, true)
	test.NoError(err)
	test.Equal(JoinTable{Name: "role_roles", ForeignKey: "role_id", AssociationForeignKey: "child_id"}, j)

	// self reference custom tags
	j, err = joinTable(r, r, tags, true)
	test.NoError(err)
	test.Equal(JoinTable{Name: "custom_car_drivers", ForeignKey: "custom_car_id", AssociationForeignKey: "custom_driver_id"}, j)
}

func TestModel_isTagRelationAllowed(t *testing.T) {
	test := assert.New(t)

	// struct
	v := reflect.TypeOf(car{})
	f, exist := v.FieldByName("Radio")
	test.True(exist)
	test.True(isTagRelationAllowed(f, HasOne))
	test.True(isTagRelationAllowed(f, BelongsTo))
	test.False(isTagRelationAllowed(f, HasMany))
	test.False(isTagRelationAllowed(f, ManyToMany))

	// ptr to struct
	f, exist = v.FieldByName("Owner")
	test.True(exist)
	test.True(isTagRelationAllowed(f, HasOne))
	test.True(isTagRelationAllowed(f, BelongsTo))
	test.False(isTagRelationAllowed(f, HasMany))
	test.False(isTagRelationAllowed(f, ManyToMany))

	// slice
	f, exist = v.FieldByName("Driver")
	test.True(exist)
	test.False(isTagRelationAllowed(f, HasOne))
	test.False(isTagRelationAllowed(f, BelongsTo))
	test.True(isTagRelationAllowed(f, HasMany))
	test.True(isTagRelationAllowed(f, ManyToMany))

	// slice to ptr
	f, exist = v.FieldByName("CustomDriver")
	test.True(exist)
	test.False(isTagRelationAllowed(f, HasOne))
	test.False(isTagRelationAllowed(f, BelongsTo))
	test.True(isTagRelationAllowed(f, HasMany))
	test.True(isTagRelationAllowed(f, ManyToMany))

	// string
	f, exist = v.FieldByName("Brand")
	test.True(exist)
	test.False(isTagRelationAllowed(f, HasOne))
	test.False(isTagRelationAllowed(f, BelongsTo))
	test.False(isTagRelationAllowed(f, HasMany))
	test.False(isTagRelationAllowed(f, ManyToMany))
}

func TestMode_relationKind(t *testing.T) {
	test := assert.New(t)

	c := &car{}
	c.name = "orm2.car"

	r := &Role{}
	r.name = "orm2.Role"

	var tests = []struct {
		Tag      map[string]string
		Field    string
		Expected string
		Error    bool
	}{
		// *Struct
		{Tag: nil, Field: "Owner", Expected: HasOne, Error: false},
		{Tag: map[string]string{tagRelation: BelongsTo}, Field: "Owner", Expected: BelongsTo, Error: false},
		{Tag: map[string]string{tagRelation: HasOne}, Field: "Owner", Expected: HasOne, Error: false},
		{Tag: map[string]string{tagRelation: HasMany}, Field: "Owner", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: ManyToMany}, Field: "Owner", Expected: "", Error: true},
		// Struct
		{Tag: nil, Field: "Radio", Expected: HasOne, Error: false},
		{Tag: map[string]string{tagRelation: BelongsTo}, Field: "Radio", Expected: BelongsTo, Error: false},
		{Tag: map[string]string{tagRelation: HasOne}, Field: "Radio", Expected: HasOne, Error: false},
		{Tag: map[string]string{tagRelation: HasMany}, Field: "Radio", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: ManyToMany}, Field: "Radio", Expected: "", Error: true},
		// Slice
		{Tag: nil, Field: "Driver", Expected: HasMany, Error: false},
		{Tag: map[string]string{tagRelation: BelongsTo}, Field: "Driver", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: HasOne}, Field: "Driver", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: HasMany}, Field: "Driver", Expected: HasMany, Error: false},
		{Tag: map[string]string{tagRelation: ManyToMany}, Field: "Driver", Expected: ManyToMany, Error: false},
		// Slice*
		{Tag: nil, Field: "CustomDriver", Expected: HasMany, Error: false},
		{Tag: map[string]string{tagRelation: BelongsTo}, Field: "CustomDriver", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: HasOne}, Field: "CustomDriver", Expected: "", Error: true},
		{Tag: map[string]string{tagRelation: HasMany}, Field: "CustomDriver", Expected: HasMany, Error: false},
		{Tag: map[string]string{tagRelation: ManyToMany}, Field: "CustomDriver", Expected: ManyToMany, Error: false},
		// string type
		{Tag: nil, Field: "Brand", Expected: "", Error: true},
		// role self-reference
		{Tag: nil, Field: "Roles", Expected: ManyToMany, Error: false},
	}

	for _, tt := range tests {
		t.Run(tt.Field, func(t *testing.T) {
			var rel string
			var err error

			if tt.Field == "Roles" {
				v := reflect.TypeOf(Role{})
				f, exist := v.FieldByName(tt.Field)
				test.True(exist)
				rel, err = r.relationKind(tt.Tag, f)
			} else {
				v := reflect.TypeOf(car{})
				f, exist := v.FieldByName(tt.Field)
				test.True(exist)
				rel, err = c.relationKind(tt.Tag, f)
			}

			if tt.Error {
				test.Error(err)
			} else {
				test.NoError(err)
			}
			test.Equal(tt.Expected, rel)
		})
	}
}

func TestModel_implementsInterface(t *testing.T) {
	test := assert.New(t)

	v := reflect.TypeOf(car{})

	// ok - Owner hasOne implements orm interface
	f, exist := v.FieldByName("Owner")
	test.True(exist)
	test.True(implementsInterface(f))

	// err - type string
	f, exist = v.FieldByName("Brand")
	test.True(exist)
	test.False(implementsInterface(f))
}

func TestModel_newValueInstanceFromType(t *testing.T) {
	test := assert.New(t)
	v := reflect.TypeOf(car{})

	// ptr struct
	f, exist := v.FieldByName("Owner")
	test.True(exist)
	val := newValueInstanceFromType(f.Type)
	test.Equal(f.Type.String(), val.Type().String())

	// struct
	f, exist = v.FieldByName("Radio")
	test.True(exist)
	val = newValueInstanceFromType(f.Type)
	test.Equal(f.Type.String(), val.Type().String())

	// slice
	f, exist = v.FieldByName("Driver")
	test.True(exist)
	val = newValueInstanceFromType(f.Type)
	test.Equal(f.Type.Elem().String(), val.Type().String())

	// slice*
	f, exist = v.FieldByName("CustomDriver")
	test.True(exist)
	val = newValueInstanceFromType(f.Type)
	test.Equal(f.Type.Elem().Elem().String(), val.Type().String())
}

func TestModel_initializeModelByValue(t *testing.T) {
	test := assert.New(t)
	c := car{}
	c.setCaller(&c)
	v := reflect.TypeOf(car{})

	// ptr struct
	f, exist := v.FieldByName("Owner")
	test.True(exist)
	val := newValueInstanceFromType(f.Type)
	i, err := c.initializeModelByValue(val)
	test.NoError(err)
	test.NotNil(i)

	// err cache is not set
	c1 := carCacheErr{}
	c1.setCaller(&c1)
	v = reflect.TypeOf(carCacheErr{})
	val = newValueInstanceFromType(f.Type)
	i, err = c1.initializeModelByValue(val)
	test.Error(err)
	test.Equal(errNoCache.Error(), err.Error())
}

// test create Fields
func TestModel_createRelations(t *testing.T) {

	test := assert.New(t)

	c := &car{}
	err := c.Init(c)
	assert.NoError(t, err)

	// checking if all fields exist (added createdAt, ignored Owner relation)
	if test.Equal(6, len(c.relations)) {
		// table driven tests
		var tests = []struct {
			Kind                  string
			Field                 string
			ForeignKey            Field
			AssociationForeignKey Field
			SelfReference         bool
			JoinTable             JoinTable
			Polymorphic           Polymorphic
			Custom                bool
		}{
			{Kind: BelongsTo, Field: "Owner", ForeignKey: Field{Name: "OwnerID"}, AssociationForeignKey: Field{Name: "ID"}, SelfReference: false, JoinTable: JoinTable{}, Polymorphic: Polymorphic{}},
			{Kind: ManyToMany, Field: "Driver", ForeignKey: Field{Name: "ID"}, AssociationForeignKey: Field{Name: "ID"}, SelfReference: false, JoinTable: JoinTable{Name: "car_drivers", ForeignKey: "car_id", AssociationForeignKey: "driver_id"}, Polymorphic: Polymorphic{}},
			{Kind: HasMany, Field: "Wheels", ForeignKey: Field{Name: "ID"}, AssociationForeignKey: Field{Name: "CarID"}, SelfReference: false, JoinTable: JoinTable{}, Polymorphic: Polymorphic{}},
			//poly
			{Kind: HasOne, Field: "Radio", ForeignKey: Field{Name: "ID"}, AssociationForeignKey: Field{}, SelfReference: false, JoinTable: JoinTable{}, Polymorphic: Polymorphic{Field: Field{Name: "CarID"}, Type: Field{Name: "CarType"}, Value: "radio"}},
			{Kind: HasMany, Field: "Liquid", ForeignKey: Field{Name: "ID"}, AssociationForeignKey: Field{}, SelfReference: false, JoinTable: JoinTable{}, Polymorphic: Polymorphic{Field: Field{Name: "CarID"}, Type: Field{Name: "CarType"}, Value: "liquid"}},
			//custom
			{Kind: HasMany, Field: "CustomDriver", Custom: true},
		}

		for k, tt := range tests {
			t.Run(tt.Field, func(t *testing.T) {
				test.Equal(tt.Kind, c.relations[k].Kind)
				test.Equal(tt.Field, c.relations[k].Field)
				test.Equal(tt.ForeignKey.Name, c.relations[k].ForeignKey.Name)
				test.Equal(tt.AssociationForeignKey.Name, c.relations[k].AssociationForeignKey.Name)
				test.Equal(tt.SelfReference, c.relations[k].SelfReference)
				test.Equal(tt.JoinTable, c.relations[k].JoinTable)
				if tt.Polymorphic.Value != "" {
					test.Equal(tt.Polymorphic.Field.Name, c.relations[k].Polymorphic.Field.Name)
					test.Equal(tt.Polymorphic.Type.Name, c.relations[k].Polymorphic.Type.Name)
					test.Equal(tt.Polymorphic.Value, c.relations[k].Polymorphic.Value)
				} else {
					test.Equal(tt.Polymorphic, c.relations[k].Polymorphic)
				}
				test.Equal(tt.Custom, c.relations[k].Custom)
			})
		}
	}

	// self reference
	r := Role{}
	err = r.Init(&r)
	test.NoError(err)
	test.Equal(1, len(r.relations))
	test.Equal(ManyToMany, r.relations[0].Kind)
	test.Equal(true, r.relations[0].SelfReference)

	// Errors
	// err self reference on hasMany
	rs := RoleWrongRelation{}
	err = rs.Init(&rs)
	test.Error(err)
	test.Equal(errSelfReference.Error(), err.Error())

	// err poly m2m
	ce := carErrPolyM2m{}
	err = ce.Init(&ce)
	test.Error(err)
	test.Equal(fmt.Sprintf(errPolymorphicNotAllowed.Error(), "Liquid", "m2m"), err.Error())

	// err poly belongsTo
	c2 := carErrPolyB2{}
	err = c2.Init(&c2)
	test.Error(err)
	test.Equal(fmt.Sprintf(errPolymorphicNotAllowed.Error(), "Liquid", "belongsTo"), err.Error())

	// err relation type
	c3 := carErrRelation{}
	err = c3.Init(&c3)
	test.Error(err)
	test.Equal(fmt.Sprintf(errRelationKind.Error(), "orm2.carErrRelation", "Liquid", "m2m", "struct"), err.Error())

	// err belongsTo FK
	c4 := carBelongsToFKErr{}
	err = c4.Init(&c4)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "foo", "orm2.carBelongsToFKErr"), err.Error())

	// err belongsTo AFK
	c5 := carBelongsToAFKErr{}
	err = c5.Init(&c5)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "bar", "orm2.owner"), err.Error())

	// err M2M FK
	c6 := carM2MFkErr{}
	err = c6.Init(&c6)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "foo", "orm2.carM2MFkErr"), err.Error())

	// err M2M AFK
	c7 := carM2MAfkErr{}
	err = c7.Init(&c7)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "bar", "orm2.driver"), err.Error())

	// err M2M joinTable
	c8 := carM2MJoinErr{}
	err = c8.Init(&c8)
	test.Error(err)
	test.True(strings.Contains(err.Error(), "sqlquery: table"))

	// err hasMany fk err
	c9 := carHasManyFkErr{}
	err = c9.Init(&c9)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "foo", "orm2.carHasManyFkErr"), err.Error())

	// err hasMany afk err
	c10 := carHasManyAfkErr{}
	err = c10.Init(&c10)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "bar", "orm2.wheel"), err.Error())

	// err hasMany poly err
	c11 := carHasManyPolyErr{}
	err = c11.Init(&c11)
	test.Error(err)
	test.Equal(fmt.Sprintf(errStructField.Error(), "polyID", "orm2.wheel"), err.Error())

	// err cache
	c12 := carCacheErr{}
	err = c12.Init(&c12)
	test.Error(err)
	test.Equal(errNoCache.Error(), err.Error())
}
