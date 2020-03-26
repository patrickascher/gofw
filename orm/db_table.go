package orm

import (
	"fmt"
	"github.com/patrickascher/gofw/sqlquery"
	"strings"
)

// Permission constants
const (
	READDB = iota
	READVIEW
	READALL //for relations
	WRITEDB
)

// Table is representing a db table.
type Table struct {
	Builder *sqlquery.Builder

	Name         string
	Database     string
	Cols         []*Column
	Associations Associations

	softDelete *Column
	strategy   Strategy
}

// Columns is returning all Columns of the model.
// Its taking care of the white/blacklist and privileger.
func (t *Table) Columns(privilege int) []*Column {
	var cols []*Column
	switch privilege {
	case READDB:
		for _, col := range t.Cols {
			if col.Permission.Read {
				cols = append(cols, col)
			}
		}
	case READVIEW:
		for _, col := range t.Cols {
			if col.Permission.Read || (col.Information.Type != nil && col.Information.Type.Kind() == CustomImpl) {
				cols = append(cols, col)
			}
		}
	case WRITEDB:
		for _, col := range t.Cols {
			if col.Permission.Write {
				cols = append(cols, col)
			}
		}
	}

	return cols
}

// Relations is returning all associations of the model.
// Its taking car of the white/blacklist and privilege.
func (t *Table) Relations(wb *WhiteBlackList, privilege int) Associations {

	//if wb == nil {
	//	return t.Associations
	//}

	association := Associations{}
	for name, rel := range t.Associations {

		// skipping custom implementations
		if privilege == READDB && (rel.Type == CustomStruct || rel.Type == CustomSlice) {
			continue
		}

		// add to not break any fk constraints
		// this is needed if there is a BelongsTo relation, a whitelist is used but the reference field is not set!
		// TODO improvements, now the whole belongsTo is updated, but in theory only the reference key has to get updated.
		if privilege == WRITEDB && rel.Type == BelongsTo {
			association[name] = rel
		}

		if wb == nil || !wb.isRelationDisabled(name) {
			association[name] = rel
		}
	}

	return association
}

// getDatabaseAndTableByString returns the database and table name of a string.
// If the string does not contain any . notation, it will assume that only the table name is set.
func getDatabaseAndTableByString(t string) (database string, table string) {
	tn := strings.Split(t, ".")
	if len(tn) == 2 {
		return strings.TrimSpace(tn[0]), strings.TrimSpace(tn[1])
	}
	return "", strings.TrimSpace(tn[0])
}

// columnNames return all defined column names of the table.
// used to select only needed columns.
// if a customSelects is enabled, the custom column sql are added if not empty.
func (t *Table) columnNames(privilege int, customSelects bool) []string {
	columns := t.Columns(privilege)
	names := make([]string, len(columns))
	for i, cols := range columns {
		tmpName := cols.Information.Name
		if customSelects && cols.SqlSelect != "" {
			tmpName = sqlquery.Raw("(" + cols.SqlSelect + ") AS " + t.Builder.QuoteIdentifier(cols.Information.Name))
		}
		names[i] = tmpName
	}
	return names
}

// columnByName returns a ptr to the given column.
// error will return if the column does not exist.
func (t *Table) columnByName(colName string) (*Column, error) {
	for _, col := range t.Cols {
		if col.Information.Name == colName {
			return col, nil
		}
	}
	return nil, fmt.Errorf(ErrModelColumnNotFound.Error(), colName, t.Name)
}

// columnNamesWithoutAutoincrement returns all columns except the autoincrement column.
func (t *Table) columnNamesWithoutAutoincrement() []string {
	names := make([]string, 0)
	for _, cols := range t.Cols {
		if cols.Information.Autoincrement {
			continue
		}
		names = append(names, cols.Information.Name)
	}
	return names
}

// primaryKeys returns a slice of columns which are the primary key in the database table.
func (t *Table) PrimaryKeys() []*Column {
	var pkeys []*Column
	for _, col := range t.Cols {
		if col.Information.PrimaryKey {
			pkeys = append(pkeys, col)
		}
	}
	return pkeys
}

// describe table and add the columns to the table.
// Only the exported struct fields are added to the table columns.
// Return an error if the structField ID is missing or its not a primary key in the database.
func (t *Table) describe() error {
	// describe table columns
	var cols []string
	for _, col := range t.Cols {
		cols = append(cols, col.Information.Name)
	}
	describeCols, err := t.Builder.Information(t.Database + "." + t.Name).Describe(cols...)
	if err != nil {
		return err
	}

	// checking if table has primary key
	hasPrimary := false

	// adding database column structure to the column
Columns:
	for _, col := range t.Cols {
		for i, describeCol := range describeCols {
			if describeCol.Name == col.Information.Name {
				col.Information = describeCol
				if describeCol.PrimaryKey && col.StructField == "ID" {
					hasPrimary = true
				}
				//decrease columns
				describeCols = append(describeCols[:i], describeCols[i+1:]...)
				continue Columns
			}
		}

		// If the field does not exist in the DB, the write and read permission is set to false.
		// So this field can be used internally, but will never be posted in any sql query.
		// TODO check if the controller still gets the Field data? whats with the frontend useful?
		col.Permission.Write = false
		col.Permission.Read = false
		//return fmt.Errorf(ErrTableColumnNotFound.Error(), col.Information.Name, t.Name)
	}

	// if no primary was found in the defined columns, an error will return
	if hasPrimary == false {
		return fmt.Errorf(ErrIDPrimary.Error(), t.Name)
	}

	return nil
}
