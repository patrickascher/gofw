<h1>sqlquery</h1>

sqlquery is a simple programmatically sql query builder.
The idea was to create a unique Builder which can be used with any database driver in go.

<h1>Features:</h1>

 - Unique Placeholder for all database drivers. **user-friendly**
 - Batching function for large Inserts. **performance**
 - Whitelist  **security**
 - Quote Identifiers **security**
 - SQL queries and durations log **debugging**
 - Can be used  with MySQL, PostgreSQL and other compatible database if needed. **user-friendly**

# Builder


## Install

```go
go get github.com/fullhouse-productions/go-sqlquery
```

## Configuration

To make the builder work with different database drivers, you have to config the following parameters.

| Param      | Description|
|-------------|------------------------------------------------------------|
| Adapter     | name of the database driver which is normally added to `sql.Open`                                                           |
| Host        | hostname of the database                                                           |
| Port        | port number of the database                                                           |
| Username    | username for the database                                                           |
| Password    | password for the database                                                          |
| Database    | database name                                                           |
| QuoteChar       | quote character of the database. psql `"` mysql `                                                         |
| Debug       | active debug log `default:false`                                                           |
| Placeholder | see [Placeholder](sqlquery?id=placeholder)  

```go
c := Config{
  Adapter: "postgres",
  Host: "localhost",
  Port: 5432,
  Username: "postgres",
  Password: "",
  Database: "tests",
  Debug: false,
  QuoteChar: "`",
  Placeholder: Placeholder {
  	Char: "$",
    Numeric: true
  }
}
```

## Creating the Builder

There are two methods to create a new Builder.

```go
// Creates a new *sql.DB in the background
b := NewBuilderFromConfig(conf)
```

```go
// Using a global defined *sql.DB
b := NewBuilderFromAdapter(db,conf)
```

!> You should avoid opening an closing a db connection as performance reasons. Use the connection pool instead.

## Raw
all columns are getting escaped by default. You can use raw sqls with this method.

```go

sqlquery.Raw("Concat(name, id)")
```

## Select

The method `Select` requires one parameter which should be the `FROM` table.
It returns the `sqlSelect` struct which offers the following methods.

| Method      | Description|
|-------------|------------------------------------------------------------|
| Columns(column ...string)     | multiple strings can get added as columns.        |
| String()    |String returns the statement and arguments.An error will return if the arguments and placeholders mismatch.        |
| First()    |First will return only one row `*sql.Row`. Its a wrapper for DB.QueryRow. So errors are only available in the `Scan` method.         |
| FirstTx(*sql.Tx)    | Same as `First` with a transaction.          |
| All()    |All will return all rows `*sql.Rows`. An error will return if the arguments and placeholders mismatch.            |
| AllTx(*sql.Tx)    | Same as `All` with a transaction.             |
| | |
| Condition(*Condition)        | see [Condition](sqlquery)                                                         |
| Join        | see [Condition.Join](sqlquery?id=on)                                                         |
| Where        | see [Condition.Where](sqlquery?id=where)                                                            |
| Group    | see [Condition.Group](sqlquery?id=group)   |
| Having    | see [Condition.Having](sqlquery?id=where)    |        
| Order    |see [Condition.Order](sqlquery?id=order)          |   
| Limit    | see [Condition.Limit](sqlquery?id=limit-amp-offset)            |
| Offset    |see [Condition.Offset](sqlquery?id=limit-amp-offset)           |
  
If you add a complete Condition the `ON` will get reset, because its not supported in a select.
Also all further method calls like `Where`,`Having`,... are getting added to the given condition.   

```go
// join condition
joinC := Condition{}
joinC.On("robots.id = parts.robot_id AND robots.id != ?",2)

// select only one row
b.Select("robots")
.Columns("name","brand","version")
.Join(LEFT, "parts", joinC)
.Where("id = ?",1)
.Order("name","-version")
.First()

// first with transaction
tx,err := b.NewTx()
// ... error handling
b.Select("robots")
.Columns("name","brand","version")
.Join(LEFT, "parts", joinC)
.Where("id = ?",1)
.Order("name","-version")
.FirstTx(tx)
// ... other statements 
err = b.CommitTx(tx)

