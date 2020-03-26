package orm_test

import (
	"database/sql"
	"github.com/guregu/null"
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/json"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"os"
)

var Strategy = &StrategyMock{}

func init() {
	orm.Register("mock", Strategy)
	orm.GlobalBuilder, _ = HelperCreateBuilder()
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

func HelperCreateBuilder() (*sqlquery.Builder, error) {
	cfg, err := HelperParseConfig()
	if err != nil {
		return nil, err
	}
	b, err := sqlquery.New(cfg, nil)
	orm.GlobalBuilder = &b
	return &b, err
}

type StrategyMock struct {
	model        orm.Interface
	methodCalled string
	c            *sqlquery.Condition
	res          interface{}
}

func (s *StrategyMock) First(m orm.Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "First"
	s.c = c
	return nil
}

func (s *StrategyMock) All(res interface{}, m orm.Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "All"
	s.res = res
	s.c = c
	return nil
}

func (s *StrategyMock) Create(m orm.Interface) error {
	s.model = m
	s.methodCalled = "Create"
	return nil
}

func (s *StrategyMock) Update(m orm.Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "Update"
	s.c = c

	return nil
}

func (s *StrategyMock) Delete(m orm.Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "Delete"
	s.c = c

	return nil
}

type CommonX struct {
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

type Customerfk struct {
	orm.Model

	ID        int
	FirstName sql.NullString
	LastName  sql.NullString
	//CommonX

	Info      Contactfk   // hasOne
	Orders    []Orderfk   // hasMany
	Service   []Servicefk // manyToMany
	Account   Accountfk   // belongsTo
	AccountId int
}

type CustomerNoSoftDelete struct {
	orm.Model

	ID        int
	FirstName sql.NullString
	LastName  sql.NullString
}

func (c *CustomerNoSoftDelete) TableName() string {

	return "customers"
}

type Accountfk struct {
	orm.Model

	ID   int
	Name string
}

type Customer struct {
	orm.Model

	ID        int
	FirstName sql.NullString
	LastName  sql.NullString
	CommonX

	Info   Contact `relation:"hasOne" fk:"ID"`  // hasOne
	Orders []Order `relation:"hasMany" fk:"ID"` // hasMany
	//Service []Service `relation:"manyToMany"` // not working
}

type Orderfk struct {
	orm.Model

	ID         int
	CustomerID int
	CreatedAt  null.Time

	Product Productfk
}

type Order struct {
	orm.Model

	ID         int
	CustomerID int
	ProductID  int
	CreatedAt  null.Time

	Product  Product  `relation:"hasOne" fk:"field:ID;associationField:ProductID"`     // hasOne
	Customer Customer `relation:"belongsTo" fk:"field:ID;associationField:CustomerID"` // belongsTo
}

type Productfk struct {
	orm.Model

	ID        int
	Name      sql.NullString
	Price     sql.NullFloat64
	CreatedAt null.Time
	UpdatedAt null.Time
	OrderId   int
}

type Product struct {
	orm.Model

	ID        int
	Name      sql.NullString
	Price     sql.NullFloat64
	CreatedAt null.Time
	UpdatedAt null.Time
}

type Contactfk struct {
	orm.Model

	ID         int
	CustomerID int
	Phone      sql.NullString
}

type Contact struct {
	orm.Model

	ID         int
	CustomerID int
	Phone      sql.NullString
}

type Servicefk struct {
	orm.Model

	ID   int
	Name sql.NullString
}

type Service struct {
	orm.Model

	ID   int
	Name sql.NullString
}
