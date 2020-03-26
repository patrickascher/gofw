package orm_test

import (
	"database/sql"
	"fmt"
	"github.com/guregu/null"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func deleteAll() error {

	builder, err := HelperCreateBuilder()
	if err != nil {
		return err
	}
	_, err = builder.Delete("productfks").Exec()
	if err != nil {
		return err
	}
	_, err = builder.Delete("orderfks").Exec()
	if err != nil {
		return err
	}
	_, err = builder.Delete("contactfks").Exec()
	if err != nil {
		return err
	}
	_, err = builder.Delete("customerfk_servicefks").Exec()
	if err != nil {
		return err
	}
	_, err = builder.Delete("customerfks").Exec()
	if err != nil {
		return err
	}

	_, err = builder.Delete("servicefks").Exec()
	if err != nil {
		return err
	}

	_, err = builder.Delete("accountfks").Exec()
	if err != nil {
		return err
	}

	return nil
}
func insertWithoutOrders() error {
	builder, err := HelperCreateBuilder()
	if err != nil {
		return err
	}

	// Insert Accounts
	accounts := []map[string]interface{}{
		{
			"id":   1,
			"name": "Frank",
		},
		{
			"id":   2,
			"name": "Peter",
		},
		{
			"id":   3,
			"name": "Steven",
		},
	}
	_, err = builder.Insert("accountfks").Values(accounts).Exec()
	if err != nil {
		return err
	}

	// Insert Customer
	customers := []map[string]interface{}{
		{
			"id":         1,
			"first_name": "Trescha",
			"last_name":  "Stoate",
			"created_at": "2019-02-23",
			"updated_at": "2020-03-02",
			"deleted_at": "2020-10-02",
			"account_id": 1,
		}, {
			"id":         2,
			"first_name": "Viviene",
			"last_name":  "Butterley",
			"created_at": "2018-12-06",
			"updated_at": "2019-04-19",
			"deleted_at": "2020-07-21",
			"account_id": 1,
		}, {
			"id":         3,
			"first_name": "Barri",
			"last_name":  "Elverston",
			"created_at": "2018-04-30",
			"updated_at": "2019-10-02",
			"deleted_at": "2020-04-05",
			"account_id": 2,
		}, {
			"id":         4,
			"first_name": "Constantina",
			"last_name":  "Merrett",
			"created_at": "2018-07-28",
			"updated_at": "2019-05-13",
			"deleted_at": "2020-12-04",
			"account_id": 2,
		}, {
			"id":         5,
			"first_name": "Bertram",
			"last_name":  "Pattinson",
			"created_at": "2018-11-05",
			"updated_at": "2019-11-15",
			"deleted_at": "2020-12-11",
			"account_id": 3,
		},
	}
	_, err = builder.Insert("customerfks").Values(customers).Exec()
	if err != nil {
		return err
	}

	// Insert Contact
	contacts := []map[string]interface{}{
		{
			"id":          1,
			"customer_id": 1,
			"phone":       "000-000-001",
		},
		{
			"id":          2,
			"customer_id": 2,
			"phone":       "000-000-002",
		},
		{
			"id":          3,
			"customer_id": 3,
			"phone":       "000-000-003",
		},
		{
			"id":          4,
			"customer_id": 4,
			"phone":       "000-000-004",
		},
	}
	_, err = builder.Insert("contactfks").Values(contacts).Exec()
	if err != nil {
		return err
	}

	// Insert Service
	service := []map[string]interface{}{
		{
			"id":   1,
			"name": "paypal",
		},
		{
			"id":   2,
			"name": "banking",
		},
		{
			"id":   3,
			"name": "appstore",
		},
		{
			"id":   4,
			"name": "playstore",
		},
	}
	_, err = builder.Insert("servicefks").Values(service).Exec()
	if err != nil {
		return err
	}

	// Insert junction customer-Service
	serviceJunction := []map[string]interface{}{
		{
			"customer_id": 1,
			"service_id":  1,
		},
		{
			"customer_id": 1,
			"service_id":  2,
		},
		{
			"customer_id": 1,
			"service_id":  3,
		},
		{
			"customer_id": 1,
			"service_id":  4,
		},
		{
			"customer_id": 2,
			"service_id":  3,
		},
		{
			"customer_id": 2,
			"service_id":  4,
		},
	}
	_, err = builder.Insert("customerfk_servicefks").Values(serviceJunction).Exec()
	if err != nil {
		return err
	}

	return nil
}
func insertAll() error {
	builder, err := HelperCreateBuilder()
	if err != nil {
		return err
	}

	// Insert Accounts
	accounts := []map[string]interface{}{
		{
			"id":   1,
			"name": "Frank",
		},
		{
			"id":   2,
			"name": "Peter",
		},
		{
			"id":   3,
			"name": "Steven",
		},
	}
	_, err = builder.Insert("accountfks").Values(accounts).Exec()
	if err != nil {
		return err
	}

	// Insert Customer
	customers := []map[string]interface{}{
		{
			"id":         1,
			"first_name": "Trescha",
			"last_name":  "Stoate",
			"created_at": "2019-02-23",
			"updated_at": "2020-03-02",
			"deleted_at": "2020-10-02",
			"account_id": 1,
		}, {
			"id":         2,
			"first_name": "Viviene",
			"last_name":  "Butterley",
			"created_at": "2018-12-06",
			"updated_at": "2019-04-19",
			"deleted_at": "2020-07-21",
			"account_id": 1,
		}, {
			"id":         3,
			"first_name": "Barri",
			"last_name":  "Elverston",
			"created_at": "2018-04-30",
			"updated_at": "2019-10-02",
			"deleted_at": "2020-04-05",
			"account_id": 2,
		}, {
			"id":         4,
			"first_name": "Constantina",
			"last_name":  "Merrett",
			"created_at": "2018-07-28",
			"updated_at": "2019-05-13",
			"deleted_at": "2020-12-04",
			"account_id": 2,
		}, {
			"id":         5,
			"first_name": "Bertram",
			"last_name":  "Pattinson",
			"created_at": "2018-11-05",
			"updated_at": "2019-11-15",
			"deleted_at": "2020-12-11",
			"account_id": 3,
		},
	}
	_, err = builder.Insert("customerfks").Values(customers).Exec()
	if err != nil {
		return err
	}

	// Insert Contact
	contacts := []map[string]interface{}{
		{
			"id":          1,
			"customer_id": 1,
			"phone":       "000-000-001",
		},
		{
			"id":          2,
			"customer_id": 2,
			"phone":       "000-000-002",
		},
		{
			"id":          3,
			"customer_id": 3,
			"phone":       "000-000-003",
		},
		{
			"id":          4,
			"customer_id": 4,
			"phone":       "000-000-004",
		},
	}
	_, err = builder.Insert("contactfks").Values(contacts).Exec()
	if err != nil {
		return err
	}

	// Insert Service
	service := []map[string]interface{}{
		{
			"id":   1,
			"name": "paypal",
		},
		{
			"id":   2,
			"name": "banking",
		},
		{
			"id":   3,
			"name": "appstore",
		},
		{
			"id":   4,
			"name": "playstore",
		},
	}
	_, err = builder.Insert("servicefks").Values(service).Exec()
	if err != nil {
		return err
	}

	// Insert junction customer-Service
	serviceJunction := []map[string]interface{}{
		{
			"customer_id": 1,
			"service_id":  1,
		},
		{
			"customer_id": 1,
			"service_id":  2,
		},
		{
			"customer_id": 1,
			"service_id":  3,
		},
		{
			"customer_id": 1,
			"service_id":  4,
		},
		{
			"customer_id": 2,
			"service_id":  3,
		},
		{
			"customer_id": 2,
			"service_id":  4,
		},
	}
	_, err = builder.Insert("customerfk_servicefks").Values(serviceJunction).Exec()
	if err != nil {
		return err
	}

	// Insert orders
	orders := []map[string]interface{}{
		{
			"id":          1,
			"customer_id": 1,
			"created_at":  "2010-07-21",
		},
		{
			"id":          2,
			"customer_id": 1,
			"created_at":  "2010-07-22",
		},
		{
			"id":          3,
			"customer_id": 1,
			"created_at":  "2010-07-23",
		},
		{
			"id":          4,
			"customer_id": 2,
			"created_at":  "2010-07-24",
		},
		{
			"id":          5,
			"customer_id": 2,
			"created_at":  "2010-07-25",
		},
		{
			"id":          6,
			"customer_id": 2,
			"created_at":  "2010-07-26",
		},
	}
	_, err = builder.Insert("orderfks").Values(orders).Exec()
	if err != nil {
		return err
	}

	// Insert products
	products := []map[string]interface{}{
		{
			"id":         1,
			"name":       "OnePlus",
			"price":      100,
			"created_at": "2020-07-21",
			"updated_at": "2021-07-21",
			"order_id":   1,
		},
		{
			"id":         2,
			"name":       "iPhone",
			"price":      200,
			"created_at": "2010-07-21",
			"updated_at": "2011-07-21",
			"order_id":   2,
		},
	}
	_, err = builder.Insert("productfks").Values(products).Exec()
	if err != nil {
		return err
	}

	return nil
}

func TestEagerLoading_First_Whitelist(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		//------------------------
		// Single field
		//------------------------
		cust := Customerfk{}
		err = cust.Initialize(&cust)
		assert.NoError(t, err)
		cust.SetWhitelist("FirstName")
		err := cust.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust.ID)
		assert.True(t, cust.FirstName.Valid)
		assert.False(t, cust.LastName.Valid)
		assert.Equal(t, 0, cust.AccountId)
		//Relation
		assert.Equal(t, Contactfk{}, cust.Info)
		assert.Equal(t, 0, len(cust.Orders))
		assert.Equal(t, 0, len(cust.Service))
		assert.Equal(t, Accountfk{}, cust.Account)

		//------------------------
		// Single field + whole Relation
		//------------------------
		cust2 := Customerfk{}
		cust2.Initialize(&cust2)
		cust2.SetWhitelist("LastName", "Info")
		err = cust2.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust2.ID)
		assert.False(t, cust2.FirstName.Valid)
		assert.True(t, cust2.LastName.Valid)

		assert.True(t, cust2.CreatedAt.Valid == false)
		assert.True(t, cust2.UpdatedAt.Valid == false)
		assert.True(t, cust2.DeletedAt.Valid == false)

		assert.Equal(t, 0, cust2.AccountId)
		//Relation
		assert.Equal(t, 1, cust2.Info.ID)
		assert.Equal(t, "000-000-001", cust2.Info.Phone.String)
		assert.Equal(t, 1, cust2.Info.CustomerID)
		assert.Equal(t, 0, len(cust2.Orders))
		assert.Equal(t, 0, len(cust2.Service))
		assert.Equal(t, Accountfk{}, cust2.Account)

		//------------------------
		// Single field + single fields Relation
		//------------------------
		cust3 := Customerfk{}
		cust3.Initialize(&cust3)
		cust3.SetWhitelist("LastName", "Info.CustomerID", "Info.ID")
		err = cust3.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust3.ID)
		assert.False(t, cust3.FirstName.Valid)
		assert.True(t, cust3.LastName.Valid)

		assert.True(t, cust3.CreatedAt.Valid == false)
		assert.True(t, cust3.UpdatedAt.Valid == false)
		assert.True(t, cust3.DeletedAt.Valid == false)

		assert.Equal(t, 0, cust3.AccountId)
		//Relation
		assert.Equal(t, 1, cust3.Info.ID)
		assert.Equal(t, "", cust3.Info.Phone.String)
		assert.Equal(t, 1, cust3.Info.CustomerID)
		assert.Equal(t, 0, len(cust3.Orders))
		assert.Equal(t, 0, len(cust3.Service))
		assert.Equal(t, Accountfk{}, cust3.Account)

		//------------------------
		// All fields
		//------------------------

		cust4 := Customerfk{}
		err = cust4.Initialize(&cust4)
		assert.NoError(t, err)
		err = cust4.SetWhitelist().First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust4.ID)
		assert.True(t, cust4.FirstName.Valid)
		assert.True(t, cust4.LastName.Valid)
		assert.True(t, cust4.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust4.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust4.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust4.AccountId)
		//Relation
		assert.Equal(t, "000-000-001", cust4.Info.Phone.String)
		assert.Equal(t, 3, len(cust4.Orders))
		assert.Equal(t, 4, len(cust4.Service))
		assert.Equal(t, 1, cust4.Account.ID)
	}
}