// select all data from the givne select and condition
b.Select("robots")
.Columns("name","brand","version")
.Join(LEFT, "parts", joinC)
.Where("id = ?",1)
.Order("name","-version")
.Limit(5)
.Offset(10)
.All()

// select all  with transaction
tx,err := b.NewTx()
// ... error handling
b.Select("robots")
.Columns("name","brand","version")
.Join(LEFT, "parts", joinC)
.Where("id = ?",1)
.Order("name","-version")
.Limit(5)
.Offset(10)
.AllTx(tx)
// ... other statements 
err = b.CommitTx(tx)


// unescaped columns
b.Select("robots")
.Columns("name","!IF(version>1,'new','old') as version")
.String()
```

!> Be aware that `.All()` returns a sql.Rows struct which has to get closed `.Close()`!

?> Columns, the FROM table and the JOIN table are getting quoted!

## Insert

The method `Insert` requires one parameter which should be the `INTO` table.
It returns the `sqlInsert` struct which offers the following methods.

| Method      | Description|
|-------------|------------------------------------------------------------|
| Columns(column ...string)     | multiple strings can get added as columns.        |
| LastInsertedID(column,ptr)     | get last inserted id over different drivers        |
| Batch()    | Default `50`         |
| Values()    | Key Value pairs           |
| String()    | String returns the statement and arguments An error will return if the arguments and placeholders mismatch.        |
| Exec()    | Executes the statement. An error will return if the arguments and placeholders mismatch or no value was set.       |
| ExecTx(*sql.Tx)    | Same as `Exec` but with a transaction.|

Insert take the Columns from the given `Value` which is a key/value map.
If you add the Columns manually you can create a Whitelist for the Values and only the entered columns are getting inserted. As well the columns are rendered in the given order.
If you do not add a column definition `.Columns()` then the column order in the query is not guaranteed because its a map in the background.
```go
// insert data
values := []map[string]interface{}
values = values.append(values,map[string]interface{}{"name":"Wall-e","name":"Cozmo"})

// normal insert
b.Insert("robots")
.Columns("name")
.Values(values)
.Exec()

// batched Values - make sense with bigger values amounts
b.Insert("robots")
.Batch(50) // As soon as there are more than 50 values added, a batch will get preformed
.Columns("name")
.Values(values)
.Exec()

// normal insert with a transaction
tx,err := b.NewTx()
// ... error handling
b.Insert("robots")
.Columns("name")
.Values(values)
.ExecTx(tx)
// ... other statements 
err = b.CommitTx(tx)
```

!>If `Exec` is called and batch is triggered, it will automatically use a transaction for all the inserts!

!>IF you are using `LastInsertedID` the return value of `sql.Results` will be null

?> Columns and the INTO table name are getting quoted!


## Update

The method `Update` requires one parameter which should be the `UPDATE` table.
It returns the `sqlUpdate` struct which offers the following methods.

| Method      | Description|
|-------------|------------------------------------------------------------|
| Columns(column ...string)     | multiple strings can get added as columns.        |
| Set()    | Set the column/value pair          |
| String()    | String returns the statement and arguments. An error will return if the arguments and placeholders mismatch or no value was set.      |
| Exec()    | Executes the statement. An error will return if the arguments and placeholders mismatch, no value was set or the sql query returns one|
| ExecTx(*sql.Tx)    | Same as `Exec` but with a transaction.|
|||
| Condition(*Condition)        | see [Condition](sqlquery)                                                         |
| Where()    |see [Conditions.Where](sqlquery?id=where)  |


Update take the Columns from the given `Set` which is a key/value map.
If you add the Columns manually you can create a Whitelist and order for the Set.
If you do not add a column definition `.Columns()` then the column order in the query is not guaranteed because its a map in the background.

```go
// update value
values := map[string]interface{}
values["name"] = "Wall-e"

// update the name of robot #1
b.Update("robots")
.Columns("name")
.Set(values)
.Where("id = ?",1)
.Exec()

