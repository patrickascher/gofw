package orm

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func init() {
	GlobalBuilder, _ = HelperCreateBuilder()
}

func TestRelation_getRelationByTag(t *testing.T) {

	str, err := getRelationByTag("hasOne")
	assert.NoError(t, err)
	assert.Equal(t, HasOne, str)

	str, err = getRelationByTag("belongsTo")
	assert.NoError(t, err)
	assert.Equal(t, BelongsTo, str)

	str, err = getRelationByTag("hasMany")
	assert.NoError(t, err)
	assert.Equal(t, HasMany, str)

	str, err = getRelationByTag("manyToMany")
	assert.NoError(t, err)
	assert.Equal(t, ManyToMany, str)

	str, err = getRelationByTag("")
	assert.NoError(t, err)
	assert.Equal(t, "", str)

	str, err = getRelationByTag("somethingElse")
	assert.Error(t, err)
	assert.Equal(t, "", str)
}

func TestRelation_getRelationByType(t *testing.T) {

	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Orderfk{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	info := Contactfk{}
	err = info.Initialize(&info)
	assert.NoError(t, err)

	service := Servicefk{}
	err = service.Initialize(&service)
	assert.NoError(t, err)

	// hasOne
	fieldOrders, exists := reflect.TypeOf(customer).FieldByName("Info")
	if assert.True(t, exists) {
		rel, err := getRelationByType(&customer, &info, fieldOrders)
		if assert.NoError(t, err) {
			assert.Equal(t, HasOne, rel)
		}
	}

	// hasMany
	fieldOrders, exists = reflect.TypeOf(customer).FieldByName("Orders")
	if assert.True(t, exists) {
		rel, err := getRelationByType(&customer, &order, fieldOrders)
		if assert.NoError(t, err) {
			assert.Equal(t, HasMany, rel)
		}
	}

	// hasMany
	fieldOrders, exists = reflect.TypeOf(customer).FieldByName("Service")
	if assert.True(t, exists) {
		rel, err := getRelationByType(&customer, &service, fieldOrders)
		if assert.NoError(t, err) {
			assert.Equal(t, ManyToMany, rel)
		}
	}

	// belongsTo
	fieldOrders, exists = reflect.TypeOf(order).FieldByName("Customer")
	if assert.True(t, exists) {
		rel, err := getRelationByType(&order, &customer, fieldOrders)
		if assert.NoError(t, err) {
			assert.Equal(t, BelongsTo, rel)
		}
	}
}

func TestRelation_hasManyToMany(t *testing.T) {

	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	service := Servicefk{}
	err = service.Initialize(&service)
	assert.NoError(t, err)

	order := Orderfk{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	// customer_services exists in the db
	assert.True(t, hasManyToMany(&customer, &service))
	// customer_orders does not exist in db
	assert.False(t, hasManyToMany(&customer, &order))
}

func TestRelation_getManyToMany(t *testing.T) {

	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	service := Servicefk{}
	err = service.Initialize(&service)
	assert.NoError(t, err)

	fks, err := getManyToMany(&customer, &service)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(fks))

	assert.Equal(t, "customerfk_servicefks_ibfk_1", fks[0].Name)   // fk name of the table - relation
	assert.Equal(t, "customerfk_servicefks", fks[0].Primary.Table) //junction table
	assert.Equal(t, "customer_id", fks[0].Primary.Column)          // connection field
	assert.Equal(t, "customerfks", fks[0].Secondary.Table)         // association table
	assert.Equal(t, "id", fks[0].Secondary.Column)                 // association field

	assert.Equal(t, "customerfk_servicefks_ibfk_2", fks[1].Name)   // fk name of the table - relation
	assert.Equal(t, "customerfk_servicefks", fks[1].Primary.Table) //junction table
	assert.Equal(t, "service_id", fks[1].Primary.Column)           // connection field
	assert.Equal(t, "servicefks", fks[1].Secondary.Table)          // association table
	assert.Equal(t, "id", fks[1].Secondary.Column)                 // association field
}

func TestRelation_getRelation_DbDefinition(t *testing.T) {
	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Orderfk{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	info := Contactfk{}
	err = info.Initialize(&info)
	assert.NoError(t, err)

	service := Servicefk{}
	err = service.Initialize(&service)
	assert.NoError(t, err)

	infoField, exists := reflect.TypeOf(customer).FieldByName("Info")
	assert.True(t, exists)
	rel, err := getRelation(&customer, &info, infoField)
	assert.NoError(t, err)
	assert.Equal(t, HasOne, rel)

	ordersField, exists := reflect.TypeOf(customer).FieldByName("Orders")
	assert.True(t, exists)
	rel, err = getRelation(&customer, &order, ordersField)
	assert.NoError(t, err)
	assert.Equal(t, HasMany, rel)

	serviceField, exists := reflect.TypeOf(customer).FieldByName("Service")
	assert.True(t, exists)
	rel, err = getRelation(&customer, &service, serviceField)
	assert.NoError(t, err)
	assert.Equal(t, ManyToMany, rel)

	customerField, exists := reflect.TypeOf(order).FieldByName("Customer")
	assert.True(t, exists)
	rel, err = getRelation(&order, &customer, customerField)
	assert.NoError(t, err)
	assert.Equal(t, BelongsTo, rel)

}

// TODO m2m relation not working at the moment by tags.
func TestRelation_getRelation_TagDefinition(t *testing.T) {
	customer := Customer{}
	order := Order{}
	info := Contact{}
	//service := Service{}

	infoField, exists := reflect.TypeOf(customer).FieldByName("Info")
	assert.True(t, exists)
	rel, err := getRelation(&customer, &info, infoField)
	assert.NoError(t, err)
	assert.Equal(t, HasOne, rel)

	ordersField, exists := reflect.TypeOf(customer).FieldByName("Orders")
	assert.True(t, exists)
	rel, err = getRelation(&customer, &order, ordersField)
	assert.NoError(t, err)
	assert.Equal(t, HasMany, rel)

	//serviceField,exists := reflect.TypeOf(customer).FieldByName("Service")
	//assert.True(t, exists)
	//rel,err = getRelation(&customer,&service,serviceField)
	//assert.NoError(t, err)
	//assert.Equal(t,ManyToMany,rel)

	customerField, exists := reflect.TypeOf(order).FieldByName("Customer")
	assert.True(t, exists)
	rel, err = getRelation(&order, &customer, customerField)
	assert.NoError(t, err)
	assert.Equal(t, BelongsTo, rel)
}

func TestRelation_getForeignKeyByTag_EmptyTag(t *testing.T) {
	type customer struct {
		Model
		ID   int
		Info Contact
	}

	c := customer{}
	contact := Contact{}

	serviceField, exists := reflect.TypeOf(c).FieldByName("Info")
	assert.True(t, exists)
	fk, err := getForeignKeyByTag(&c, &contact, serviceField)
	assert.NoError(t, err)
	assert.True(t, fk == nil)
}

func TestRelation_getForeignKeyByTag_ShortTag(t *testing.T) {

	type customer struct {
		Model
		ID   int
		Info Contact `fk:"ID"`
		//		Info Contact `fk:"field:ID;associationField:CustomerID"`
	}

	c := customer{}
	err := c.Initialize(&c)
	assert.NoError(t, err)

	contact := Contact{}
	err = contact.Initialize(&contact)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(c).FieldByName("Info")
	assert.True(t, exists)
	fk, err := getForeignKeyByTag(&c, &contact, serviceField)
	assert.NoError(t, err)

	assert.Equal(t, "tag", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customers", fk.Primary.Table)
	assert.Equal(t, "id", fk.Primary.Column)
	assert.Equal(t, "contacts", fk.Secondary.Table)
	assert.Equal(t, "customer_id", fk.Secondary.Column)
}

func TestRelation_getForeignKeyByTag_ShortTagWithTagSpaces(t *testing.T) {

	type customer struct {
		Model
		ID   int
		Info Contact `fk:"  ID  "`
		//		Info Contact `fk:"field:ID;associationField:CustomerID"`
	}

	c := customer{}
	err := c.Initialize(&c)
	assert.NoError(t, err)

	contact := Contact{}
	err = contact.Initialize(&contact)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(c).FieldByName("Info")
	assert.True(t, exists)
	fk, err := getForeignKeyByTag(&c, &contact, serviceField)
	assert.NoError(t, err)

	assert.Equal(t, "tag", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customers", fk.Primary.Table)
	assert.Equal(t, "id", fk.Primary.Column)
	assert.Equal(t, "contacts", fk.Secondary.Table)
	assert.Equal(t, "customer_id", fk.Secondary.Column)

}

func TestRelation_getForeignKeyByTag_Tag(t *testing.T) {

	type customer struct {
		Model
		ID   int
		Info Contact `fk:"field:ID;associationField:CustomerID"`
	}

	c := customer{}
	err := c.Initialize(&c)
	assert.NoError(t, err)

	contact := Contact{}
	err = contact.Initialize(&contact)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(c).FieldByName("Info")
	assert.True(t, exists)
	fk, err := getForeignKeyByTag(&c, &contact, serviceField)
	assert.NoError(t, err)

	assert.Equal(t, "tag", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customers", fk.Primary.Table)
	assert.Equal(t, "id", fk.Primary.Column)
	assert.Equal(t, "contacts", fk.Secondary.Table)
	assert.Equal(t, "customer_id", fk.Secondary.Column)

}

func TestRelation_getForeignKeyByTag_TagWithSpaces(t *testing.T) {

	type customer struct {
		Model
		ID   int
		Info Contact `fk:"field:  ID  ;associationField:  CustomerID  "`
	}

	c := customer{}
	err := c.Initialize(&c)
	assert.NoError(t, err)

	contact := Contact{}
	err = contact.Initialize(&contact)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(c).FieldByName("Info")
	assert.True(t, exists)
	fk, err := getForeignKeyByTag(&c, &contact, serviceField)
	assert.NoError(t, err)

	assert.Equal(t, "tag", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customers", fk.Primary.Table)
	assert.Equal(t, "id", fk.Primary.Column)
	assert.Equal(t, "contacts", fk.Secondary.Table)
	assert.Equal(t, "customer_id", fk.Secondary.Column)

}

func TestRelation_getForeignKeyByDb(t *testing.T) {
	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Orderfk{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	// Customer FKs
	fk, err := getForeignKeyByDb(&customer, &order, "")
	assert.NoError(t, err)
	assert.True(t, fk == nil)

	// Customer FKs
	fk, err = getForeignKeyByDb(&order, &customer, "")
	assert.NoError(t, err)
	assert.Equal(t, "orderfks_ibfk_1", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "orderfks", fk.Primary.Table)
	assert.Equal(t, "customer_id", fk.Primary.Column)
	assert.Equal(t, "customerfks", fk.Secondary.Table)
	assert.Equal(t, "id", fk.Secondary.Column)
}

func TestRelation_getForeignKeyByDb_Junction(t *testing.T) {
	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	service := Servicefk{}
	err = service.Initialize(&service)
	assert.NoError(t, err)

	// junction services FKs
	fk, err := getForeignKeyByDb(&customer, &service, "customerfk_servicefks")
	assert.NoError(t, err)
	assert.Equal(t, "customerfk_servicefks_ibfk_2", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customerfk_servicefks", fk.Primary.Table)
	assert.Equal(t, "service_id", fk.Primary.Column)
	assert.Equal(t, "servicefks", fk.Secondary.Table)
	assert.Equal(t, "id", fk.Secondary.Column)

	// junction customer FKs
	fk, err = getForeignKeyByDb(&service, &customer, "customerfk_servicefks")
	assert.NoError(t, err)
	assert.Equal(t, "customerfk_servicefks_ibfk_1", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customerfk_servicefks", fk.Primary.Table)
	assert.Equal(t, "customer_id", fk.Primary.Column)
	assert.Equal(t, "customerfks", fk.Secondary.Table)
	assert.Equal(t, "id", fk.Secondary.Column)
}

func TestRelation_getForeignKey_DB(t *testing.T) {
	customer := Customerfk{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Orderfk{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(customer).FieldByName("Orders")
	assert.True(t, exists)

	fk, err := getForeignKey(&order, &customer, serviceField)
	assert.NoError(t, err)
	assert.Equal(t, "orderfks_ibfk_1", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "orderfks", fk.Primary.Table)
	assert.Equal(t, "customer_id", fk.Primary.Column)
	assert.Equal(t, "customerfks", fk.Secondary.Table)
	assert.Equal(t, "id", fk.Secondary.Column)

	fk, err = getForeignKey(&customer, &order, serviceField)
	assert.NoError(t, err)
	assert.True(t, fk == nil)
}

func TestRelation_getForeignKey_TAG(t *testing.T) {
	customer := Customer{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Order{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	serviceField, exists := reflect.TypeOf(customer).FieldByName("Orders")
	assert.True(t, exists)

	fk, err := getForeignKey(&customer, &order, serviceField)
	assert.NoError(t, err)
	assert.Equal(t, "tag", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "customers", fk.Primary.Table)
	assert.Equal(t, "id", fk.Primary.Column)
	assert.Equal(t, "orders", fk.Secondary.Table)
	assert.Equal(t, "customer_id", fk.Secondary.Column)

	fk, err = getForeignKey(&order, &customer, serviceField)
	assert.NoError(t, err)
	assert.Equal(t, "belongsTo", fk.Name) // identifier that its a fk added by a tag
	assert.Equal(t, "orders", fk.Primary.Table)
	assert.Equal(t, "customer_id", fk.Primary.Column)
	assert.Equal(t, "customers", fk.Secondary.Table)
	assert.Equal(t, "id", fk.Secondary.Column)
}

func TestRelation_addRelation(t *testing.T) {
	customer := Customer{}
	err := customer.Initialize(&customer)
	assert.NoError(t, err)

	order := Order{}
	err = order.Initialize(&order)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(customer.Table().Associations))
	assert.Equal(t, HasOne, customer.Table().Associations["Info"].Type)
	assert.Equal(t, "customers", customer.Table().Associations["Info"].StructTable.Information.Table)
	assert.Equal(t, "id", customer.Table().Associations["Info"].StructTable.Information.Name)
	assert.Equal(t, "contacts", customer.Table().Associations["Info"].AssociationTable.Information.Table)
	assert.Equal(t, "customer_id", customer.Table().Associations["Info"].AssociationTable.Information.Name)
	assert.Equal(t, HasMany, customer.Table().Associations["Orders"].Type)
	assert.Equal(t, "customers", customer.Table().Associations["Orders"].StructTable.Information.Table)
	assert.Equal(t, "id", customer.Table().Associations["Orders"].StructTable.Information.Name)
	assert.Equal(t, "orders", customer.Table().Associations["Orders"].AssociationTable.Information.Table)
	assert.Equal(t, "customer_id", customer.Table().Associations["Orders"].AssociationTable.Information.Name)
	//assert.Equal(t,ManyToMany,customer.Table().Associations["Services"].Type) // TODO ManyToMany with TAG not working yet
	//assert.Equal(t,BelongsTo,order.Table().Associations["Customer"].Type) // TODO Belongs with TAG not working correctly yet

	customerfk := Customerfk{}
	err = customerfk.Initialize(&customerfk)
	assert.NoError(t, err)

	orderfk := Orderfk{}
	err = orderfk.Initialize(&orderfk)
	assert.NoError(t, err)

	//assert.Equal(t, 2, len(orderfk.Table().Associations)) // TODO Belongs with TAG not working correctly yet

	assert.Equal(t, 3, len(customerfk.Table().Associations))
	assert.Equal(t, HasOne, customerfk.Table().Associations["Info"].Type)
	assert.Equal(t, "customerfks", customerfk.Table().Associations["Info"].StructTable.Information.Table)
	assert.Equal(t, "id", customerfk.Table().Associations["Info"].StructTable.Information.Name)
	assert.Equal(t, "contactfks", customerfk.Table().Associations["Info"].AssociationTable.Information.Table)
	assert.Equal(t, "customer_id", customerfk.Table().Associations["Info"].AssociationTable.Information.Name)
	assert.Equal(t, HasMany, customerfk.Table().Associations["Orders"].Type)
	assert.Equal(t, "customerfks", customerfk.Table().Associations["Orders"].StructTable.Information.Table)
	assert.Equal(t, "id", customerfk.Table().Associations["Orders"].StructTable.Information.Name)
	assert.Equal(t, "orderfks", customerfk.Table().Associations["Orders"].AssociationTable.Information.Table)
	assert.Equal(t, "customer_id", customerfk.Table().Associations["Orders"].AssociationTable.Information.Name)
	assert.Equal(t, ManyToMany, customerfk.Table().Associations["Service"].Type) // TODO ManyToMany with TAG not working yet
	assert.Equal(t, "customerfks", customerfk.Table().Associations["Service"].StructTable.Information.Table)
	assert.Equal(t, "id", customerfk.Table().Associations["Service"].StructTable.Information.Name)
	assert.Equal(t, "servicefks", customerfk.Table().Associations["Service"].AssociationTable.Information.Table)
	assert.Equal(t, "id", customerfk.Table().Associations["Service"].AssociationTable.Information.Name)
	assert.Equal(t, "customerfk_servicefks", customerfk.Table().Associations["Service"].JunctionTable.Table)
	assert.Equal(t, "customer_id", customerfk.Table().Associations["Service"].JunctionTable.StructColumn)
	assert.Equal(t, "service_id", customerfk.Table().Associations["Service"].JunctionTable.AssociationColumn)

	//assert.Equal(t,BelongsTo,orderfk.Table().Associations["Customer"].Type)
}
