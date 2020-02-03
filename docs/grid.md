<h1>go-grid</h1>

Grid is a module which transforms a go_orm struct into a CRUD backend. 

# Install

```go
go get github.com/fullhouse-productions/go-grid
```


# Usage

```go
// orm
u := User{}

// grid
grid := New(c) //c is the contoller
grid.Source(u)

// special options for field ID
id,err := grid.Field("id")
if err != nil{
	//...
}
id.SetTitle("User Identifier")

grid.Render()
//....
```

# New
Is used to create a new grid instance. As argument the controller has to get set because the request,response and router data (cache) is needed/getting set.


```go
grid := New(&controller)
//...
```

## Source
As source a struct is added which must implement the `go_orm.Interface`.
As second parameter, a condition can be added. If its not needed set as param `nil`.

The `controller` **must have** a cache, otherwise an error will return.

**Workflow:**
- The cache is getting added to the model.
- Then the model is getting initialized.
- All of the model fields are getting added to the grid (see createFields()).
    - Autoincrement fields or reference fiels are getting set to `hide=true` by default for all views.
    - if the field has no orm `Permission.Read`, the field gets skipped. (for example: CreatedAt, UpdatedAt, DeletedAt)
- Add an identifier that a source was set.

```go
u:=User{}

c:=&sqlquery.Condition{}
c.Where("...")
err := g.Source(u,c)
if err != nil{
	return err
}
```

!> ManyToMany self referencing relations, are only going to a depth-1 at the moment. 

## Action

To configure the grid table view and secure some options, you can configure the Grid object.
At the moment only these settings are available.

```go
g.Config().Action.DisplayLeft = true
g.Config().Action.New.Disable = true
g.Config().Action.Edit.Disable = true
g.Config().Act ion.Delete.Disable = true
```

!> **TODO** create Setters which return the action struct.

## Mode

Mode will return the correct grid mode with the request information of the controller context request.
The return value will be an int. Please use the defined constants for it.
`VIEW_GRID`,`VIEW_NEW`,`VIEW_EDIT`,`VIEW_DETAILS`,`CREATE`,`UPDATE`,`DELETE`

```go
mode := g.Mode()
```

| HTTP Method   | Grid Mode      | Params   | 
|---------------|----------------|----------------|
| GET           | ViewGrid      |   | 
| GET (mode)    | ViewCreate       | ?mode=new  | 
| GET (pk, mode)| ViewEdit      | ?mode=edit&id=1  | 
| GET (pk, mode)| ViewDetails   | ?mode=details&id=1  | 
| GET (pk?)| SelectValues   | ?select=xy  | 
| POST          | CREATE         |   | 
| PUT           | UPDATE         | ?id=1  | 
| DELETE        | DELETE         | ?id=1  | 

## Disable 
With the method `Disable` some default logics can be disabled.
The following constants should be used for it:

- `PAGINATION` is removing the pagination
- `HEADER` is removing the header information of the orm
- `CALLBACK` is removing the `BeforeView` and `AfterView` callbacks.

An error will return if the option is unknown.


```go
err := g.Disable(HEADER)
if err != nil{
	return err
}
//...
```

## Field
Field is returning a grid field by its struct name.
Error will return when no field was found.

```go
id,err := g.Field("ID")
if err != nil{
	return err
}
//...
```

## Relation
Relation is returning a grid relation by its struct name.
Error will return when no relation was found.

```go
account,err := g.Relation("Account")
if err != nil{
	return err
}
accID,err := account.Field("ID")
//...
```

## SetFields{ReadOnly,Remove,Hidden}
This is a small helper to set options for a set of fields.

```go
err := g.SetFieldsRemove(true,"Name","Age","Parts.Number")
if err != nil{
	return err
}

// or with the valuer

err := g.SetFieldsRemove(Value(true).Edit(false),"Name","Age","Parts.Number")
if err != nil{
	return err
}
```

## Render
Render is handling the different grid modes.

