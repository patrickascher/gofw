<h1>go-orm</h1>

The orm transfers a struct into an ORM by simple embedding the `go_orm.Model` struct.

* A struct must have a field with the name `ID` which must be the primary key in the database table.
* If the field `CreatedAt`,`UpdatedAt` or `DeletedAt` exists in the database, the timestamp will get set automatically. Also a soft delete will happen instead of a real delete. The fields are added at the end of the struct.
* Embedded Structs are allowed
* Automatic relation detection

!> Tests for postgres are disabled atm because the version of travis != local version and an error occurs.
Time is different in mysql +UTC than in postgres +0000.

## Install

```go
go get github.com/fullhouse-productions/go-orm
```

## Usage

```go
type User struct{
	go_orm.Model
	
	ID int
	Name null.String
}

user := &User{}
err := user.Initialize(user)
if err != nil{
	// ...
}

user.First()
```


# Cache
Returns the actual cache.
If no custom cache was set by method, the DefaultCache of the caller model will be called.

```go
user := &User{}
cache, ttl, err := user.Cache()

```

# SetCache
Must be called before the model gets initialized. 

```go

c, err := cache.Get("memory", 5*time.Minute)
if err !=nil{
	//...
}

user := &User{}
err := user.SetCache(c,6*time.Hour)

```

# Initialize

To convert a simple struct into a full working ORM you have to do two things.
* Embedd the `go_orm.Model`
* call `Initialize` on it

```go
type MyStruct struct{
	go_orm.Model
	
	ID int
	//....
}
myStruct := MyStruct{}
myStruct.Initialize(&myStruct)

// orm ready
```

**Workflow:**

* Initialize is setting the caller
* Checking if the struct is already cached (if so, return from cache)
* Validator is added
* Initialize the DB Table
    * Initialize default builder 
        * Call the struct `Builder`
        * Check if builder is defined
    * Call the struct `DatabaseName` and check if a specific database name is defined, otherwise use the one from the config
    * Call the struct `TableName`
    * Create and set the `go_orm.Table` struct to the model
    * Set the loading strategy `Eager` (atm hardcoded, already prepaired to make it config able)
    * Set variable `loadedRel` - to avoid struct loops
    * Add all exported struct fields as `go_orm.Column`
    * Describe the database table with the needed columns.
        * Add db table column information to struct (name, position, nullable, primarykey, type, defaultvalue, length, autoincrement)
        * RETURN ERROR IF STRUCT FIELD ID IS MISSING OR IS NO PRIMARYKEY.
* Initialize Relations
* Add information that the struct is initialized
* Add to cache



# Defaults

A model can be configured by struct methods. All of the following configurations are optional and are already defined by default.

## DefaultCache

The DefaultCache method has to return `cache.Cache, time.Duration, error`.
By default it checks of `go_server.Cache()` is defined, if not a new `memory` cache is used with the duration of `6 hours`.

```go
// Default value
func (m *YourStruct) DefaultCache() (cache.Cache, time.Duration, error) {
	c, err := cache.Get("memory", 5*time.Minute)
	return c, 6 * time.Hour, err
}

```

!> It's highly recommended that a cache is uses, otherwise the whole `Initialize` workflow will be called over and over.

## DatabaseName

By default the database name from the sql connection is taken (if defined).
If a struct is depending on a different database, it can be defined by returning the database name as string.

```go
func (m *YourStruct) DatabaseName() string {
	return "tests"
}
```

## TableName

By default the sql table name of a struct will be the struct name in snake_case pluralized.
Example to set a different table name, simply return the name as string.

```go
func (m *YourStruct) TableName() string {
	return "new_table_name"
}
```


## Builder

By default its checking if the global variable `GlobalBuilder` is defined. This is useful because you can define a global db connection for the whole project.
Anyway, if you need a different connection you can define a `sqlquery.Builder`. It must return a `(*sqlquery.Builder, error)`.

Example of the default value:
 ```go
// Builder returns the GlobalBuilder.
// If it's not defined, a error will return.
func (m *YourStruct) Builder() (*sqlquery.Builder, error) {
	if GlobalBuilder == nil {
		return nil, ErrNoBuilder
	}
	return GlobalBuilder, nil
}

 ```