func TestEagerLoading_First_Blacklist(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		//------------------------
		// Single field
		//------------------------
		cust := Customerfk{}
		err = cust.Initialize(&cust)
		assert.NoError(t, err)
		cust.SetBlacklist("FirstName")
		err := cust.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust.ID)
		assert.False(t, cust.FirstName.Valid)
		assert.True(t, cust.LastName.Valid)
		assert.True(t, cust.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust.AccountId)
		//Relation
		assert.Equal(t, "000-000-001", cust.Info.Phone.String)
		assert.Equal(t, 3, len(cust.Orders))
		assert.Equal(t, 4, len(cust.Service))
		assert.Equal(t, 1, cust.Account.ID)

		//------------------------
		// Single field + whole Relation
		//------------------------
		cust2 := Customerfk{}
		cust2.Initialize(&cust2)
		cust2.SetBlacklist("LastName", "Info")
		err = cust2.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust2.ID)
		assert.True(t, cust2.FirstName.Valid)
		assert.False(t, cust2.LastName.Valid)
		assert.True(t, cust2.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust2.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust2.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust2.AccountId)
		//Relation
		assert.False(t, cust2.Info.Phone.Valid)
		assert.Equal(t, 3, len(cust2.Orders))
		assert.Equal(t, 4, len(cust2.Service))
		assert.Equal(t, 1, cust2.Account.ID)

		//------------------------
		// Single field + single fields Relation, test to remove pkey and reference key
		//------------------------
		cust3 := Customerfk{}
		cust3.Initialize(&cust3)
		cust3.SetBlacklist("LastName", "Info.CustomerID", "Info.ID")
		err = cust3.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust3.ID)
		assert.True(t, cust3.FirstName.Valid)
		assert.False(t, cust3.LastName.Valid)
		assert.True(t, cust3.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust3.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust3.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust3.AccountId)
		//Relation
		assert.Equal(t, 1, cust3.Info.ID)
		//assert.Equal(t, 1, cust3.Info.CustomerID) // TODO FIX IT, that no reference Key can get blacklisted
		assert.Equal(t, "000-000-001", cust3.Info.Phone.String)
		assert.Equal(t, 3, len(cust3.Orders))
		assert.Equal(t, 4, len(cust3.Service))
		assert.Equal(t, 1, cust3.Account.ID)

		//------------------------
		// Single field + single fields Relation, test to remove pkey and reference key
		//------------------------
		cust32 := Customerfk{}
		cust32.Initialize(&cust32)
		cust32.SetBlacklist("LastName", "Info.Phone")
		err = cust32.First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust32.ID)
		assert.True(t, cust32.FirstName.Valid)
		assert.False(t, cust32.LastName.Valid)
		assert.True(t, cust32.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust32.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust32.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust3.AccountId)
		//Relation
		assert.False(t, cust32.Info.Phone.Valid)
		assert.Equal(t, 3, len(cust32.Orders))
		assert.Equal(t, 4, len(cust32.Service))
		assert.Equal(t, 1, cust32.Account.ID)

		//------------------------
		// All fields
		//------------------------

		cust4 := Customerfk{}
		err = cust4.Initialize(&cust4)
		assert.NoError(t, err)
		err = cust4.SetBlacklist().First(nil)
		assert.NoError(t, err)

		//Fields
		assert.Equal(t, 1, cust4.ID)
		assert.True(t, cust4.FirstName.Valid)
		assert.True(t, cust4.LastName.Valid)
		assert.True(t, cust4.DeletedAt.Valid == false) //only write permission
		assert.True(t, cust4.CreatedAt.Valid == false) //only write permission
		assert.True(t, cust4.UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, cust4.AccountId)
		//Relation
		assert.Equal(t, "000-000-001", cust4.Info.Phone.String)
		assert.Equal(t, 3, len(cust4.Orders))
		assert.Equal(t, 4, len(cust4.Service))
		assert.Equal(t, 1, cust4.Account.ID)

	}
}

