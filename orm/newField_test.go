package orm

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Head struct {
	Model

	ID      int
	Brand   string
	RobotID int
	//Brain Brain
}

/*
func (h *Head) GridFieldType(g * grid.Grid) grid.FieldType{
	ft := grid.DefaultFieldType(g)
	ft.SetName("JOOOOO")
	return ft
}
*/

type Owner struct {
	Model

	ID   int
	Name string
}

type Func struct {
	Model

	ID      int
	Name    string
	RobotID int
}

type Parts struct {
	Model

	ID   int
	Name string
}

type PP []Parts

type Robot struct {
	Model

	ID               int
	Name             string `json:"test" orm:"select:CONCAT(name,id)"`
	Age              int    `orm:"permission:rw"`
	OwnerID          int
	FieldNotExisting string `orm:"custom"`

	Head  Head   // hasOne
	Owner Owner  // belongsTo
	Funcs []Func //hasMany
	Parts PP     // m2m

	//Files []File `fk:"field:ID; associationField:RobotIDs"`
	//Files []File `fk:"ID"`
	Files []File
}

type Bot struct {
	Model

	ID               int
	Name             string `json:"test" orm:"select:CONCAT(name,id)"`
	Age              int    `orm:"permission:rw"`
	OwnerID          int
	FieldNotExisting string `orm:"custom"`

	Head  Head   // hasOne
	Owner Owner  // belongsTo
	Funcs []Func //hasMany
	Parts PP     // m2m

	//Files File `fk:"field:ID; associationField:RobotIDs"`
	//Files File `fk:"ID"`
	Files File
}

func (b *Bot) TableName() string {
	return "robots"
}

type File struct {
	Model

	ID      int
	RobotID int
	BotID   int
	unext   int

	Path string
}

func (f *File) Custom() bool {
	return true
}

func (f *File) OrmBeforeFirst(cbk *Callback) error {
	f.ID = 1
	fmt.Println("called FILE BEFORE")
	return nil
}

func (f *File) OrmAfterFirst(cbk *Callback) error {

	fmt.Println("called FILE AFTER First")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	// slice
	if cbk.ResultSet() != nil {
		result := cbk.ResultSet().(*[]File)
		*result = append(*result, File{ID: 12})
	} else {
		f.ID = 12
	}

	return nil
}

func (f *File) OrmAfterAll(cbk *Callback) error {
	fmt.Println("called FILE AFTER ALL")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	//result := cbk.ResultSet().(*[]File)
	//*result = append(*result, File{ID:1})

	p, err := cbk.Parent()
	fmt.Println("Files in parent", err)

	if err == nil {
		robots := *p.ResultSet().(*[]Robot)
		for _, r := range robots {
			fmt.Println("RID ->", r.ID)
		}
	}

	if cbk.ResultSet() != nil {
		result := cbk.ResultSet().(*[]File)
		*result = append(*result, File{ID: 1, RobotID: 4, BotID: 4, Path: "aaaa"})
	}

	return nil
}

func (f *File) OrmAfterCreate(cbk *Callback) error {
	fmt.Println("called FILE AFTER Create")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	fmt.Println(f)

	return nil
}

func (f *File) OrmBeforeUpdate(cbk *Callback) error {
	fmt.Println("called FILE BEFORE Update")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	fmt.Println(f)

	return nil
}
func (f *File) OrmAfterUpdate(cbk *Callback) error {
	fmt.Println("called FILE AFTER Update")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	fmt.Println(f)

	return nil
}

func (f *File) OrmBeforeDelete(cbk *Callback) error {
	fmt.Println("called FILE BEFORE Delete")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	fmt.Println(f.ID)

	return nil
}
func (f *File) OrmAfterDelete(cbk *Callback) error {
	fmt.Println("called FILE AFTER Delete")

	//	f.Table().PrimaryKeys()
	//	f.ValueByFieldName("ID")

	fmt.Println(f.ID)

	return nil
}