The method is not returning any error anymore. 
The errors are set as HTTP ERROR with the code 500.

ViewGrid - grid view with all the data
ViewCreate  - create view
ViewUpdate,ViewDetails - edit or details view
CREATE - create data
UPDATE - update data
DELETE - delete data

```go
g.Render()
//...
```

## Internals described

The following is just internal information which should not be visible for the end-user later on.

### getRelationName
Is returning the struct string name. 
Its getting checked against a slice, slice ptr, ptr and struct and always returns the underlaying struct name as string.


### createFields
createFields is adding the model fields and relations to grid
needed for the later configuration of the grid

**normal fields:**
- Title will be the struct field name
- If a field is a reference field and autoincrement fields like (id or fk) it will be set to hidden.
- Set the position by the db position.
- Set the type of the field.(Integer,Text,TextArea,Float,Time,Date,DateTime,Enum,Set) (not raw database column type - see sqlquery.information)
- Set reference to the go_orm.Column.
- class should be callable on itself.

**relations**
- Title will be the struct field name
- If given relation name does not exist, its getting skipped.
- Position starts at 500 and is getting incremented. (this is fixed in a later state). The relations are sorted alphabetical.
- Getting the related model from the cache
- Adding the relation to grid fields slice.
- calling createFields on itself to recursive loop over all relations and add there fields.

**Frontend**
belongsTo should be a Select
hasOne should be added as normal fields. maybe with a hr tag + title.
hasMany should be an extra table.
m2m should be a Multiselect. Maybe with the option to add additional entries on a gui.


### headInfo
Head is setting the head information of the controller response if its not disabled by the user. 
This includes, title, position, type, filter, sort, remove, hide ,....

#### sortHeaderInfo
is sorting the model with the position attribute. 
By default the position is the same as in the database defined, after that all relations are added in alphabetical order.
The relations do not have a "fixed order" because its a map. You have to define a fixed position for them.

#### headerFieldsLoop
headerFieldsLoop is going recursive over all fields and relations to fetch the info.

### marshalModel
marshal Model is checking the request Body, validates it if its a json and then tries to map it to the model.
Unknown fields are not allowed and an error be thrown.
**not yet** It is also checking if a model is empty - if so, a error will be thrown.

### conditionAll
Is returning a new condition for the `GridView`.
Its checking all request params and checking if a sort or filter is existing.
**Params**
- `sort` is used for sorting. ex: ?sort=id,-name (if there is a - in front it will get translated to a DESC)
The field name has to be the database column name.
- `filter_` Params with the prefix will be added to the filter condition. Everything after `filter_` will be recognice as the column name.
ex: ?filter_id=1,2,3 creates a WHERE id IN (1,2,3)

If the column name does not exist in the struct, a error will return.

#### addFilterCondition, addSortCondition
Are just helper methods for the sorting and filtering.

#### isFilterAllowed, isSortAllowed
returns if a filter/sort is allowed with the given field.

### getFieldByDbName 
returns the field interface of the grid if the db column exists. Otherwise an error will return.

### checkPrimaryParams
Creates a condition for the primary key(s) of the orm.model.
Checking if all primary keys are given by the request params, if not an error will return.
Used for readOne, update and delete.

### readAll, readOne, 
handling the crud. 
headerInfo is added if the user did not disable it.
readAll also has a pagination if the user did not disable it.

###create, update, delete
Is marshal the json request and tries to save the go_orm.model.
Error will return if something went wrong

# Value

The method is used to set a value to a given field for different render modes.

```go
v := Value("Reference number")
grid.SetTitle(v)

// or shorter 
grid.SetTitle("Reference number")
```

This will set the value for the title to `Reference number` to all modes (ViewGrid, ViewDetails, ViewCreate, ViewUpdate)

## Different value by view
You can change the value to different Views.
Let´s assume you want to have in the Grid view a short name and in all others the full title.