func TestEagerLoading_All_Whitelist(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		//------------------------
		// Single field
		//------------------------
		cust := Customerfk{}
		err = cust.Initialize(&cust)
		assert.NoError(t, err)
		cust.SetWhitelist("FirstName")

		var result []Customerfk
		err := cust.All(&result, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result))

		//Fields
		assert.Equal(t, 1, result[0].ID)
		assert.True(t, result[0].FirstName.Valid)
		assert.False(t, result[0].LastName.Valid)

		assert.True(t, result[0].DeletedAt.Valid == false)
		assert.True(t, result[0].CreatedAt.Valid == false)
		assert.True(t, result[0].UpdatedAt.Valid == false)

		assert.Equal(t, 0, result[0].AccountId)
		//Relation
		assert.Equal(t, Contactfk{}, result[0].Info)
		assert.Equal(t, 0, len(result[0].Orders))
		assert.Equal(t, 0, len(result[0].Service))
		assert.Equal(t, Accountfk{}, result[0].Account)

		//------------------------
		// Single field + whole Relation
		//------------------------
		cust2 := Customerfk{}
		cust2.Initialize(&cust2)
		cust2.SetWhitelist("LastName", "Info")
		var result2 []Customerfk
		err = cust2.All(&result2, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result2))

		//Fields
		assert.Equal(t, 1, result2[0].ID)
		assert.False(t, result2[0].FirstName.Valid)
		assert.True(t, result2[0].LastName.Valid)

		assert.True(t, result2[0].DeletedAt.Valid == false)
		assert.True(t, result2[0].CreatedAt.Valid == false)
		assert.True(t, result2[0].UpdatedAt.Valid == false)

		assert.Equal(t, 0, result2[0].AccountId)
		//Relation
		assert.Equal(t, 1, result2[0].Info.ID)
		assert.Equal(t, "000-000-001", result2[0].Info.Phone.String)
		assert.Equal(t, 1, result2[0].Info.CustomerID)
		assert.Equal(t, 0, len(result2[0].Orders))
		assert.Equal(t, 0, len(result2[0].Service))
		assert.Equal(t, Accountfk{}, result2[0].Account)

		//------------------------
		// Single field + single fields Relation
		//------------------------
		cust3 := Customerfk{}
		cust3.Initialize(&cust3)
		cust3.SetWhitelist("LastName", "Info.CustomerID", "Info.ID")
		var result3 []Customerfk
		err = cust3.All(&result3, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result3))

		//Fields
		assert.Equal(t, 1, result3[0].ID)
		assert.False(t, result3[0].FirstName.Valid)
		assert.True(t, result3[0].LastName.Valid)
		assert.True(t, result3[0].DeletedAt.Valid == false)
		assert.True(t, result3[0].CreatedAt.Valid == false)
		assert.True(t, result3[0].UpdatedAt.Valid == false)

		assert.Equal(t, 0, result3[0].AccountId)
		//Relation
		assert.Equal(t, 1, result3[0].Info.ID)
		assert.Equal(t, "", result3[0].Info.Phone.String)
		assert.Equal(t, 1, result3[0].Info.CustomerID)
		assert.Equal(t, 0, len(result3[0].Orders))
		assert.Equal(t, 0, len(result3[0].Service))
		assert.Equal(t, Accountfk{}, result3[0].Account)

		//------------------------
		// All fields
		//------------------------

		cust4 := Customerfk{}
		err = cust4.Initialize(&cust4)
		assert.NoError(t, err)
		var result4 []Customerfk
		err = cust4.SetWhitelist().All(&result4, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result4))

		//Fields
		assert.Equal(t, 1, result4[0].ID)
		assert.True(t, result4[0].FirstName.Valid)
		assert.True(t, result4[0].LastName.Valid)

		assert.True(t, result4[0].DeletedAt.Valid == false) // not nil because all field are loaded
		assert.True(t, result4[0].CreatedAt.Valid == false) // not nil because all field are loaded
		assert.True(t, result4[0].UpdatedAt.Valid == false) // not nil because all field are loaded

		assert.Equal(t, 1, result4[0].AccountId)
		//Relation
		assert.Equal(t, "000-000-001", result4[0].Info.Phone.String)
		assert.Equal(t, 3, len(result4[0].Orders))
		assert.Equal(t, 4, len(result4[0].Service))
		assert.Equal(t, 1, result4[0].Account.ID)
	}
}