func (r *Bot) OrmBefore(cbk *Callback) error {
	fmt.Println("called BOT BEFORE", cbk.Mode())

	r.FieldNotExisting += "added Before" + cbk.Mode()

	//result := cbk.ResultSet().(*[]File)
	//*result = append(*result, File{ID:1})

	return nil
}

func (r *Robot) OrmBefore(cbk *Callback) error {
	fmt.Println("called ROBOT BEFORE", cbk.Mode())
	//p := cbk.Relation()
	r.FieldNotExisting += "added Before" + cbk.Mode()

	//result := cbk.ResultSet().(*[]File)
	//*result = append(*result, File{ID:1})

	return nil
}
func (r *Bot) OrmAfter(cbk *Callback) error {
	fmt.Println("called ROBotBOT AFTER", cbk.Mode())

	r.FieldNotExisting += "added After" + cbk.Mode()
	return nil
}

func (r *Robot) OrmAfter(cbk *Callback) error {
	fmt.Println("called ROBOT AFTER", cbk.Mode())

	r.FieldNotExisting += "added After" + cbk.Mode()
	return nil
}

func TestNewField_First(t *testing.T) {

	robot := &Robot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	// fetch First
	robot.FieldNotExisting = "manually added"

	err = robot.First(nil)
	assert.NoError(t, err)

	fmt.Println("DB Columns: ", robot.Table().columnNames(READDB, false))
	fmt.Println("View Columns: ", robot.Table().columnNames(READVIEW, false))

	for rel, relModel := range robot.Table().Relations(nil, READDB) {
		fmt.Println("Relations DB: ", rel, relModel.StructTable, relModel.Type)
	}

	for rel, relModel := range robot.Table().Relations(nil, READVIEW) {
		fmt.Println("Relations VIEW: ", rel, relModel.StructTable, relModel.Type)
	}

	//fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	//fmt.Println(robot.ID, "-", len(robot.Files), robot.Files)

	//fmt.Printf("%p %p\n", robot.cbk.caller, robot)

}

func TestNewFieldBot_First(t *testing.T) {

	robot := &Bot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	// fetch First
	robot.FieldNotExisting = "manually added"

	err = robot.First(nil)
	assert.NoError(t, err)

	fmt.Println("DB Columns: ", robot.Table().columnNames(READDB, false))
	fmt.Println("View Columns: ", robot.Table().columnNames(READVIEW, false))

	for rel, relModel := range robot.Table().Relations(nil, READDB) {
		fmt.Println("Relations DB: ", rel, relModel.StructTable, relModel.Type)
	}

	for rel, relModel := range robot.Table().Relations(nil, READVIEW) {
		fmt.Println("Relations VIEW: ", rel, relModel.StructTable, relModel.Type)
	}

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", robot.Files.ID)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)
}

