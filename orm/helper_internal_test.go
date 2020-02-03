package orm

import (
	"github.com/patrickascher/gofw/config"
	"github.com/patrickascher/gofw/config/reader"
	"github.com/patrickascher/gofw/sqlquery"
	"os"
)

func HelperParseConfig() (*sqlquery.Config, error) {
	var cfg sqlquery.Config
	var err error

	if os.Getenv("TRAVIS") != "" {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/travis." + os.Getenv("DB") + ".json"})
	} else {
		err = config.Parse("json", &cfg, &reader.JsonOptions{File: "tests/db.json"})
	}

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func HelperCreateBuilder() (*sqlquery.Builder, error) {
	cfg, err := HelperParseConfig()

	if err != nil {
		return nil, err
	}
	return sqlquery.NewBuilderFromConfig(cfg)
}

type StrategyMock struct {
	model        Interface
	methodCalled string
	c            *sqlquery.Condition
	res          interface{}
}

func (s *StrategyMock) First(m Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "First"
	s.c = c
	return nil
}

func (s *StrategyMock) All(res interface{}, m Interface, c *sqlquery.Condition) error {
	s.model = m
	s.methodCalled = "All"
	s.c = c
	return nil
}

func (s *StrategyMock) Create(m Interface) error {
	s.model = m
	s.methodCalled = "Create"
	return nil
}

func (s *StrategyMock) Update(m Interface) error {
	s.model = m
	s.methodCalled = "Update"
	return nil
}

func (s *StrategyMock) Delete(m Interface) error {
	s.model = m
	s.methodCalled = "Delete"
	return nil
}

type CommonX struct {
	CreatedAt sqlquery.NullTime
	UpdatedAt sqlquery.NullTime
	DeletedAt sqlquery.NullTime
}

type Customerfk struct {
	Model

	unexp int

	ID        int
	FirstName sqlquery.NullString
	LastName  sqlquery.NullString
	CommonX

	Info    Contactfk   // hasOne
	Orders  []Orderfk   // hasMany
	Service []Servicefk // manyToMany
}

type Customerptr struct {
	Model

	ID        int
	FirstName sqlquery.NullString
	LastName  sqlquery.NullString
	CommonX

	Info    *Contactfk   // hasOne
	Orders  []*Orderfk   // hasMany
	Service []*Servicefk // manyToMany
}

type CustomerNilBuilder struct {
	Model

	ID        int
	FirstName sqlquery.NullString
}

func (c *CustomerNilBuilder) Builder() (*sqlquery.Builder, error) {
	return nil, nil
}

type CustomerBuilder struct {
	Model

	ID        int
	FirstName sqlquery.NullString
}

func (c *CustomerBuilder) Builder() (*sqlquery.Builder, error) {
	return &sqlquery.Builder{}, nil
}

type Customer struct {
	Model

	ID        int
	FirstName sqlquery.NullString
	LastName  sqlquery.NullString
	CommonX

	Info   Contact `relation:"hasOne" fk:"ID"`  // hasOne
	Orders []Order `relation:"hasMany" fk:"ID"` // hasMany
	//Service []Service `relation:"manyToMany"` // not working
}

type Orderfk struct {
	Model

	ID         int
	CustomerID int
	CreatedAt  sqlquery.NullTime

	Product  Productfk
	Customer Customerfk
}

type Order struct {
	Model

	ID         int
	CustomerID int
	CreatedAt  sqlquery.NullTime

	Product  Product  `relation:"hasOne" fk:"field:OrderID;associationField:ID"`       // hasOne
	Customer Customer `relation:"belongsTo" fk:"field:ID;associationField:CustomerID"` // belongsTo
}

type Productfk struct {
	Model

	ID        int
	Name      sqlquery.NullString
	Price     sqlquery.NullFloat64
	CreatedAt sqlquery.NullTime
	UpdatedAt sqlquery.NullTime
	OrderID   int
}

type Product struct {
	Model

	ID        int
	Name      sqlquery.NullString
	Price     sqlquery.NullFloat64
	CreatedAt sqlquery.NullTime
	UpdatedAt sqlquery.NullTime
	OrderID   int
}

type Contactfk struct {
	Model

	ID         int
	CustomerID int
	Phone      sqlquery.NullString
}

type Contact struct {
	Model

	ID         int
	CustomerID int
	Phone      sqlquery.NullString
}

type Servicefk struct {
	Model

	ID   int
	Name sqlquery.NullString
}

type Service struct {
	Model

	ID   int
	Name sqlquery.NullString
}
