package orm

// The orm package converts a normal struct into a full orm by embedding orm.Model.
// A primary is mandatory for the model and relations.
// If CreatedAt, UpdatedAt or DeletedAt(implement) is defined, the fields will be filled automatically on Create(), Update() or Delete().
//
//
//
//
// Grid:
// Orm models can be transformed into a grid source by simply wrap the model around orm.Grid(&yourModel).
//
// The grid has two policies. On Blacklist, all orm.Fields are added and you have to remove them manually when not needed. This means
// all relations and data is fetched. IF this is not necessary use the Whitelist policy. This means no orm.Field is added by default.
// You have to choose which field is required and only the required data is fetched by the database. This is the default policy to avoid
// unnecessary database calls. A minimum of one field has to be defined.
//
// The grid transforms the orm into a json object. To avoid overhead, use the json omitempty tag on your fields.
//
// If a Field is set to ReadOnly, the write permission will be set to false.
//
// On belongsTo and manyToMany relations only the foreign key will be updated not the relation itself.
// There is a callback "FeSelect" for select values. This is used for belongsTo and M2m relations to fetch the data.
//
// On belongsTo there is a special case. On Grid mode VUpdate the foreign key will be used as Select field and the relation itself is disabled.
// TODO: change this logic, and implement it right in the frontend.
// Keep that in mind when you are using a whitelist policy.
//
// All primary-, foreign-, association-, and polymorphic keys are set to remove by default.
//
//
//
//
//
// Defaults:
// Some default values can be set by simply create a model function to overwrite the default behaviour.
// DefaultLogger() *logger.Logger: Default the orm.GlobalLogger is used.
// DefaultCache() (manager cache.Interface, ttl time.Duration, error error): Default the orm.GlobalCache is used with the cache time infinity.
// DefaultBuilder() sqlquery.Builder: Default the orm.GlobalBuilder is used.
// DefaultTableName() string: Default the plural structname in snake case.
// DefaultDatabaseName() string: Default the database name of builder configuration.
// DefaultSchemaName() string: Default the schema name of builder configuration.
// DefaultStrategy() string: Default "eager"
//
//
//
//
// Tags:
// column: this will change the database column name. By default its the name of the struct field in snake case.
// permission: can be set to "rw" the the r stands for read permission and w for write permission. If one is missions, it will get set to false.
// select: can be used to create a custom select for the field. `orm:"select:Concat(id,name)"`. The given string will be used as raw sql, a AS Fieldname command is added automatically.
// primary: can be used to set a primary key. By default ID field will be set. TODO: make this working for all relations, at the moment there is ID hardcoded in some places.
// custom: can be used to define a field as custom field. Custom fields will be ignored by the database.
// relation, fk, afk, join_table, join_fk, join_afk, polymorphic, polymorphic_value: please see the Relation section.
//
//
//
//
// Null Values:
// The orm package implements its own Null values for:
// NullString, NullFloat, NullInt, NullBool, NullTime
//
//
//
//
//
// Relation:
// Relations can be added as struct. The struct must embed the orm.Model field.
// Relations are by default added as:
// All Structs are hasOne relations and all slices are hasMany relations.
//
// To change the type of relation, simply add a tag.
// * belongsTo: orm:"relation:belongsTo"
// * hasOne: orm:"relation:hasOne"
// * hasMany: orm:"relation:hasMany"
// * Many2Many: orm:"relation:m2m"
//	For this relation  the join table can be set.
//	By default its the {struct name singular}_{the struct name plural}. The two join fields are {structname}_id. If its a self referencing m2m relation the second field is child_id.
//	The tag to define your own join table would be: orm:"relation:m2m;join_table:test;join_fk:a;join_afk:b"
//
// Polymorphic is possible on hasOne and hasMany relations.
// orm:"relation:hasOne;polymorphic:Car;"
// This would require a CarID and CarType on the relation orm model. The {Polymorphic}Type is default the struct name of the root orm model.
// To change the Type value, the Tag can be used:
// orm:"relation:hasOne;polymorphic:Car;polymorphic_value:a;"
//
// The foreign key and association key can be changed with the tags.
// The point of view is always from the root model. fk always on the root model, afk field on the related model.
// hasOne: orm:"fk:a;afk:b"
//
//
//
//
// Scope:
// Can be used to access internal fields or relations. This can be useful on loading strategies or callbacks.
// A lot of helpers exists. Please check out the scope.go file for more details.
//
//
//
//
// Strategies:
// The whole orm model is using a loading strategy. This means a lazy-loading or eager loading can be implemented.
// By default an eager loading strategy exists. To implement your own strategy simply implement the orm.Strategy interface.
//
//
//
//
// Validation:
// In the background go-playground/validator is used.
// This means you can add the validation simply by the tag `validation:"required"`.
// Please see all available validation here https://godoc.org/gopkg.in/go-playground/validator.v10.
// The database fields will be checked against there column type and some default validations are set.
//
// required: if the field does not allow null in the database and its not an autoincrement column:
// omitempty: if null is allowed, or its a belongsTo foreign key field (value is set later).
// numeric(min,max): Integer. The min and max value is set depending on there db column type.
// numeric: Float
// max: Text, Textarea types. The max value is depending on there db column type.
// oneof: Select. The values are added to the oneof validator.
// TODO Time,Date,DateTime
//
//
//
//
// White/Blacklist:
// Fields can be Black or Whitelisted on the orm model.
// This is useful if you dont want to update all fields/relations or disable some fields for a user.
//
// All mandatory fields which are needed for relations or the orm model itself are added automatically.
// If one of the fields are getting blacklisted, it will add itself again.
//
// Fields of a relation can be referenced by a dot notation. example: m.WBList(BLACKLIST, "Owner.Name") would blacklist the Field "Name" of the Relation "Owner".
//
// A WB List can be added to the struct like this.
// White and Blacklist can not used together. A new WBList() call will overwrite the one before.
// m.WBList(Whitelist,"Field1","Field2")
// m.WBList(Blacklist,"Field1","Field2")
//
// In the background the field Permission is set to false/true. If you want to cache your settings for the whole model, just set the WBList() before the Init() call.
// Be aware, in that case the Permission is saved for every call on this model.