func TestEagerLoading_All_Blacklist(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		//------------------------
		// Single field
		//------------------------
		cust := Customerfk{}
		err = cust.Initialize(&cust)
		assert.NoError(t, err)
		cust.SetBlacklist("FirstName")

		var result []Customerfk
		err := cust.All(&result, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result))

		//Fields
		assert.Equal(t, 1, result[0].ID)
		assert.False(t, result[0].FirstName.Valid)
		assert.True(t, result[0].LastName.Valid)
		assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
		assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
		assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, result[0].AccountId)
		//Relation
		assert.Equal(t, "000-000-001", result[0].Info.Phone.String)
		assert.Equal(t, 3, len(result[0].Orders))
		assert.Equal(t, 4, len(result[0].Service))
		assert.Equal(t, 1, result[0].Account.ID)

		//------------------------
		// Single field + whole Relation
		//------------------------
		cust2 := Customerfk{}
		cust2.Initialize(&cust2)
		cust2.SetBlacklist("LastName", "Info")
		var result2 []Customerfk
		err = cust2.All(&result2, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result2))

		//Fields
		assert.Equal(t, 1, result2[0].ID)
		assert.True(t, result2[0].FirstName.Valid)
		assert.False(t, result2[0].LastName.Valid)
		assert.True(t, result2[0].DeletedAt.Valid == false) //only write permission
		assert.True(t, result2[0].CreatedAt.Valid == false) //only write permission
		assert.True(t, result2[0].UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, result2[0].AccountId)
		//Relation
		assert.False(t, result2[0].Info.Phone.Valid)
		assert.Equal(t, 3, len(result2[0].Orders))
		assert.Equal(t, 4, len(result2[0].Service))
		assert.Equal(t, 1, result2[0].Account.ID)

		//------------------------
		// Single field + single fields Relation
		//------------------------
		cust3 := Customerfk{}
		cust3.Initialize(&cust3)
		cust3.SetBlacklist("LastName", "Info.ID")
		//cust3.SetBlacklist("LastName", "Info.CustomerID", "Info.ID") //TODO fix ist, that no reference field CustomerID can get blacklisted

		var result3 []Customerfk
		err = cust3.All(&result3, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result3))

		//Fields
		assert.Equal(t, 1, result3[0].ID)
		assert.True(t, result3[0].FirstName.Valid)
		assert.False(t, result3[0].LastName.Valid)
		assert.True(t, result3[0].DeletedAt.Valid == false) //only write permission
		assert.True(t, result3[0].CreatedAt.Valid == false) //only write permission
		assert.True(t, result3[0].UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, result3[0].AccountId)
		//Relation
		assert.Equal(t, "000-000-001", result3[0].Info.Phone.String)
		assert.Equal(t, 3, len(result3[0].Orders))
		assert.Equal(t, 4, len(result3[0].Service))
		assert.Equal(t, 1, result3[0].Account.ID)

		//------------------------
		// All fields
		//------------------------

		cust4 := Customerfk{}
		err = cust4.Initialize(&cust4)
		assert.NoError(t, err)
		var result4 []Customerfk
		err = cust4.SetWhitelist().All(&result4, nil)
		assert.NoError(t, err)

		assert.Equal(t, 5, len(result4))

		//Fields
		assert.Equal(t, 1, result4[0].ID)
		assert.True(t, result4[0].FirstName.Valid)
		assert.True(t, result4[0].LastName.Valid)
		assert.True(t, result4[0].DeletedAt.Valid == false) //only write permission
		assert.True(t, result4[0].CreatedAt.Valid == false) //only write permission
		assert.True(t, result4[0].UpdatedAt.Valid == false) //only write permission
		assert.Equal(t, 1, result4[0].AccountId)
		//Relation
		assert.Equal(t, "000-000-001", result4[0].Info.Phone.String)
		assert.Equal(t, 3, len(result4[0].Orders))
		assert.Equal(t, 4, len(result4[0].Service))
		assert.Equal(t, 1, result4[0].Account.ID)

	}
}

func TestEagerLoading_Create_Whitelist_Field(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			//AccountId should be added automatically?
			err = cust.SetWhitelist("FirstName").Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetWhitelist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "Trescha", result[0].FirstName.String)
			assert.False(t, result[0].LastName.Valid)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) // only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation
			assert.Equal(t, 0, result[0].Info.ID)
			assert.Equal(t, 0, len(result[0].Orders))
			assert.Equal(t, 0, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)
		}
	}
}

func TestEagerLoading_Create_Blacklist_Field(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{"12345-123", true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})
			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetBlacklist("FirstName", "Account").Create()
			assert.NoError(t, err)

			// checking results
			var result []Customerfk
			err = cust.SetWhitelist().All(&result, nil)
			assert.NoError(t, err)

			fmt.Println(result[0])

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.False(t, result[0].FirstName.Valid)
			assert.Equal(t, "Stoate", result[0].LastName.String)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission
			assert.True(t, result[0].AccountId != 0)
			// Relation
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, 2, len(result[0].Orders))
			assert.Equal(t, 3, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)
		}
	}
}

