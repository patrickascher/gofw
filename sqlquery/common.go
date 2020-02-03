package sqlquery

import (
	"strings"
)

// replacePlaceholders the package placeholder with the db driver placeholder
func (b *Builder) replacePlaceholders(stmt string) string {
	n := strings.Count(stmt, PLACEHOLDER)
	for i := 1; i <= n; i++ {
		stmt = strings.Replace(stmt, PLACEHOLDER, b.Placeholder.placeholder(), 1)
	}

	return stmt
}

// addColumns adding columns from a key/value pair if they were not added manually before
func (b *Builder) addColumns(columns []string, values map[string]interface{}) []string {
	if len(columns) == 0 {
		for column := range values {
			columns = append(columns, column)
		}
	}

	return columns
}

// QuoteIdentifier is used to quote columns and table names
// database.table = `database`.`table`
// database.table alias = `database`.`table` `alias`
func (b *Builder) QuoteIdentifier(name string) string {

	// empty string
	if name == "" {
		return name
	}

	var rv string

	// all quote characters are getting replaced in the given name
	// the fact that we only use this for identifiers this is OK.
	// - values are added as placeholder, so the driver is handling the values

	if name[0:1] == "!" {
		return name[1:]
	}

	name = strings.Replace(name, b.quote, "", -1)
	alias := strings.Split(name, " ")

	var sp []string
	sp = strings.Split(alias[0], ".")

	for _, i := range sp {
		if rv != "" {
			rv += "."
		}
		rv += b.quote + i + b.quote
	}
	if len(alias) >= 2 {
		rv += " " + b.QuoteIdentifier(alias[len(alias)-1])
	}

	return rv
}

// escapeColumns is used to escape a slice of column names
func (b *Builder) escapeColumns(columns []string) string {
	colStmt := ""
	for _, c := range columns {
		if colStmt != "" {
			colStmt += ", "
		}

		colStmt += b.QuoteIdentifier(c)
	}

	return colStmt
}