```go
v := Value("Reference number").Grid("Ref.nr")
grid.SetTitle(v)
```
The following methods are available `Grid`, `Details`, `Create`, `Edit`.

# Field/Relation


| Method   | type     | Grid Mode      | Field / Relation   | Description   | 
|---------------|---------------|----------------|----------------|----------------|
| SetTitle       | `string`   | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`      | both   | | 
| SetDescription | `string`    | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit  `     | both | | 
| SetPosition| `int` | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`      | both  | | 
| SetView  | `string`     | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`          | both  | Sets the header `view` - if exists. A vue component with that name will get rendered. Its available in edit&grid. The field value is added as v-model.| 
| SetViewValue  | `interface`     | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`          | both  | Sets the header `viewValue` - if exists. The propertie `viewValue` will get added to the component.| 
| SetHide| `bool` | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`  | both  | | 
| SetRemove| `bool`         | `ViewGrid`, `ViewDetails`, `ViewCreate`, `ViewEdit`         | both  | `SetRemove` is removing the field from the orm select. In the background all fields are getting blacklisted. | 
| SetSort  | `bool`        | `ViewGrid`        | both  | | 
| SetFilter  | `bool`     | `ViewGrid`         | both  | | 
| SetReadOnly  | `bool`     | `ViewGrid`, `ViewEdit`         | both  | The field will get set as readOnly automatically if the tag permission:r is set. In the Grid view the readOnly field are shown light grey and with a font-style italic. In the edit mode the most fields are getting disabled. | 
| SetCallback   |     | `ViewGrid`          | both | please see the callback section |
| SetSelect   | `string`    | `ViewGrid`, `ViewDetails`         | Field  | custom column sql can get set. see orm tag `select` | 
| Select   | `SelectI`    | -         | Relation  | **TODO** delete when field type is ready.  The `Select` method is available on `BelongsTo`,`ManyToMany` and `ManyToManySR` types. By default the select key = the second field of the struct and the value = the primary key. Data will get fetched by ajax automatically. If you want, you can set or get the following data. `Items()`, `SetItems()`, `ValueKey()`, `SetValueKey()`, `TextKey()`,`SetValueKey()`. If a custom item set is set, no ajax request will be made from the frontend anymore. | 
| Field   | `obj`    | -         | Relation  | | 
| Relation   | `obj`    | -         | Relation  | | 

?> All Methods can be called with the underlaying type or the Value struct to define specific values by mode.

```go
id,err := grid.Field("ID")
if err != nil{
	return err
}
//sets the Title for all views.
id.SetTitle("Identifier") 

// set different values for the modes.
v := Value("Identifier").Grid("ID") // Sets the title Identifier to all modes except the GridView.
id.SetTitle(v) 

id.SetCallback(Decorator,"{{ID}} - {{Name}}",false)

```

# Callbacks

Callbacks are added automatically if a struct method with the specific name exists.

Callbacks can get disabled by `g.Disable(CALLBACK)`

| Trigger | Arguments | Description |
|-----|-------|-------------|
| GridBeforeView      | Grid `*grid`     | Will get called on every mode `Edit,All,Create,Details` if no specific mode callback is defined. |
| GridAfterView      |Grid `*grid`      |  Will get called on every mode `Edit,All,Create,Details` if no specific mode callback is defined. |
| GridBeforeViewEdit       | Grid `*grid`       |        |
| GridAfterViewEdit       | Grid `*grid`       |           |
| GridBeforeViewDetail       |Grid `*grid`        |        |
| GridAfterViewDetail      | Grid `*grid`       |           |
| GridBeforeViewCreate       | Grid `*grid`       |        |
| GridAfterViewCreate      | Grid `*grid`       |           |
| GridBeforeViewAll       | Grid `*grid`       |        |
| GridAfterViewAll      | Grid `*grid`       |           |



# (Field) Callbacks

