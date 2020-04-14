package orm2_test

import (
	"fmt"
	"github.com/patrickascher/gofw/orm2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"

	_ "github.com/patrickascher/gofw/orm2/strategy"
)

type car struct {
	orm2.Model

	ID      int
	OwnerID int
	Brand   string

	// relation tests (belongsTo, m2m, hasMany)
	Owner  *owner   `orm:"relation:belongsTo"`
	Driver []driver `orm:"relation:m2m"`
	Wheels []wheel

	// polymorphic tests (hasOne, hasMany)
	Radio  radio    `orm:"polymorphic:Car;polymorphic_value:radio"`
	Liquid []liquid `orm:"polymorphic:Car;polymorphic_value:liquid"`
}

type owner struct {
	orm2.Model

	ID   int
	Name string

	// relation test hasOne
	//Car car
}

type driver struct {
	orm2.Model

	ID   int
	Name string
}

type wheel struct {
	orm2.Model

	ID    int
	Brand string
	Note  string

	CarID int
}
type radio struct {
	orm2.Model
	Component
}

func (r radio) DefaultTableName() string {
	return "components"
}

type liquid struct {
	orm2.Model
	Component
}

func (l liquid) DefaultTableName() string {
	return "components"
}

type Component struct {
	ID    int
	Brand string
	Note  string

	CarID   int
	CarType string
}

func TestModel_Init(t *testing.T) {
	test := assert.New(t)

	start := time.Now()
	r := car{}
	err := r.Init(&r)
	test.NoError(err)

	err = r.First(nil)
	test.NoError(err)
	test.Equal(1, r.ID)
	test.Equal("BMW", r.Brand)
	fmt.Println(time.Since(start))
	test.Equal(1, r.Owner.ID)
	test.Equal("Ascher", r.Owner.Name)

	// cached
	start = time.Now()
	r = car{}
	err = r.Init(&r)
	test.NoError(err)

	r.SetWBList(orm2.BLACKLIST, "ID", "OwnerID", "Brand", "Radio.Brand")
	err = r.First(nil)
	test.NoError(err)
	test.Equal(1, r.ID)
	test.Equal(1, r.Radio.ID)
	test.Equal(1, r.Owner.ID)

	test.Equal("radio", r.Radio.CarType)
	test.Equal(1, r.Radio.CarID)

	fmt.Println(time.Since(start))
}