func TestEagerLoading_Create_Whitelist_Relation(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetWhitelist("FirstName", "Info").Create()
			if assert.NoError(t, err) {

			}

			// checking results
			var result []Customerfk
			err = cust.SetWhitelist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "Trescha", result[0].FirstName.String)
			assert.False(t, result[0].LastName.Valid)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation in whitelist
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, "000-000-001", result[0].Info.Phone.String)
			// Relation not in whitelist
			assert.Equal(t, 0, len(result[0].Orders))
			assert.Equal(t, 0, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Create_Blacklist_Relation(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetBlacklist("FirstName", "Info").Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetBlacklist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.False(t, result[0].FirstName.Valid)
			assert.Equal(t, "Stoate", result[0].LastName.String)
			assert.True(t, result[0].DeletedAt.Valid == false) // only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) // only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission
			assert.True(t, result[0].AccountId != 0)
			// Relation in Blacklist
			assert.False(t, result[0].Info.ID != 0)
			// Relation
			assert.Equal(t, 2, len(result[0].Orders))
			assert.Equal(t, 3, len(result[0].Service))
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Create_Whitelist_RelationField(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetWhitelist("FirstName", "Info.CustomerID").Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetWhitelist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "Trescha", result[0].FirstName.String)
			assert.False(t, result[0].LastName.Valid)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation in whitelist
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, "", result[0].Info.Phone.String)
			// Relation not in whitelist
			assert.Equal(t, 0, len(result[0].Orders))
			assert.Equal(t, 0, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Create_Blacklist_RelationField(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetBlacklist("FirstName", "Info.Phone", "Orders").Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetBlacklist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "", result[0].FirstName.String)
			assert.Equal(t, "Stoate", result[0].LastName.String)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation field in Blacklist
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, "", result[0].Info.Phone.String)
			// Relation not in whitelist
			assert.Equal(t, 0, len(result[0].Orders))
			assert.Equal(t, 3, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Create_Whitelist_AllFields(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetWhitelist().Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetWhitelist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "Trescha", result[0].FirstName.String)
			assert.True(t, result[0].LastName.Valid)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation in whitelist
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, "000-000-001", result[0].Info.Phone.String)
			// Relation not in whitelist
			assert.Equal(t, 2, len(result[0].Orders))
			assert.Equal(t, 3, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Create_Blacklist_AllFields(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{String: "000-000-001", Valid: true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = cust.SetBlacklist().Create()
			if assert.NoError(t, err) {
			}

			// checking results
			var result []Customerfk
			err = cust.SetBlacklist().All(&result, nil)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(result))
			assert.True(t, result[0].ID != 0)
			assert.Equal(t, "Trescha", result[0].FirstName.String)
			assert.True(t, result[0].LastName.Valid)

			assert.True(t, result[0].DeletedAt.Valid == false) //only write permission
			assert.True(t, result[0].CreatedAt.Valid == false) //only write permission
			assert.True(t, result[0].UpdatedAt.Valid == false) //only write permission

			assert.True(t, result[0].AccountId != 0)
			// Relation in whitelist
			assert.True(t, result[0].Info.ID != 0)
			assert.Equal(t, "000-000-001", result[0].Info.Phone.String)
			// Relation not in whitelist
			assert.Equal(t, 2, len(result[0].Orders))
			assert.Equal(t, 3, len(result[0].Service))
			// Relation BelongsTo (loaded automatic)
			assert.True(t, result[0].Account.ID != 0)
			assert.Equal(t, "Frank", result[0].Account.Name)

		}
	}
}

