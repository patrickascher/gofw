package orm

import (
	"database/sql"
	driverI "database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/patrickascher/gofw/sqlquery"
	strings2 "github.com/patrickascher/gofw/strings"
)

const (
	model = "orm.Model" // used to identify the embedded orm field.

	tagName      = "orm"
	tagSkip      = "-"
	tagSeparator = ";"
	tagKeyValue  = ":"

	tagCustom     = "custom"
	tagColumn     = "column"
	tagPermission = "permission"
	tagSelect     = "select"
	tagPrimary    = "primary"
)

var (
	errDBColumn           = "orm: column %s does not exist table %s"
	errDbSync             = errors.New("orm: database is not in sync with the struct")
	errPrimaryKey         = "orm: no primary key or ID FIELD is set in  %s"
	errSoftDelete         = "orm: the soft delete field %s does not exists in the orm %s"
	errSoftDeleteDatabase = "orm: the soft delete field %s does not exists in the db table %s"
	errNullField          = "orm: db: \"%s\" column: \"%s\" is null able but field \"%s\" does not implement the sql.Scanner interface"
)

// Field is holding some struct field information.
type Field struct {
	Name string

	SqlSelect   string
	Permission  Permission
	Information sqlquery.Column

	Validator *validator
	Custom    bool // defines if the Field is not in the database table.
}

// Permission of the field.
// If read or write is disabled, the database strategy will ignore this field.
type Permission struct {
	Read  bool
	Write bool
}

// createFields will map all exported struct fields to model fields.
// The fields are configured by tag, checked if the primary exists and the
// fields are existing in the database.
func (m *Model) createFields() error {
	for _, field := range m.structFields(m.caller, false) {

		// create field and db column
		f := Field{}
		f.Name = field.Name
		f.Information.Name = strings2.CamelToSnake(field.Name)
		f.Permission = Permission{Read: true, Write: true}

		// parse tag and config the column
		configFieldByTag(&f, field.Tag.Get(tagName))

		// validator
		f.Validator = &validator{Config: field.Tag.Get(TagValidate)}

		// add to model fields
		m.fields = append(m.fields, f)
	}

	// check if soft deleting field exists
	err := m.checkSoftDeleteField()
	if err != nil {
		return err
	}

	// check if at least one primary key is set.
	// this is checked before describe fields, because it checks if the Field ID exists or the user used the primary tag.
	// after that its compared to the database.
	// Permission.internal is set to true
	err = m.checkPrimaryKey()
	if err != nil {
		return err
	}

	// check if the fields exist in the database table.
	return m.describeFields()
}

func (m *Model) checkSoftDeleteField() error {
	f, _, _ := m.caller.SoftDelete()
	exists := false
	for k, _ := range m.fields {
		if m.fields[k].Name == f {
			m.softDeleteField = &m.fields[k]
			exists = true
		}
	}
	if !exists {
		return fmt.Errorf(errSoftDelete, f, m.modelName(true))
	}
	return nil
}

// checkPrimaryKey is testing if a user primary key was set.
// If not, the field "ID" will be used as primary key.
// An error return if none primary was set and no field with the name ID exists.
func (m Model) checkPrimaryKey() error {
	var IDField *Field
	pkCounter := 0

	for k, field := range m.fields {
		if field.Name == "ID" {
			IDField = &m.fields[k]
		}
		if field.Information.PrimaryKey == true {
			pkCounter++
		}
	}

	if pkCounter == 0 {
		if IDField == nil {
			return fmt.Errorf(errPrimaryKey, m.modelName(true))
		}
		IDField.Information.PrimaryKey = true
	}

	return nil
}