# Embedded Struct

It is possible to embed a struct to the model. 
For example this can be used to define some common fields.

**Example with common fields `CreatedAt`, `UpdatedAt` and `DeletedAt`**

```go

type Common struct{
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

type User struct{
	model.Model
	ID int
	Name string
	Common
}
```

PS: There is no use to add these fields, because these are added through the model automatic.

!>Keep in mind, the fields of the embed struct are added to the `User` struct. Which means the database table `users` must contain these, otherwise an error will return.


# Tags

Fot a struct the following tags can get defined

| Tag | Key | Value | Description |
|-----|-----|-------|-------------|
| orm    | -    |       | skip the field            |
|     | column    | {name}      | set a different db column name            |
|   |permission     | rw      | default read and write is defined            |
|     | custom    |       | defines that this field or relation is a custom type and gets excluded from all DB queries.           |
|     |select     | sql stmt for the column    | this is only taking care of in `First` and `All`.  example column name: `CONCAT(id,name)` will get set as `(CONCAT(id,name)) as name`. The return value must match the struct field type.            |
|relation     | `hasOne` `belongsTo` `hasMany` `manyToMany`     |       | relation can be one of the values           |
|fk     | only the value    |`fk:"ID"`      | This would mean primary struct Field `ID` + `PrimaryStructNameID` (see more in Relations)         |
|fk     | `field` and `associationField`    | `fk:"field:ID; associationField:UserID"`     | set the fk field (see more in Relations)            |
|validate | || see https://godoc.org/gopkg.in/go-playground/validator.v9|

```go

type User struct{
    // ...
	Name string `model:"column:firstName;permission:rw" validate:"required"`
	Adr Address `relation:"hasOne" fk:"field:ID; associationField:UserID"`
    // ...
}
```

!> ManyToMany is not working with alone with tags at the moment. Also if there just exists a junction table, the code will throw an error because fk are missing. These can not be set on m2m yet.

!> BelongsTo is not working correctly yet with tags.

# Validation

One or more validation can get set over the struct field tag. In the back (https://godoc.org/gopkg.in/go-playground/validator.v9) is used. 
At the moment only tag validators are allowed. By default, validators are added automatically by the db type and relation.
To make it possible for the user to add his own validator by function, `appendConfig` must be exported.

```go
	Name string `validate:"required"`
```

# Relations

A relation will get defined by field type or field tag.
The relation can be a value a pointer or a slice. 

!> If you are using a slice use a slice-to-value or slice-to-pointer but no ptr-to-slice.

The ORM supports by default a infinity loop detector with the counter of 10.
This can not be disabled at the moment, (in the background only a method has to get created for thie `loopDetection` bool)
The solution is not clean, so this will get changed in the future. maybe its getting a own validator?

## HasOne

**By Default** 

The field type address will be recognized as a `HasOne` if the DB table `addresses` column `user_id` has a FK to the table `users` on column `id`

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	//...
}

type User struct{
    go_orm.Model

    ID int
    Adr Address
}
```

**By Tag**

If you don't have a database foreign key you can also define a `HasOne` relation on your own.

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	//...
}

type User struct{
    go_orm.Model

    ID int
    Adr Address `relation:"hasOne" fk:"ID"` 
    //The shortcut of the tag fk is the same as fk:"field:ID; associationField:UserID"
}
```

## BelongsTo

**By Default** 

The struct `Addresses.Usr` will be recogniced as `BelongsTo` relation if a foreign key is defined in `addresses` column `user_id` on the db table `users` column `id`

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	Usr User
	//...
}

type User struct{
    go_orm.Model
    ID int
}
```

**By Tag**

If you don't have a database foreign key you can also define a `BelongsTo` relation on your own.

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	Usr User `relation:"belongsTo" fk:"ID"` 
	//The shortcut of the tag fk is the same as fk:"field:UserID; associationField:ID"
	//...
}

type User struct{
    go_orm.Model
    ID int
    //...
}
```

!> if a belongsTo relation is used, ID field must be provided. Ex: Address belongs to a User but must also contain the UserID field. This is needed because this key has to get updated!