// update with a transaction
tx,err := b.NewTx()
// ... error handling
b.Update("robots")
.Columns("name")
.Set(values)
.Where("id = ?",1)
.ExecTx(tx)
// ... other statements 
err = b.CommitTx(tx)
```

?> Columns and the UPDATE table name are getting quoted!


## Delete

The method `Delete` requires one parameter which should be the `DELETE` table.
It returns the `sqlDelete` struct which offers the following methods.

| Method      | Description|
|-------------|------------------------------------------------------------|
| String()    | String returns the statement and arguments. An error will return if the arguments and placeholders mismatch.|
| Exec()    | Executes the statement. An error will return if the arguments and placeholders mismatch or the sql.Exec creates with an error.|
| ExecTx(*sql.Tx)    | Same as `Exec` but with a transaction.|
|||
| Condition(*Condition)        | see [Condition](sqlquery)                                                         |
| Where()    |see [Conditions.Where](sqlquery?id=where) |


```go
// delete robot #1
b.Delete("robots")
.Where("id = ?",1)
.Exec()

// delete robot #1 with transaction
tx,err := b.NewTx()
// ... error handling
b.Delete("robots")
.Where("id = ?",1)
.ExecTx(tx)
// ... other statements 
err = b.CommitTx(tx)
```

?> The DELETE table name are getting quoted!


## Information

The method `Information` requires one parameter which should be the table name.
On an `Information` struct, `Describe` and `ForeignKeys` is available.

To have a unique result over multiple database drivers, an `Driver` interface is defined. 
At the moment only mysql and postgres are implemented.
```go
type Driver interface {
	Describe(db string, table string, builder *Builder,cols []string) *Select
	ForeignKeys(db string, table string, builder *Builder) *Select
	ConvertColumnType(t string, column *Column) Type
}
```

### Describe

Describe is returning all columns of the table. It returns a `[]Column` and an `error`.
Error will return if something went wrong with the database select or the tables has zero columns.

If you only need to describe some columns, you can enter the column name as argument.
```go
type Column struct {
	Name          string
	Position      int
	NullAble      bool
	PrimaryKey    bool
	Type          Type
	DefaultValue  NullString
	Length        NullInt64
	Autoincrement bool
}
```

The `Type` struct is used to globalize the sql type for different databases.
These types are defined:

| Type | Options | Mysql | Postgres |
|------|---------|-------|----------|
| Integer     |  `Min`, `Max`, `Raw()`, `Kind()`  | bigint, int, mediumint, smallint, tinyint      | bigint, int8, integer, smallint         |
| Text        |   `Size`, `Raw()`, `Kind()`        | varchar, char      | character, character varying         |
| TextArea    |  `Size`, `Raw()` , `Kind()`        | tinytext, text, mediumtext, longtext      | text         |
| Float       | `Raw()`, `Kind()`                 | decimal, float, double      | real, double precision, numeric         |
| Time        |  `Raw()`, `Kind()`                | time      | time without time zone, time with time zone         |
| Date        |   `Raw()`, `Kind()`                | date     |  date        |
| DateTime    |  `Raw()`, `Kind()`                 | datetime, timestamp      |  timestamp with time zone, timestamp without time zone        |
| Enum        | `Values`, `Raw()`, `Kind()`       | -      |  -        |
| Set         | `Values`, `Raw()`, `Kind()`       | -      |  -        |


```go
// describe the table robots
col,err := b.Information("robots").Describe()

// describe column id,name
col,err := b.Information("robots").Describe("id","name")

// describe the database xy and table z 
col,err := b.Information("xy.z").Describe()
```

?> If you are implementing another driver, keep in mind that the describe select must be in the same order as the Column struct `Name`, `Position`, `NullAble`, `PrimaryKey`, `Type`, `DefaultValue`, `Length`, `Autoincrement`
                                                                                                                                                     	
                                                                                                                                                     	
### ForeignKeys                                                                                                                                                 
                                                                                                                                                     
ForeignKeys is returning all fk´s of the table. It returns a `[]ForeignKey` and an  `error`.
 
 A ForeignKey struct includes the Name, Primary Table, Primary Table Column, Secondary Table and Secondary Table Column.
 ```go
type ForeignKey struct {
	Name      string
	Primary   *Relation
	Secondary *Relation
}

// Relation defines the table and column of a relation
type Relation struct {
	Table  string
	Column string
}
 ```
 
 ```go
 // foreign keys of the table robots
 fkeys,err := b.Information("robots").ForeignKeys()
 
 // foreign keys the database xy and table z 
 fkeys,err := b.Information("xy.z").ForeignKeys()
 ```                                                                                                                                               

?> If you are implementing another driver, keep in mind that the foreignKey select must be in the same that order `name`, `primary table`, `primary table column`, `secondary table`, `secondary table column`


## Transaction

Transactions can be used like this.

```go
var err error