For callbacks, use the `SetCallback` method on a Field or Relation. 
The firsts parameter must be a `string` or `func`.
* `func` the following arguments are getting added `own defined argumensts...` and at the last position the row data as map `row data`
* `string` its getting checked if the function exists in the main struct. If so, `own defined argumensts...` getting added.

The functions must return a value. This value is getting set into the map result. This means the returned value must not match the struct field type.
But if they match, the struct field is getting set as well. This is needed for nested callbacks.

The callbacks are only available in the `GridView` mode.

`Default`: Every struct/field is getting checked if a method `GridCallback{Fieldname}` exists in the main struct.
For whole relation `GridCallback` is getting checked if exists in the relation struct.
If so, the callback will get added automatically. 

!> By default, a argument is getting added which is the actual row. Be aware of this, when you create your callback.
**TODO** Create a better solution for this.

```go 
// automatic callback is set
type Test struct{
    ID int
    Name string
    ...
}
func (t *Test) GridCallbackName() string{
    return t.Name + "some decoration" 
}

// manual added callback
n,err := grid.Field("Name")
if err != nil{
	return err
}
n.SetCallback(Decorator,"{{ID}} - {{Name}}",false)
func Decorator(format string, htmlescape bool, data interface{}) string {
	//...
	return result
}

// automatic callback with a nested callback
type Test struct{
    ID int
    T2 []Test2
    ...
}
type Test2 struct{
    ID int
    Name string
    ...
}
func (t2 *Test2) GridCallbackT2() string{
    return t2.Name + " decorated"
}
func (t *Test) GridCallbackT2() string{
    for _,field := t.T2{
        // field would have the callback value "T2.Name decoreated" - if the return value of the callback fits the field type!
        //....
    }
    //..
}

```

**Predefined callbacks:**


| fn   | return type     |  Arguments   |Description   |
|------|-----------------|----------------|----------------|
| go_grid.Decorator| `string`| `format string, htmlescape bool`|  The `Decorator` callback can be used to change the result. You can use any struct field of the row by wrapping it into `{{ FIELDNAME}}`. Its possible to call child fields by dot notation `{{Parts.ID}}`. But you can not mix different depths, for example `{{Parts.ID}} {{Name}}` will through an error. For more complicated things, please use a model function. The point of view (start-point) is always the main struct. |

```go
id.SetCallback(Decorator,"{{ID}} - {{Name}}",false)
// or with a relation
id.SetCallback(Decorator,"{{Parts.ID}} - {{Parts.Name}}",false)

```


# FieldType
 
By default each field is getting a field type. This is the underlaying type of the database column or relation type. (see: http://localhost:3000/#/skeleton_frontend?id=input-base)

If you need your own typ, you can create a struct function with the name `GridFieldType`. The return value must be a FieldInterface.
You then have to set a View and implement the vue frontend component. The `GridFieldType` gets one argument and this is a ptr to the actual grid instance.

```go
func (r *Robot) GridFieldType(g * go_grid.Grid) go_grid.FieldType{
	ft := go_grid.DefaultFieldType(g)
	ft.SetName("RobotType")
	ft.SetView("RobotType")
	return ft
}
```

**Predefined Options:**

 - noEscaping (in Grid Table the element will not get escaped used for example: <br/>)
 - view (special view for the vue)
 - select.textKey (For the select elements the displayed text field)
 - select.valueKey (For the select elements the value field)
 - select.items (For the select elements the displayed options)
 
# Pagination (Offset)
Simple pagination by offset. 

The pagination is added to the controller data by the key `pagination`.
`Prev` and `Next` will have a zero value if its the first or last page.

```json
{"Limit":2,"Prev":1,"Next":3,"CurrentPage":2,"Total":6,"TotalPages":3}
```


The following url params are available:
 - limit (rows per page)
 - page (page number)


# Issues & Ideas

- [ ] Create an interface for pagination that also an infinity pagination can be added in the future without any code changes
- [ ] Create option to set a condition to delete and update. so that we are also able to edit the pkey.