## HasMany

**By Default** 
The struct `User.Adr` will be recogniced as `HasMany` relation if a foreign key is defined in `addresses` column `user_id` on the db table `users` column `id`

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	//...
}

type User struct{
    go_orm.Model
    ID int
    Adr []Address
    //...
}
```


**By Tag**

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	//...
}

type User struct{
    go_orm.Model
    ID int
    Adr []Address `relation:"hasMany" fk:"ID"`
    //The shortcut of the tag fk is the same as fk:"field:ID; associationField:UserID"
    //...
}
```


## ManyToMany , ManyToManySR

**By Default** 
There must be a junction table with the name `user_addresses` (which is the two struct names in snake_case and pluralized ex: UserPost = checking table name -> user_posts).
The junction table must have two foreign keys which are pointing to the struct. The fields must exist in the struct.

```go
type Address struct{
	go_orm.Model
	ID int
	UserID int
	//...
}

type User struct{
    go_orm.Model
    ID int
    Adr []Address
    //...
}
```

?> ManyToManySR will be set if the many to many relation is self referencing.

!> At the moment the junction table or FK can not be set by tag!

# CRUD Methods

All operations are called on the underlaying loading strategy. At the moment only `Eager` is defined and used as default.

## First

First request a single row by its condition and fills the struct fields.
Is not using a transaction and atm not checking if a custom tx was added.

```go
m := &Model{}
err := m.Initialize(m)
//...
c:=&Condition{}
m.First(c)
```

?> Eager Strategy: (Status OK) HasMany and ManyToMany relations are combined to one request each for better performance.
?> At the moment no error will return if there is no result set.... needed?

## All

All gets all results and fills the result slice.
Is not using a transaction and atm not checking if a custom tx was added.

```go
m := &Model{}
var result []Model
err := m.Initialize(m)
//...
c:=&Condition{}
m.All(result,c)
```

?> Eager Strategy: (Status improvements possible) ManyToMany relations is requesting for each result a query.

## Create

Create a new database entry. 
All SQL statements are executed in a transaction. 
You can define a user defined transaction, just be aware of committing it also on your own.
If there is a autoincrement field, the value will be set after the insert automatic in the struct.
 
Is using a transaction by default.

```go
m := &Model{}
err := m.Initialize(m)
//...
m.Name = "Wall-E"
m.Year = 2002
err := m.Create()
```

?> Eager Strategy: (Status improvements possible) HasMany & ManyToMany relations are inserted one per sql insert. This could be batched. Problem is the last inserted id at the moment. Solution last_inserted id which would be the first of the batch + affected rows?
ManyToMany Relations are adding also the association entry if the primary key of the entry is a zero value. If not, only the junction table will be filled.


## Update 

Update the struct values by its primary key(s).
Is using a transaction by default.

With a update, a snapshot will get created for the comprising new vs old data.
If nothing changed, no update will get performed. Its checking the normal fields and all relations.
There are some improvements with the hasMany return values of the equal function that you can see what exactly changed, but its not needed atm.
If there are changes, only the changed fields will get updated by a whitelist.
The snapshot can get disabled by `model.DisableSnapshot(true)`

```go
m := &Model{}
err := m.Initialize(m)
//...
m.ID = 1
m.Name = "Wall-E"
m.Year = 2002
err := m.Update()
```

?> Edit a primary key is destroying the update statement at the moment. Solution `m.Update(id)`?

!> Update has problems with a depth >1 relation and hasMany. There it deletes everything first, and there could be some fk errors.
Solution is to just delete and update, create the correct data with a correct model loop. for that i need a snapshot first. For this first,all must be performant!

?> Eager Strategy: (Status improvements possible)
ManyToMany Relations are adding also the association entry if the primary key of the entry is a zero value. If not, only the junction table will be filled.
If a new hasOne Relation is added, delete old one? Case, if a Struct is build without a `.First()` before, a user can add a different hasOne ID und it would create a new entry in the db. 
HasMany logic is at the moment delete * for this id and then add it again. Also it is creating for each entry a sql statement.
ManyToMany source has a duplicated code part DRY.

