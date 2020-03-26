package orm_test

import (
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	_ "github.com/patrickascher/gofw/sqlquery/driver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModel_Initialize(t *testing.T) {

}
func TestModel_First(t *testing.T) {

	customer := Customerfk{}

	// not initialized
	err := customer.First(nil)
	if assert.Error(t, err) {
		assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
	}

	err = customer.Initialize(&customer)
	if assert.NoError(t, err) {
		customer.SetStrategy("mock")
		err = customer.First(nil)
		if assert.NoError(t, err) {
			assert.Equal(t, "First", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
			assert.Equal(t, &sqlquery.Condition{}, Strategy.c)
		}

		c := &sqlquery.Condition{}
		c.Where("test IS NULL")
		err = customer.First(c)
		if assert.NoError(t, err) {
			assert.Equal(t, "First", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
			assert.Equal(t, c, Strategy.c)
		}
	}
}
func TestModel_All(t *testing.T) {
	customer := Customerfk{}

	var res []Customerfk
	// not initialized
	err := customer.All(res, nil)
	if assert.Error(t, err) {
		assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
	}

	err = customer.Initialize(&customer)
	if assert.NoError(t, err) {
		customer.SetStrategy("mock")
		err = customer.All(res, nil)
		if assert.NoError(t, err) {
			assert.Equal(t, "All", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
			assert.Equal(t, &sqlquery.Condition{}, Strategy.c)
			assert.Equal(t, res, Strategy.res)

		}

		c := &sqlquery.Condition{}
		c.Where("test IS NULL")
		err = customer.All(res, c)
		if assert.NoError(t, err) {
			assert.Equal(t, "All", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
			assert.Equal(t, c, Strategy.c)
			assert.Equal(t, res, Strategy.res)
		}
	}
}
func TestModel_Create(t *testing.T) {
	customer := Customerfk{}

	// not initialized
	err := customer.Create()
	if assert.Error(t, err) {
		assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
	}

	err = customer.Initialize(&customer)
	if assert.NoError(t, err) {
		customer.SetStrategy("mock")
		customer.FirstName.String = "abc"
		customer.FirstName.Valid = true

		customer.LastName.String = "def"
		customer.LastName.Valid = true
		err = customer.Create()
		if assert.NoError(t, err) {
			assert.Equal(t, "Create", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
		}

		// 2nd call to check if the old transaction gets deleted after committing
		err = customer.Create()
		if assert.NoError(t, err) {
			assert.Equal(t, "Create", Strategy.methodCalled)
			assert.Equal(t, &customer, Strategy.model)
		}
	}
}
func TestModel_Update(t *testing.T) {
	customer := Customerfk{}

	// not initialized
	err := customer.Update()
	if assert.Error(t, err) {
		assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
	}

	err = customer.Initialize(&customer)
	if assert.NoError(t, err) {
		customer.SetStrategy("mock")
		err = customer.Update()

		// error because the primary mandatory fields are empty
		assert.Error(t, err)

		// Everything OK
		customer.ID = 5
		err = customer.Update()
		assert.NoError(t, err)
		//		assert.Equal(t, &customer, Strategy.model)
		assert.Equal(t, "Update", Strategy.methodCalled)
		cExp := &sqlquery.Condition{}
		b, _ := customer.Builder()

		cExp.Where(b.QuoteIdentifier("id")+" = ?", 5)
		assert.Equal(t, cExp, Strategy.c)

		// call update again to check if the tx will get handled correctly
		customer.ID = 5
		err = customer.Update()
		assert.NoError(t, err)
		assert.Equal(t, &customer, Strategy.model)
		assert.Equal(t, "Update", Strategy.methodCalled)
		cExp = &sqlquery.Condition{}
		cExp.Where(b.QuoteIdentifier("id")+" = ?", 5)
		assert.Equal(t, cExp, Strategy.c)
	}
}

func TestModel_Delete_SoftDelete(t *testing.T) {
	err := deleteAll()
	if assert.NoError(t, err) {
		customer := Customerfk{}

		// not initialized
		err := customer.Delete()
		if assert.Error(t, err) {
			assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
		}

		err = customer.Initialize(&customer)
		if assert.NoError(t, err) {
			customer.SetStrategy("mock")
			err = customer.Delete()

			// error because the primary mandatory fields are empty
			assert.Error(t, err)

			// error because zero rows were affected
			customer.ID = 5
			err = customer.Delete()
			// assert.Error(t, err)

			assert.NoError(t, err) // not anymore
		}
	}
}

func TestModel_Delete(t *testing.T) {
	customer := CustomerNoSoftDelete{}

	// not initialized
	err := customer.Delete()
	if assert.Error(t, err) {
		assert.Equal(t, orm.ErrModelNotInitialized.Error(), err.Error())
	}

	err = customer.Initialize(&customer)
	if assert.NoError(t, err) {
		customer.SetStrategy("mock")
		err = customer.Delete()

		// error because the primary mandatory fields are empty
		assert.Error(t, err)

		// error because zero rows were affected
		customer.ID = 5
		err = customer.Delete()
	}
}

func TestModel_Count(t *testing.T) {
	customer := Customerfk{}

	// not initialized
	count, err := customer.Count(nil)
	assert.Error(t, err)
	assert.Equal(t, 0, count)

	// ok
	err = customer.Initialize(&customer)
	assert.NoError(t, err)
	count, err = customer.Count(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestModel_Table(t *testing.T) {

}

/*
type Common struct {
	DeletedAt sqlquery.NullTime
}

type Global struct {
	Model

	ID   int
	Test sqlquery.NullString
	//UserID User
}

type Post struct {
	Model

	ID   int
	Post null.String
	//UserID User
}

type History struct {
	Model

	ID     int
	UserID int
	Text   null.String
}

type Address struct {
	Model

	ID      int
	UserID  int
	Street  null.String
	Zip     null.Int
	Country null.String

	//User   User // BelongsTo
}

type User struct {
	Model

	ID          int `orm:"column:id;permission:rw"`
	Name        string
	notExported bool

	Adr      Address    //`fk:"table:"` //hasOne
	History  []*History //hasmany
	Posts    []Post     //m2m (auto check if user_posts table + 2 fields user_id, post_id), or defined in tag
	Global   *Global
	GlobalId int

	Common
}

func TestModel_Initialize(t *testing.T) {

	var err error
	GlobalBuilder, err = helperBuilderMysql()
	assert.NoError(t, err)

	start := time.Now()
	u := User{}
	err2 := u.Initialize(&u)
	fmt.Println(u.SoftDelete())

	assert.NoError(t, err2)

	fmt.Println(u.table.Associations["Adr"])
	elapsed := time.Since(start)
	fmt.Println("INIT took ", elapsed)

	fmt.Println("-----------------------")

	start = time.Now()
	u2 := User{}
	err2 = u2.Initialize(&u2)
	elapsed = time.Since(start)
	fmt.Println("Cache took ", elapsed)

	//Global
	assert.Equal(t, BelongsTo, u2.table.Associations["Global"].Type)
	assert.Equal(t, "users", u2.table.Associations["Global"].StructTable.Information.Table)
	assert.Equal(t, "global_id", u2.table.Associations["Global"].StructTable.Information.Name)
	assert.Equal(t, "globals", u2.table.Associations["Global"].AssociationTable.Information.Table)
	assert.Equal(t, "id", u2.table.Associations["Global"].AssociationTable.Information.Name)

	//Address
	assert.Equal(t, HasOne, u2.table.Associations["Adr"].Type)
	assert.Equal(t, "users", u2.table.Associations["Adr"].StructTable.Information.Table)
	assert.Equal(t, "id", u2.table.Associations["Adr"].StructTable.Information.Name)
	assert.Equal(t, "addresses", u2.table.Associations["Adr"].AssociationTable.Information.Table)
	assert.Equal(t, "user_id", u2.table.Associations["Adr"].AssociationTable.Information.Name)

	//HistoryID
	assert.Equal(t, HasMany, u2.table.Associations["History"].Type)
	assert.Equal(t, "users", u2.table.Associations["History"].StructTable.Information.Table)
	assert.Equal(t, "id", u2.table.Associations["History"].StructTable.Information.Name)
	assert.Equal(t, "histories", u2.table.Associations["History"].AssociationTable.Information.Table)
	assert.Equal(t, "user_id", u2.table.Associations["History"].AssociationTable.Information.Name)

	// Posts
	assert.Equal(t, ManyToMany, u2.table.Associations["Posts"].Type)
	assert.Equal(t, "users", u2.table.Associations["Posts"].StructTable.Information.Table)
	assert.Equal(t, "id", u2.table.Associations["Posts"].StructTable.Information.Name)
	assert.Equal(t, "posts", u2.table.Associations["Posts"].AssociationTable.Information.Table)
	assert.Equal(t, "id", u2.table.Associations["Posts"].AssociationTable.Information.Name)
	assert.Equal(t, "user_posts", u2.table.Associations["Posts"].JunctionTable.Table)
	assert.Equal(t, "user_id", u2.table.Associations["Posts"].JunctionTable.StructColumn)
	assert.Equal(t, "post_id", u2.table.Associations["Posts"].JunctionTable.AssociationColumn)

	fmt.Println(u.SoftDelete())
	assert.NoError(t, err2)

	fmt.Println("-----------------------")

	start = time.Now()
	u3 := User{}
	err3 := u3.Initialize(&u3)
	assert.Equal(t, 0, u3.ID)

	err = u3.First(nil)
	assert.Equal(t, 1, u3.ID)

	assert.Equal(t, 1, u3.Adr.ID)
	assert.Equal(t, 1, u3.Adr.UserID)
	assert.Equal(t, "Obere Feld", u3.Adr.Street.String)
	assert.Equal(t, int64(6500), u3.Adr.Zip.Int64)
	assert.Equal(t, "AT", u3.Adr.Country.String)

	assert.Equal(t, 1, u3.Global.ID)
	assert.Equal(t, "test", u3.Global.Test.String)

	assert.Equal(t, 2, len(u3.Posts))

	assert.NoError(t, err)

	elapsed = time.Since(start)
	fmt.Println("Cache took ", elapsed)

	//Global
	assert.Equal(t, BelongsTo, u3.table.Associations["Global"].Type)
	assert.Equal(t, "users", u3.table.Associations["Global"].StructTable.Information.Table)
	assert.Equal(t, "global_id", u3.table.Associations["Global"].StructTable.Information.Name)
	assert.Equal(t, "globals", u3.table.Associations["Global"].AssociationTable.Information.Table)
	assert.Equal(t, "id", u3.table.Associations["Global"].AssociationTable.Information.Name)

	//Address
	assert.Equal(t, HasOne, u3.table.Associations["Adr"].Type)
	assert.Equal(t, "users", u3.table.Associations["Adr"].StructTable.Information.Table)
	assert.Equal(t, "id", u3.table.Associations["Adr"].StructTable.Information.Name)
	assert.Equal(t, "addresses", u3.table.Associations["Adr"].AssociationTable.Information.Table)
	assert.Equal(t, "user_id", u3.table.Associations["Adr"].AssociationTable.Information.Name)

	//HistoryID
	assert.Equal(t, HasMany, u3.table.Associations["History"].Type)
	assert.Equal(t, "users", u3.table.Associations["History"].StructTable.Information.Table)
	assert.Equal(t, "id", u3.table.Associations["History"].StructTable.Information.Name)
	assert.Equal(t, "histories", u3.table.Associations["History"].AssociationTable.Information.Table)
	assert.Equal(t, "user_id", u3.table.Associations["History"].AssociationTable.Information.Name)

	// Posts
	assert.Equal(t, ManyToMany, u3.table.Associations["Posts"].Type)
	assert.Equal(t, "users", u3.table.Associations["Posts"].StructTable.Information.Table)
	assert.Equal(t, "id", u3.table.Associations["Posts"].StructTable.Information.Name)
	assert.Equal(t, "posts", u3.table.Associations["Posts"].AssociationTable.Information.Table)
	assert.Equal(t, "id", u3.table.Associations["Posts"].AssociationTable.Information.Name)
	assert.Equal(t, "user_posts", u3.table.Associations["Posts"].JunctionTable.Table)
	assert.Equal(t, "user_id", u3.table.Associations["Posts"].JunctionTable.StructColumn)
	assert.Equal(t, "post_id", u3.table.Associations["Posts"].JunctionTable.AssociationColumn)

	fmt.Println(u.SoftDelete())
	assert.NoError(t, err3)

	var users []User
	test3 := User{}
	err4 := test3.Initialize(&test3)
	if assert.NoError(t, err4) {
		err5 := test3.All(&users, nil)
		assert.NoError(t, err5)

		assert.Equal(t, 2, len(users))

		assert.Equal(t, 1, users[0].ID)
		assert.Equal(t, "Wall-E", users[0].Name)
		assert.Equal(t, 2, len(users[0].Posts))
		assert.Equal(t, "Obere Feld", users[0].Adr.Street.String)
		assert.Equal(t, "test", users[0].Global.Test.String)
		assert.Equal(t, 2, len(users[0].History))

		assert.Equal(t, 3, users[1].ID)
		assert.Equal(t, "Ascher", users[1].Name)
		assert.Equal(t, 1, len(users[1].Posts))
		//assert.Equal(t,(*Address)(nil), reflect.ValueOf(users[1].Adr).IsNil())
		assert.Equal(t, "3test3", users[1].Global.Test.String)
		assert.Equal(t, 0, len(users[1].History))
		fmt.Println(users[0])
		fmt.Println(users[1])

	}

	// create user
	createUser := &User{}
	err = createUser.Initialize(createUser)
	assert.NoError(t, err)
	createUser.Name = "LOL"
	createUser.GlobalId = 1

	//createUser.Global = &Global{}
	//createUser.Global.Test = null.StringFrom("Test3")

	createUser.History = append(createUser.History, &History{Text: null.StringFrom("#H1")})
	createUser.History = append(createUser.History, &History{Text: null.StringFrom("#H2")})

	createUser.Posts = append(createUser.Posts, Post{Post: null.StringFrom("azebenja")})
	createUser.Posts = append(createUser.Posts, Post{})

	err = createUser.Create()
	assert.NoError(t, err)

	fmt.Println(createUser)
}

func TestModel_Update(t *testing.T) {
	// create user
	updateUser := &User{}
	err := updateUser.Initialize(updateUser)
	assert.NoError(t, err)
	updateUser.ID = 78
	//updateUser.GlobalId=1
	updateUser.Name = "aalfsdaa"
	updateUser.Adr = Address{Street: null.StringFrom("updateds OBE")}
	//	updateUser.Global = &Global{ID:56,Test:null.StringFrom("JOOH123O")}

	updateUser.History = append(updateUser.History, &History{Text: null.StringFrom("aaaaaaa")})
	updateUser.History = append(updateUser.History, &History{ID: 62, Text: null.StringFrom("bbbbbbb")})
	updateUser.History = append(updateUser.History, &History{Text: null.StringFrom("cccccccc")})

	updateUser.Posts = append(updateUser.Posts, Post{ID: 1000039, Post: null.StringFrom("test-a")})
	updateUser.Posts = append(updateUser.Posts, Post{Post: null.StringFrom("test-b")})

	err = updateUser.Update()
	assert.NoError(t, err)

	c := &sqlquery.Condition{}

	updateUser.First(c.Where("id = ?", 78))
	fmt.Println(updateUser)

}

func TestModel_Delete(t *testing.T) {
	// create user
	createUser := &User{}
	err := createUser.Initialize(createUser)
	assert.NoError(t, err)
	createUser.ID = 80
	err = createUser.Delete()
	assert.NoError(t, err)

}
*/