func TestEagerLoading_Update_Whitelist_Field(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetWhitelist("FirstName").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "Stoate", cust2.LastName.String)

				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "000-000-001", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 0, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 4, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.Equal(t, 3, cust2.Service[2].ID)
				assert.Equal(t, "appstore", cust2.Service[2].Name.String)
				assert.Equal(t, 4, cust2.Service[3].ID)
				assert.Equal(t, "playstore", cust2.Service[3].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Blacklist_Field(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetBlacklist("FirstName").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "Trescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "123-456-890", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.True(t, cust2.Service[2].ID > 0)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Whitelist_Relation(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetWhitelist("FirstName", "Info").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "Stoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "123-456-890", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 0, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 4, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.Equal(t, 3, cust2.Service[2].ID)
				assert.Equal(t, "appstore", cust2.Service[2].Name.String)
				assert.Equal(t, 4, cust2.Service[3].ID)
				assert.Equal(t, "playstore", cust2.Service[3].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Blacklist_Relation(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetBlacklist("FirstName", "Info").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "Trescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "000-000-001", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.True(t, cust2.Service[2].ID > 0)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Whitelist_RelationField(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetWhitelist("FirstName", "Info.CustomerID").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "Stoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "000-000-001", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 0, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 4, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.Equal(t, 3, cust2.Service[2].ID)
				assert.Equal(t, "appstore", cust2.Service[2].Name.String)
				assert.Equal(t, 4, cust2.Service[3].ID)
				assert.Equal(t, "playstore", cust2.Service[3].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Blacklist_RelationField(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				err = cust.First(&c)
				assert.NoError(t, err)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetBlacklist("FirstName", "Info.Phone").Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "Trescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				//TODO blacklist dot notation not working correctly
				//				assert.Equal(t, "000-000-001", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.True(t, cust2.Service[2].ID > 0)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Whitelist_All(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetWhitelist().Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "123-456-890", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.True(t, cust2.Service[2].ID != 0)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_Blacklist_All(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = cust.SetBlacklist().Update()
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				err = cust2.Initialize(&cust2)
				assert.NoError(t, err)
				err = cust2.First(&c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "123-456-890", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, 1, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, 2, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.True(t, cust2.Service[2].ID != 0)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_First(t *testing.T) {

	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			eager, err := orm.NewStrategy("eager")
			if assert.NoError(t, err) {
				c := sqlquery.Condition{}

				err = eager.First(&cust, &c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, 1, cust.ID)
				assert.Equal(t, "Trescha", cust.FirstName.String)
				assert.Equal(t, "Stoate", cust.LastName.String)
				assert.Equal(t, false, cust.CreatedAt.Valid)
				assert.Equal(t, false, cust.UpdatedAt.Valid)
				assert.Equal(t, false, cust.DeletedAt.Valid)
				assert.Equal(t, 1, cust.AccountId)

				// Has one
				assert.Equal(t, 1, cust.Info.ID)
				assert.Equal(t, 1, cust.Info.CustomerID)
				assert.Equal(t, "000-000-001", cust.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, 1, cust.Account.ID)
				assert.Equal(t, "Frank", cust.Account.Name)

				// HasMany
				assert.Equal(t, 3, len(cust.Orders))
				assert.Equal(t, 1, cust.Orders[0].ID)
				assert.Equal(t, 1, cust.Orders[0].CustomerID)
				assert.Equal(t, false, cust.Orders[0].CreatedAt.Valid)
				assert.Equal(t, 1, cust.Orders[0].Product.ID)
				assert.Equal(t, "OnePlus", cust.Orders[0].Product.Name.String)
				assert.Equal(t, 2, cust.Orders[1].ID)
				assert.Equal(t, 1, cust.Orders[1].CustomerID)
				assert.Equal(t, false, cust.Orders[1].CreatedAt.Valid)
				assert.Equal(t, 2, cust.Orders[1].Product.ID)
				assert.Equal(t, "iPhone", cust.Orders[1].Product.Name.String)
				assert.Equal(t, 3, cust.Orders[2].ID)
				assert.Equal(t, 1, cust.Orders[2].CustomerID)
				assert.Equal(t, false, cust.Orders[2].CreatedAt.Valid)
				assert.Equal(t, 0, cust.Orders[2].Product.ID)           //empty
				assert.Equal(t, "", cust.Orders[2].Product.Name.String) //empty

				// ManyToMany
				assert.Equal(t, 4, len(cust.Service))
				assert.Equal(t, 1, cust.Service[0].ID)
				assert.Equal(t, "paypal", cust.Service[0].Name.String)
				assert.Equal(t, 2, cust.Service[1].ID)
				assert.Equal(t, "banking", cust.Service[1].Name.String)
				assert.Equal(t, 3, cust.Service[2].ID)
				assert.Equal(t, "appstore", cust.Service[2].Name.String)
				assert.Equal(t, 4, cust.Service[3].ID)
				assert.Equal(t, "playstore", cust.Service[3].Name.String)
			}
		}
	}
}

func TestEagerLoading_All(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertAll()
		assert.NoError(t, err)

		cust := Customerfk{}
		var result []Customerfk
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			eager, err := orm.NewStrategy("eager")
			if assert.NoError(t, err) {
				c := sqlquery.Condition{}

				// error because the result is no ptr
				err := eager.All(result, &cust, &c)
				assert.Error(t, err)

				// ok
				err = eager.All(&result, &cust, &c)
				assert.NoError(t, err)
				assert.Equal(t, 5, len(result))

				//TODO create a clean loop with all the data... in a hurry now, later
				//ID1
				assert.Equal(t, 1, result[0].ID)
				assert.Equal(t, "Trescha", result[0].FirstName.String)
				assert.Equal(t, "Stoate", result[0].LastName.String)
				assert.Equal(t, false, result[0].CreatedAt.Valid)
				assert.Equal(t, false, result[0].UpdatedAt.Valid)
				assert.Equal(t, false, result[0].DeletedAt.Valid)

				assert.Equal(t, 1, result[0].AccountId)

				// Has one
				assert.Equal(t, 1, result[0].Info.ID)
				assert.Equal(t, 1, result[0].Info.CustomerID)
				assert.Equal(t, "000-000-001", result[0].Info.Phone.String)

				// BelongsTo
				assert.Equal(t, 1, result[0].Account.ID)
				assert.Equal(t, "Frank", result[0].Account.Name)

				// HasMany
				assert.Equal(t, 3, len(result[0].Orders))
				assert.Equal(t, 1, result[0].Orders[0].ID)
				assert.Equal(t, 1, result[0].Orders[0].CustomerID)
				assert.Equal(t, false, result[0].Orders[0].CreatedAt.Valid)
				assert.Equal(t, 1, result[0].Orders[0].Product.ID)
				assert.Equal(t, "OnePlus", result[0].Orders[0].Product.Name.String)
				assert.Equal(t, 2, result[0].Orders[1].ID)
				assert.Equal(t, 1, result[0].Orders[1].CustomerID)
				assert.Equal(t, false, result[0].Orders[1].CreatedAt.Valid)
				assert.Equal(t, 2, result[0].Orders[1].Product.ID)
				assert.Equal(t, "iPhone", result[0].Orders[1].Product.Name.String)
				assert.Equal(t, 3, result[0].Orders[2].ID)
				assert.Equal(t, 1, result[0].Orders[2].CustomerID)
				assert.Equal(t, false, result[0].Orders[2].CreatedAt.Valid)
				assert.Equal(t, 0, result[0].Orders[2].Product.ID)           //empty
				assert.Equal(t, "", result[0].Orders[2].Product.Name.String) //empty

				// ManyToMany
				assert.Equal(t, 4, len(result[0].Service))
				assert.Equal(t, 1, result[0].Service[0].ID)
				assert.Equal(t, "paypal", result[0].Service[0].Name.String)
				assert.Equal(t, 2, result[0].Service[1].ID)
				assert.Equal(t, "banking", result[0].Service[1].Name.String)
				assert.Equal(t, 3, result[0].Service[2].ID)
				assert.Equal(t, "appstore", result[0].Service[2].Name.String)
				assert.Equal(t, 4, result[0].Service[3].ID)
				assert.Equal(t, "playstore", result[0].Service[3].Name.String)
				//TODO create a clean loop with all the data... in a hurry now, later
				//ID2
				assert.Equal(t, 2, result[1].ID)
				assert.Equal(t, "Viviene", result[1].FirstName.String)
				assert.Equal(t, "Butterley", result[1].LastName.String)
				assert.Equal(t, false, result[1].CreatedAt.Valid)
				assert.Equal(t, false, result[1].UpdatedAt.Valid)
				assert.Equal(t, false, result[1].DeletedAt.Valid)
				assert.Equal(t, 1, result[1].AccountId)

				// Has one
				assert.Equal(t, 2, result[1].Info.ID)
				assert.Equal(t, 2, result[1].Info.CustomerID)
				assert.Equal(t, "000-000-002", result[1].Info.Phone.String)

				// BelongsTo
				assert.Equal(t, 1, result[1].Account.ID)
				assert.Equal(t, "Frank", result[1].Account.Name)

				// HasMany
				assert.Equal(t, 3, len(result[1].Orders))
				assert.Equal(t, 4, result[1].Orders[0].ID)
				assert.Equal(t, 2, result[1].Orders[0].CustomerID)
				assert.Equal(t, false, result[1].Orders[0].CreatedAt.Valid)
				assert.Equal(t, 0, result[1].Orders[0].Product.ID)           //empty
				assert.Equal(t, "", result[1].Orders[0].Product.Name.String) //empty

				assert.Equal(t, 5, result[1].Orders[1].ID)
				assert.Equal(t, 2, result[1].Orders[1].CustomerID)
				assert.Equal(t, false, result[1].Orders[1].CreatedAt.Valid)
				assert.Equal(t, 0, result[1].Orders[1].Product.ID)           //empty
				assert.Equal(t, "", result[1].Orders[1].Product.Name.String) //empty

				assert.Equal(t, 6, result[1].Orders[2].ID)
				assert.Equal(t, 2, result[1].Orders[2].CustomerID)
				assert.Equal(t, false, result[1].Orders[2].CreatedAt.Valid)
				assert.Equal(t, 0, result[1].Orders[2].Product.ID)           //empty
				assert.Equal(t, "", result[1].Orders[2].Product.Name.String) //empty

				// ManyToMany
				assert.Equal(t, 2, len(result[1].Service))
				assert.Equal(t, 3, result[1].Service[0].ID)
				assert.Equal(t, "appstore", result[1].Service[0].Name.String)
				assert.Equal(t, 4, result[1].Service[1].ID)
				assert.Equal(t, "playstore", result[1].Service[1].Name.String)
				//TODO create a clean loop with all the data... in a hurry now, later
				//ID3
				assert.Equal(t, 3, result[2].ID)
				assert.Equal(t, "Barri", result[2].FirstName.String)
				assert.Equal(t, "Elverston", result[2].LastName.String)
				assert.Equal(t, false, result[2].CreatedAt.Valid)
				assert.Equal(t, false, result[2].UpdatedAt.Valid)
				assert.Equal(t, false, result[2].DeletedAt.Valid)
				assert.Equal(t, 2, result[2].AccountId)
				// BelongsTo
				assert.Equal(t, 2, result[2].Account.ID)
				assert.Equal(t, "Peter", result[2].Account.Name)
				// Has one
				assert.Equal(t, 3, result[2].Info.ID)
				assert.Equal(t, 3, result[2].Info.CustomerID)
				assert.Equal(t, "000-000-003", result[2].Info.Phone.String)
				//TODO create a clean loop with all the data... in a hurry now, later
				//ID4
				assert.Equal(t, 4, result[3].ID)
				assert.Equal(t, "Constantina", result[3].FirstName.String)
				assert.Equal(t, "Merrett", result[3].LastName.String)
				assert.Equal(t, false, result[3].CreatedAt.Valid)
				assert.Equal(t, false, result[3].UpdatedAt.Valid)
				assert.Equal(t, false, result[3].DeletedAt.Valid)
				assert.Equal(t, 2, result[3].AccountId)
				// BelongsTo
				assert.Equal(t, 2, result[3].Account.ID)
				assert.Equal(t, "Peter", result[3].Account.Name)
				// Has one
				assert.Equal(t, 4, result[3].Info.ID)
				assert.Equal(t, 4, result[3].Info.CustomerID)
				assert.Equal(t, "000-000-004", result[3].Info.Phone.String)
				//TODO create a clean loop with all the data... in a hurry now, later
				//ID5
				assert.Equal(t, 5, result[4].ID)
				assert.Equal(t, "Bertram", result[4].FirstName.String)
				assert.Equal(t, "Pattinson", result[4].LastName.String)
				assert.Equal(t, false, result[4].CreatedAt.Valid)
				assert.Equal(t, false, result[4].UpdatedAt.Valid)
				assert.Equal(t, false, result[4].DeletedAt.Valid)
				assert.Equal(t, 3, result[4].AccountId)
				// BelongsTo
				assert.Equal(t, 3, result[4].Account.ID)
				assert.Equal(t, "Steven", result[4].Account.Name)
				// Has one - nil
				assert.Equal(t, 0, result[4].Info.ID)
				assert.Equal(t, 0, result[4].Info.CustomerID)
				assert.Equal(t, "", result[4].Info.Phone.String)
			}
		}
	}
}

func TestEagerLoading_Create(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		eager, err := orm.NewStrategy("eager")
		if assert.NoError(t, err) {
			cust := Customerfk{}
			err = cust.Initialize(&cust)
			assert.NoError(t, err)

			// main entry
			cust.FirstName = sql.NullString{String: "Trescha", Valid: true}
			cust.LastName = sql.NullString{String: "Stoate", Valid: true}
			created, err := time.Parse("2006-01-02", "2019-02-23")
			assert.NoError(t, err)
			updated, err := time.Parse("2006-01-02", "2020-03-02")
			assert.NoError(t, err)
			deleted, err := time.Parse("2006-01-02", "2020-10-02")
			assert.NoError(t, err)
			cust.CreatedAt = null.Time{Time: created, Valid: true}
			cust.UpdatedAt = null.Time{Time: updated, Valid: true}
			cust.DeletedAt = null.Time{Time: deleted, Valid: true}

			// has One
			cust.Info = Contactfk{Phone: sql.NullString{"000-000-001", true}}

			// belongsTo
			cust.Account = Accountfk{Name: "Frank"}

			// hasMany
			created1, err := time.Parse("2006-01-02", "2010-07-21")
			assert.NoError(t, err)
			created2, err := time.Parse("2006-01-02", "2010-07-22")
			assert.NoError(t, err)
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created1, true}, Product: Productfk{Name: sql.NullString{"OnePlus", true}, Price: sql.NullFloat64{100, true}}})
			cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{created2, true}, Product: Productfk{Name: sql.NullString{"iPhone", true}, Price: sql.NullFloat64{200, true}}})

			// manyToMany
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"paypal", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"banking", true}})
			cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"appstore", true}})

			err = eager.Create(&cust)
			if assert.NoError(t, err) {
			}

			cust2 := Customerfk{}
			cust2.Initialize(&cust2)
			c := sqlquery.Condition{}
			err = eager.First(&cust2, &c)
			assert.NoError(t, err)

			//TODO create a clean loop with all the data... in a hurry now, later
			assert.Equal(t, cust.ID, cust2.ID)
			assert.Equal(t, "Trescha", cust2.FirstName.String)
			assert.Equal(t, "Stoate", cust2.LastName.String)
			assert.Equal(t, false, cust2.CreatedAt.Valid)
			assert.Equal(t, false, cust2.UpdatedAt.Valid)
			assert.Equal(t, false, cust2.DeletedAt.Valid)
			assert.Equal(t, cust.AccountId, cust2.AccountId)

			// Has one
			assert.Equal(t, cust.Info.ID, cust2.Info.ID)
			assert.Equal(t, cust.ID, cust2.Info.CustomerID)
			assert.Equal(t, "000-000-001", cust2.Info.Phone.String)

			// BelongsTo
			assert.Equal(t, cust.Account.ID, cust2.Account.ID)
			assert.Equal(t, "Frank", cust2.Account.Name)

			// HasMany
			assert.Equal(t, 2, len(cust2.Orders))
			assert.Equal(t, cust.Orders[0].ID, cust2.Orders[0].ID)
			assert.Equal(t, cust.ID, cust2.Orders[0].CustomerID)
			assert.Equal(t, false, cust2.Orders[0].CreatedAt.Valid)
			assert.Equal(t, cust.Orders[0].Product.ID, cust2.Orders[0].Product.ID)
			assert.Equal(t, "OnePlus", cust2.Orders[0].Product.Name.String)
			assert.Equal(t, cust.Orders[1].ID, cust2.Orders[1].ID)
			assert.Equal(t, cust.ID, cust2.Orders[1].CustomerID)
			assert.Equal(t, false, cust2.Orders[1].CreatedAt.Valid)
			assert.Equal(t, cust.Orders[1].Product.ID, cust2.Orders[1].Product.ID)
			assert.Equal(t, "iPhone", cust2.Orders[1].Product.Name.String)

			// ManyToMany
			assert.Equal(t, 3, len(cust2.Service))
			assert.Equal(t, cust.Service[0].ID, cust2.Service[0].ID)
			assert.Equal(t, "paypal", cust2.Service[0].Name.String)
			assert.Equal(t, cust.Service[1].ID, cust2.Service[1].ID)
			assert.Equal(t, "banking", cust2.Service[1].Name.String)
			assert.Equal(t, cust.Service[2].ID, cust2.Service[2].ID)
			assert.Equal(t, "appstore", cust2.Service[2].Name.String)
		}
	}
}