?> Eager Strategy: Returns an error if no rows were affected. Problem with `mysql` is if the data did not change a error will return as well in this case.Solution change real_connection type in mysql or its automatic solved when i implement a snapshot logic.


## Delete

Delete a entry by its primary key(s).
Returns an error if no rows were affected.

Is using a transaction by default.


```go
m := &Model{}
err := m.InitModel(m)
//...
c:=&Condition{}
m.ID=2
err := m.Delete()
```

?> Eager Strategy: (Status OK) Relations are getting deleted with one sql statement each. Maybe problems if we introduce callbacks later?
Relations: HasOne and HasMany are deleted completly. ManyToMany is only deleting the junction table entry. BelongsTo is ignored. 

?> Eager Strategy: Returns an error if no rows were affected.

!> Delete has a problem with relations of relations. At the moment only the depth of 1 is taken care of. Solve with a complete model delete instead of builder delete.

## Count

Calculation all rows with the given condition.

I am  using a second select because this is in the most cases faster than SQL_CALC_FOUND_ROWS.

Count is defined in the model and not in the loading strategies like `First`, `All`, `Create`, `Update` and `Delete`.

```go
m := &Model{}
err := m.Initialize(m)
//...
c:=&Condition{}
int,err := m.Count(c)
```

## Timestamp

By default a `CreatedAt`, `UpdatedAt` and `DeletedAt` field is added by the `go_orm.Model` with a `WRITE` permission.
If the fields does not exist in the database table, the `WRITE` and `READ` permission is set to false.
The fields are getting added at the end of the model.

If the fields exist, the actual timestamp is added to the field when
 - The model created function is called the `CreatedAt` will get set. (`UpdaedAt` and `DeletedAt` `WRITE` permission is set to false)
 - The model gets updated the `UpdatedAt` will get set. (`CreatedAt` and `DeletedAt` `WRITE` permission is set to false)
 - The model gets deleted, the `DeletedAt` will get set. (`CreatedAt` and `UpdaedAt` `WRITE` permission is set to false)

These fields are excluded from the `snapshot` of the model.


# DisableSnapshot
Disables the update snapshot which is enabled by default.

```go
m.DisableSnapshot(true)
```

# DisableCustomSql
Disables the custom sql columns. Needed for query manipulation.

```go
m.DisableCustomSql(true)
```

# DisableCallback
Disables the orm callbacks `Before`,`After`,....

```go
m.DisableCallback(true)
```



# SetRelationCondition

To set a special condition to a relation this method can be used.

```go
//...
err = m.SetRelationCondition("Option",&sqlquery.Condition{...})
if err != nil{
	//...
}
m.Update()
```

# Whitelist

A whitelist sets explicit fields/relations to the CRU (Delete is not included into the white or blacklist).
The string has to be the exact field name in the struct. 
It is working for relation and normal fields.

A relation can be whitelisted completely or specific fields with a "." notation.
A Whitelist is in the background overwriting the `column.permission` of the field and sets the read, write booleans.


```go
//...
err = m.Whitelist("Name","Email","Options.Name")
if err != nil{
	//...
}
m.Update()
```

?> **READ** The primary and reference keys are added automatically to avoid sql problems.

?> **WRITE** BelongsTo relations are added automatically to avoid breaking fk constraints.

!> Be careful when u are using the Blacklist, there is no check at the moment if you remove a reference field. this will get fixed soon. Till then, an error will get thrown.


!> check if a reset is needed on a real world application?

!> **Ideas:** If a whitelist or blacklist is set and a field which is not nullabled in the database, is excluded, the crud will throw an error. To avoid this, add this fields automatically? Or create a Warning log. Also create a Notice Log if some PKEYS or REFERENCE Keys were added automatically.

# Blacklist

Same as `Whitelist` just the other way around.
They can not be used together, the last defined one has the priority.

```go
//...
err = m.Blacklist("Name","Email","Options.Name")
if err != nil{
	//...
}
m.Update()
```


# Table information

Because it's maybe in a later point important to now some information about the database table and columns, this function return the `go_orm.Table` struct.
There you can access the Builder, Database name, Columns and Associations.

```go
m := &Model{}
err := m.Initialize(m)
table := m.Table()
//...
```