tx,err = b.NewTx()
//...
err = b.Insert("robots"). ... .ExecTx(tx)
//...
err = b.Insert("parts"). ... .ExecTx(tx)
//...
err = b.CommitTx(tx)
//...

```

## QuoteIdentifier

This function is internally used to quote the identifiers. 
Some Conditions can be complex and its not possible at the moment to automatically quote the identifiers of them.
This function should help you to quote the Identifier of your Conditions.


```go
b.QuoteIdentifier("robots")
// `robots` - depending on your driver settings
b.QuoteIdentifier("robots.name")
// `robots`.`name` - depending on your driver settings

b.QuoteIdentifier("robots.name a")
// `robots`.`name` `a` - depending on your driver settings

b.QuoteIdentifier("robots.name AS a")
// `robots`.`name` `a` - depending on your driver settings
```

!> All identifiers with the prefix `!` are not getting escaped.

## Other Database

If you want to connect to a different Database as defined in the Builder, just prefix your table name with the database name.

```go
b.Select("robots") // requesting the robots table of default database of dns
b.Select("db2.robots") // requesting the robots table from the database db2
```


# Config

Sometimes its useful to get some configurations of the builder.
This function returns the Database interface.


 **Methods available**
 
| Param      | Description|
|-------------|------------------------------------------------------------|
| Driver()     | string                                                       |
| DSN()        | string                                                   |
| Placeholder()        | *Placeholder                                                      |
| QuoteCharacter()        | string                                                       |
| Debugger()        | Debugger                                      |
| DbName()        | string                                                       |


```go
b.Config().DbName()
```

# Placeholder

The `sqlquery` package is using a unique placeholder `?`.
To translate it to your specific database driver, you have to configure it once.

| Param      | Description|
|-------------|------------------------------------------------------------|
| Char     | The character you driver is using as placeholder `?`,`$`,...                                                         |
| Numeric        | If Numeric is set to true its placing a interator after your placeholder. `$1`,`$2`,...                                                       |


```go

p := Placeholder{
	Char: "$",
	Numeric: true
}
```

# Conditions

!> Condition identifiers are not getting quoted automatically at the moment.
You can use the builder function `QuoteIdentifier` for it.

The following public constants are available.
```go
WHERE = iota + 1
HAVING
LIMIT
ORDER
OFFSET
GROUP
ON

```

## Creating a Condition

```go
c := Condition{}
```


## Where
Multiple `where` conditions can be used. They get connected by an `AND`.
This means, if you need an OR Condition, be aware to set the right brackets or write the whole Condition in one Where call.

```go
c.Where("id = ?",1).Where("deleted IS NULL")
// WHERE id = 1 AND deleted IS NULL


// OR
c.Where("(id = ? OR id = ?)",1,2).Where("deleted IS NULL")
// WHERE (id = 1 OR id = 2) AND deleted IS NULL
```

## Group

Simple `GROUP BY` condition. 
Multiple columns can be set. 

```go
c.Group("id","name")
// GROUP BY id, name
```

## Having

same logic as see Where


## Order

Simple `GROUP BY` condition. 
Multiple columns can be set. 
If you prefix the column name with a `-`, `DESC` will be added.

```go
c.Group("id","-name")
// ORDER BY id ASC, name DESC
```

## Limit & Offset
```go
c.Limit(5).Offset(10)
// LIMIT 5 OFFSET 10
```


## On

On is available on JOIN statements

```go
// without arguments
c.On("robots.id = parts.robot_id")
// with arguments
c.On("robots.id = parts.robot_id AND robot.id != ?",1)
```


## Reset

Reset the defined condition by its constant.
Available for: `WHERE`, `HAVING`, `ON`, `LIMIT`, `OFFSET`, `GROUP`, `ORDER`

```go
c.Reset(WHERE)
```

## Config
Returns the rendered string of the condition.
Available on: `WHERE`, `HAVING`, `ON`, `LIMIT`, `OFFSET`, `GROUP`, `ORDER`

```go
c.Config(WHERE)
```


# Issues & Ideas

To report Issues or to improve this package, please use the github issue board or send a pull request.

https://github.com/fullhouse-productions/go-sql/issues