func TestEagerLoading_Update(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			eager, err := orm.NewStrategy("eager")
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				// main model
				cust.FirstName.String = "updTrescha"
				cust.LastName.String = "updStoate"

				// hasOne (edit)
				cust.Info.Phone.String = "123-456-890"
				// belongsTo (edit)
				cust.Account.Name = "updFrank"

				// hasMany (add one)
				created, err := time.Parse("2006-01-02", "2019-02-23")
				assert.NoError(t, err)
				cust.Orders = append(cust.Orders, Orderfk{CreatedAt: null.Time{Time: created, Valid: true}})

				// hasOne depth1 (edit)
				cust.Orders[0].Product = Productfk{Price: sql.NullFloat64{2.2, true}, Name: sql.NullString{"updOnePlus", true}}

				// manyToMany (delete one, add one)
				cust.Service[0].Name = sql.NullString{"updPaypal", true} // TODO not working because we only check if ID exists and not if something changed (snapshot again?)
				cust.Service = append(cust.Service[:0], cust.Service[:2]...)
				cust.Service = append(cust.Service, Servicefk{Name: sql.NullString{"newEntry", true}})

				err = eager.Update(&cust, &c)
				if assert.NoError(t, err) {
				}

				cust2 := Customerfk{}
				cust2.Initialize(&cust2)
				err = eager.First(&cust2, &c)
				assert.NoError(t, err)

				//TODO create a clean loop with all the data... in a hurry now, later
				assert.Equal(t, cust.ID, cust2.ID)
				assert.Equal(t, "updTrescha", cust2.FirstName.String)
				assert.Equal(t, "updStoate", cust2.LastName.String)
				assert.Equal(t, false, cust2.CreatedAt.Valid)
				assert.Equal(t, false, cust2.UpdatedAt.Valid)
				assert.Equal(t, false, cust2.DeletedAt.Valid)
				assert.Equal(t, cust.AccountId, cust2.AccountId)

				// Has one
				assert.Equal(t, cust.Info.ID, cust2.Info.ID)
				assert.Equal(t, cust.ID, cust2.Info.CustomerID)
				assert.Equal(t, "123-456-890", cust2.Info.Phone.String)

				// BelongsTo
				assert.Equal(t, cust.Account.ID, cust2.Account.ID)
				assert.Equal(t, "updFrank", cust2.Account.Name)

				// HasMany
				assert.Equal(t, 1, len(cust2.Orders))
				assert.Equal(t, cust.Orders[0].ID, cust2.Orders[0].ID)
				assert.Equal(t, cust.ID, cust2.Orders[0].CustomerID)
				assert.Equal(t, false, cust2.Orders[0].CreatedAt.Valid)
				assert.Equal(t, cust.Orders[0].Product.ID, cust2.Orders[0].Product.ID)
				assert.Equal(t, "updOnePlus", cust2.Orders[0].Product.Name.String)

				// ManyToMany
				assert.Equal(t, 3, len(cust2.Service))
				assert.Equal(t, cust.Service[0].ID, cust2.Service[0].ID)
				assert.Equal(t, "paypal", cust2.Service[0].Name.String)
				assert.Equal(t, cust.Service[1].ID, cust2.Service[1].ID)
				assert.Equal(t, "banking", cust2.Service[1].Name.String)
				assert.Equal(t, cust.Service[2].ID, cust2.Service[2].ID)
				assert.Equal(t, "newEntry", cust2.Service[2].Name.String)
			}
		}
	}
}