# Transaction

A user defined transaction can be added to the model. This could be useful if you have further sql statements in a callback.
`Create`,`Update` and `Delete` are already using a transaction by default.

If you add your own transaction, make sure to also commit it!
```go
m := &Model{}
err := m.Initialize(m)
var tx *sql.Tx
//....
table := m.SetTx(tx) // set tx to model
m.Tx()  // get tx of model
//...
```


# Custom Type

Its possible to define a field oder relation as custom type, so they dont get requested by the database.

**Field**: If there is a field which does not exist in the database, its automatically marked as `custom`. Otherwise you can add the Tag "custom" to it.

**Relation**: If the tag `custom` is set or the function `Custom` returns true, a relation is marked as custom. The relation must have a Foreignkey. 
By default the foreign key is set to `{ParentModelName}ID`. You can modify this by adding the `fk` tag.

!> Why a foreignKey? because on the `All` callback the result set will get mapped to the matched row. 


# Callbacks

Callbacks are added automatically if a struct method with the specific name exists.
Callbacks can get disabled by `m.DisableCallback(true)`

The `*Cbk` has the following methods:

| Method | return value | Description |
|-----|-------|-------|
| Mode| `string` | Returns the actual mode (First, All ,...)|
| Parent| `*Parent`,`error` | Returns an error if no parent exists. Otherwise on a ptr you have access to `.Model()` and `.ResultSet()` |
| ResultSet| `interface{}`| The ResultSet is needed if the `All` method is called. _In the ResultSet you have the database results and can manipulate it.|
| Relation| `*Relation` | On the ptr you have access to `.Type()` and `.Field()`. Type returns the actual relation Type like `hasOne,hasMany,...`. The Field returns the relation StructField - so you can access the tags,type,.. |


| Trigger | Arguments | Description |
|-----|-------|-------------|
| Before, BeforeFirst, BeforeAll, BeforeCreate, BeforeUpdate, BeforeDelete | c *Cbk      | `Before` can be seen as global callback. It will be called on every type (First,All,..) if exists and if no specific callback is set. On *All, a result set will get added. This can be accessed by `.ResultSet()`|
| After, AfterFirst, AfterAll, AfterCreate, AfterUpdate, AfterDelete | c *Cbk       |  `After` can be seen as global callback. It will be called on every type (First,All,..) if exists and if no specific callback is set. On *All, a result set will get added. This can be accessed by `.ResultSet()`|

!> If you are using the All method the callback will have a resultSet with all the data. If the foreign keys are set correctly the result will get mapped to the
correct row automatically.

```go
func (f *File) OrmAfterAll(cbk *Callback) error {
    // getting all ids of the parent result set
	p, err := cbk.Parent()
	if err == nil {
		robots := *p.ResultSet().(*[]Robot)
		for _, r := range robots {
			fmt.Println("RID ->", r.ID)
		}
	}

    // adding Data to the File resulset
    // notice the RobotID field. this is the foreign key so it can get mapped correctly
	if cbk.ResultSet() != nil {
		result := cbk.ResultSet().(*[]File)
		*result = append(*result, File{ID: 1, RobotID: 4  Path: "aaaa"})
	}

	return nil
}
```

# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-model/issues


- [ ] BUGS: BelongsTo is not working correctly. specially if the related model was loaded before... Tag also has some design mistakes.
- [ ] BUGS: skipping Relations has a design flaw. example Customer -> Order -> Customer. Customer hasOne Order and Order belongsTo Customer. If its getting initialized like this. the next time you initialize Order, the belongsTo relation is missing because its cached with a skipped relations because of the first init of Customer.
- [ ] BUGS: we need a "Schema" in the config, otherwise we can not create a fqdn over different databases. Also Frist,All is using FQDN and update,create,delete not or only partly.
- [ ] Create,Update add validator to check if pk is filled or is an autoincrement value.
- [ ] First(), All() what to show if softDelete is existing? 
- [ ] validation tag and db table validators are not added at the moment
- [ ] Relation Orderby Tag?
- [ ] RelationCondition ??? for what?
- [ ] Exclude? - option from strategy? m.Strategy().Option('Exclude',...)