func TestNewField_All(t *testing.T) {

	robot := &Robot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	// fetch First
	robot.FieldNotExisting = "manually added"

	var robots []Robot
	err = robot.All(&robots, nil)
	assert.NoError(t, err) //here should be no error but ROBOTID appiers

	fmt.Println("DB Columns: ", robot.Table().columnNames(READDB, false))
	fmt.Println("View Columns: ", robot.Table().columnNames(READVIEW, false))

	for rel, relModel := range robot.Table().Relations(nil, READDB) {
		fmt.Println("Relations DB: ", rel, relModel.StructTable, relModel.Type)
	}

	for rel, relModel := range robot.Table().Relations(nil, READVIEW) {
		fmt.Println("Relations VIEW: ", rel, relModel.StructTable, relModel.Type)
	}

	fmt.Println(robots[0].ID, "-", robots[0].FieldNotExisting)
	fmt.Println(robots[0].ID, "-", len(robots[0].Files), robots[0].Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

}

func TestNewFieldBot_All(t *testing.T) {

	robot := &Bot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	// fetch First
	robot.FieldNotExisting = "manually added"

	var robots []Robot
	err = robot.All(&robots, nil)
	assert.NoError(t, err) // Error 1054: Unknown column 'BotID' in 'where clause'

	fmt.Println("DB Columns: ", robot.Table().columnNames(READDB, false))
	fmt.Println("View Columns: ", robot.Table().columnNames(READVIEW, false))

	for rel, relModel := range robot.Table().Relations(nil, READDB) {
		fmt.Println("Relations DB: ", rel, relModel.StructTable, relModel.Type)
	}

	for rel, relModel := range robot.Table().Relations(nil, READVIEW) {
		fmt.Println("Relations VIEW: ", rel, relModel.StructTable, relModel.Type)
	}

	fmt.Println(robots[0].ID, "-", robots[0].FieldNotExisting)
	fmt.Println(robots[0].ID, "-", len(robots[0].Files), robots[0].Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

}

func TestNewField_Create(t *testing.T) {
	robot := &Robot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	robot.Name = "Test"
	robot.Age = 12
	robot.FieldNotExisting = "manually added"

	robot.Owner = Owner{ID: 1, Name: "Patrick Ascher"}

	robot.Head = Head{Brand: "AAAA"}

	robot.Files = append(robot.Files, File{Path: "ABC"})

	err = robot.Create()
	assert.NoError(t, err)

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", len(robot.Files), robot.Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

}

func TestNewFieldBot_Create(t *testing.T) {
	robot := &Bot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	robot.Name = "Test"
	robot.Age = 12
	robot.FieldNotExisting = "manually added"

	robot.Owner = Owner{ID: 1, Name: "Patrick Ascher"}

	robot.Head = Head{Brand: "AAAA"}

	robot.Files.Path = "ABC"

	err = robot.Create()
	assert.NoError(t, err)

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", robot.Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

}

func TestNewField_Update(t *testing.T) {
	robot := &Robot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)
	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

	err = robot.First(nil)
	assert.NoError(t, err)
	fmt.Printf("%p %p\n", robot.cbk.caller, robot)

	robot.FieldNotExisting = "manually added"
	robot.Name += "-"

	robot.Files = append(robot.Files, File{Path: "DEF"})

	err = robot.Update()

	assert.NoError(t, err)

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", len(robot.Files), robot.Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)
}

func TestNewFieldBot_Update(t *testing.T) {
	robot := &Bot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	err = robot.First(nil)
	assert.NoError(t, err)

	robot.FieldNotExisting = "manually added"
	robot.Name += "-"

	robot.Files.Path = "DEFGHI"

	err = robot.Update()
	assert.NoError(t, err)

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", robot.Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)
}

func TestNewField_Delete(t *testing.T) {
	robot := &Robot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	c := sqlquery.Condition{}
	c.Order("id desc")
	err = robot.First(&c)
	assert.NoError(t, err)

	err = robot.Delete() // entry could not be found (delete) - DELETE FROM `files` WHERE `id` = 12
	assert.NoError(t, err)

	fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	fmt.Println(robot.ID, "-", len(robot.Files), robot.Files)

	fmt.Printf("%p %p\n", robot.cbk.caller, robot)
}

func TestNewFieldBot_Delete(t *testing.T) {
	robot := &Bot{}

	// init
	err := robot.Initialize(robot)
	assert.NoError(t, err)

	c := sqlquery.Condition{}
	c.Order("id desc")
	err = robot.First(&c)
	assert.NoError(t, err)

	err = robot.Delete()
	assert.NoError(t, err) //entry could not be found (delete) - DELETE FROM `files` WHERE `id` = 12

	//fmt.Println(robot.ID, "-", robot.FieldNotExisting)
	//fmt.Println(robot.ID, "-", robot.Files)
	//fmt.Printf("%p %p\n", robot.cbk.caller, robot)
}