// describeFields will check if the struct fields exist in the database table.
// if pre defined time field created_at, updated_at and deleted_at dont exist, the will be removed from the model fields without throwing an error.
func (m *Model) describeFields() error {

	scope := Scope{model: m}
	// describe table columns
	var cols []string
	for _, field := range m.fields {
		if field.Custom {
			continue
		}
		cols = append(cols, field.Information.Name)
	}

	b := scope.Builder()
	describeCols, err := b.Information(m.caller.DefaultDatabaseName() + "." + m.caller.DefaultTableName()).Describe(cols...)

	// error will throw if table name does not exist.
	if err != nil {
		return err
	}
	//	m.DefaultLogger().Trace("Describe", err, describeCols, cols, m.caller.DefaultDatabaseName()+"."+m.caller.DefaultTableName())

	// adding database column structure to the column
	scannerI := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	valuerI := reflect.TypeOf((*driverI.Valuer)(nil)).Elem()

Columns:
	for i := 0; i < len(m.fields); i++ {

		if m.fields[i].Custom {
			continue
		}

		for n, describeCol := range describeCols {
			if describeCol.Name == m.fields[i].Information.Name {

				// Debug information that the database table does not fit the struct configuration.
				// At the moment only the primary key is checked.
				// TODO: In the future there should be a struct/db migration function
				if m.fields[i].Information.PrimaryKey != describeCol.PrimaryKey {
					return errDbSync
				}

				// if db column is nullable, check if a sql.scanner and driver.valuer is implemented.
				// TODO: In the future there should be a struct/db migration function
				timePtr := scope.CallerField(m.fields[i].Name)
				if timePtr.Type().Kind() == reflect.Struct {
					timePtr = timePtr.Addr()
				}
				if describeCol.NullAble == true &&
					(!timePtr.Type().Implements(scannerI) ||
						!timePtr.Type().Implements(valuerI)) {
					return fmt.Errorf(errNullField, scope.TableName(), describeCol.Name, m.fields[i].Name)
				}

				m.fields[i].Information = describeCol

				//decrease columns
				describeCols = append(describeCols[:n], describeCols[n+1:]...)
				continue Columns
			}
		}

		// if the predefined time fields does not exist in the database, delete it of the fields list.
		if m.fields[i].Information.Name == "created_at" || m.fields[i].Information.Name == "updated_at" || m.fields[i].Information.Name == "deleted_at" {
			// deleting the softDelete Field if the default DeletedAt field does not exist in the database
			// no error is shown to the user because this field is set as default with every model. Only
			// user defined fields will throw an error.
			if m.fields[i].Information.Name == "deleted_at" && m.softDeleteField != nil && m.softDeleteField.Information.Name == "deleted_at" {
				m.softDeleteField = nil
			}
			m.fields = append(m.fields[:i], m.fields[i+1:]...)
			i--
			continue
		}

		// return error if the user defined soft delete field does not exist in the database table.
		if m.softDeleteField != nil && m.fields[i].Information.Name == m.softDeleteField.Information.Name {
			return fmt.Errorf(errSoftDeleteDatabase, m.fields[i].Name, m.caller.DefaultTableName())
		}

		return fmt.Errorf(errDBColumn, m.fields[i].Name, m.caller.DefaultTableName()+"."+m.fields[i].Information.Name)
	}

	return nil
}

// structFields will return all struct fields of the given interface.
// If the relations bool is false, only the native types will return otherwise only struct types will be returned.
// The orm.Model is excluded of this logic but the default Fields CreatedAt, UpdatedAt and DeletedAt will be added at the end.
func (m Model) structFields(caller interface{}, relations bool) []reflect.StructField {

	var fields []reflect.StructField
	var timeFields []reflect.StructField

	//reflect the caller
	v := reflect.ValueOf(caller)
	in := reflect.Indirect(v)
	if in.IsValid() {
		callerType := in.Type()
		// get all Fields of the caller
		for i := 0; i < callerType.NumField(); i++ {
			field := callerType.Field(i)

			// skipping fields which are not exported or have the skip tag.
			// If the embedded type is orm.Model, the time Fields will be added.
			if isUnexportedField(field) ||
				field.Tag.Get(tagName) == tagSkip ||
				(implementsInterface(field) && relations == false) ||
				(!implementsInterface(field) && relations == true) {
				// adding CreatedAt, UpdatedAt, DeletedAt as embedded fields
				if field.Type.String() == model && relations == false {
					af := m.structFields(in.FieldByName(field.Name).Interface(), false)
					timeFields = append(timeFields, af...)
				}
				continue
			}

			// adding embedded struct fields.
			if field.Anonymous && relations == false {
				af := m.structFields(in.FieldByName(field.Name).Interface(), false)
				fields = append(fields, af...)
				continue
			}

			// adding all exported native types.
			fields = append(fields, field)
		}
	}

	// appending the time fields at the end.
	return append(fields, timeFields...)
}

// isUnexportedField returns true if its the orm.Model struct or an unexported field.
func isUnexportedField(field reflect.StructField) bool {
	if field.Type.String() == model || field.PkgPath != "" {
		return true
	}
	return false
}

// configFieldByTag is parsing the predefined tags.
func configFieldByTag(f *Field, tag string) {
	// skip if there is no defined tag
	if tag == "" {
		return
	}

	for k, v := range parseTags(tag) {
		switch k {
		case tagCustom:
			f.Custom = true
		case tagPrimary:
			f.Information.PrimaryKey = true
		case tagColumn:
			f.Information.Name = v
		case tagPermission:
			f.Permission.Read = false
			f.Permission.Write = false
			if strings.Contains(v, "r") {
				f.Permission.Read = true
			}
			if strings.Contains(v, "w") {
				f.Permission.Write = true
			}
		case tagSelect:
			f.SqlSelect = v
		}
	}
}

// parseTags returns all defined keys by map.
// If there is only the key set, the value will be an empty string.
// syntax: `column:abc;fk:rel`
// TODO this could be global written with one more map level - maybe useful in grid? = new package struct.ParseTag?
func parseTags(tag string) map[string]string {

	if tag == "" {
		return nil
	}

	// remove spaces and trailing separator
	tag = strings.TrimSpace(tag)
	if tag[len(tag)-1:] == tagSeparator {
		tag = tag[0 : len(tag)-1]
	}

	// configure model
	values := map[string]string{}
	for _, t := range strings.Split(tag, tagSeparator) {
		tag := strings.Split(t, tagKeyValue)
		if len(tag) != 2 {
			tag = append(tag, "")
		}
		if tag[0] == "" {
			continue
		}

		// remove spaces
		tag[0] = strings.TrimSpace(tag[0])
		tag[1] = strings.TrimSpace(tag[1])

		values[tag[0]] = tag[1]
	}

	return values
}
