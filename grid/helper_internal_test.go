package grid

import (
	"github.com/guregu/null"
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/json"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"os"
)

func deleteAll() error {

	builder, err := HelperCreateBuilder()
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

	return nil
}

func HelperParseConfig() (sqlquery.Config, error) {
	var cfg sqlquery.Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.New("json", &cfg, json.Options{Filepath: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.New("json", &cfg, json.Options{Filepath: "tests/db.json"})
	}

	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func HelperCreateBuilder() (sqlquery.Builder, error) {
	cfg, err := HelperParseConfig()

	if err != nil {
		return sqlquery.Builder{}, err
	}
	return sqlquery.New(cfg, nil)
}

type Customerfk struct {
	orm.Model

	unexp int

	ID        int
	FirstName null.String
	LastName  null.String
	AccountId int

	Info    Contactfk   // hasOne
	Orders  []Orderfk   // hasMany
	Service []Servicefk // manyToMany
	Account Accountfk   // belongsTo
}

type Accountfk struct {
	orm.Model

	ID   int
	Name string
}

type Orderfk struct {
	orm.Model

	ID         int
	CustomerID int
	CreatedAt  null.Time
}

type Contactfk struct {
	orm.Model

	ID         int
	CustomerID int
	Phone      null.String
}

type Servicefk struct {
	orm.Model

	ID   int
	Name null.String
}