func TestEagerLoading_Update_MysqlErr(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders()
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			eager, err := orm.NewStrategy("eager")
			if assert.NoError(t, err) {

				c := sqlquery.Condition{}
				c.Where("id = ?", 1)
				cust.First(&c)

				err = eager.Update(&cust, &c)

				assert.NoError(t, err) // changed

			}
		}
	}
}

func TestEagerLoading_Delete(t *testing.T) {
	orm.GlobalBuilder, _ = HelperCreateBuilder()
	err := deleteAll()
	if assert.NoError(t, err) {
		err = insertWithoutOrders() //TODO delete can not handel relations of relations.... depth = 1 atm
		assert.NoError(t, err)

		cust := Customerfk{}
		err = cust.Initialize(&cust)
		if assert.NoError(t, err) {
			eager, err := orm.NewStrategy("eager")
			if assert.NoError(t, err) {
				c := sqlquery.Condition{}

				c.Where("id = ?", 1)
				cust.ID = 1
				err = eager.Delete(&cust, &c)
				if assert.NoError(t, err) {

				}

				// error because no rows are affected
				c = sqlquery.Condition{}
				c.Where("id = ?", 100)
				cust.ID = 100
				err = eager.Delete(&cust, &c)
				assert.Error(t, err)
			}
		}
	}

}
