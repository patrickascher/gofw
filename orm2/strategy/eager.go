package strategy

import (
	"database/sql"
	"github.com/patrickascher/gofw/orm2"
	"github.com/patrickascher/gofw/sqlquery"
)

func init() {
	_ = orm2.Register("eager", &EagerLoading{})
}

type EagerLoading struct {
}

func (e EagerLoading) First(scope orm2.Scope, c *sqlquery.Condition) error {

	readPerm := orm2.Permission{Read: true}
	b := scope.Builder()

	// get the struct fields for the db scan
	var values []interface{}
	for _, col := range scope.Fields(readPerm) {
		values = append(values, scope.CallerField(col.Name).Addr().Interface())
	}

	// build select
	row, err := b.Select(scope.TableName()).Columns(scope.Columns(readPerm, true)...).Condition(c).First()
	if err != nil {
		return err
	}

	// scan all variables to fill it with values
	err = row.Scan(values...)
	if err != nil {
		// TODO check if there is any reason why we should not return error on no rows.
		//if err.Error() != "sql: no rows in result set" {
		//	return err
		//}
		return err
	}

	// relations
	for _, relation := range scope.Relations(readPerm) {

		// init Rel
		// setRelations
		c := &sqlquery.Condition{}
		rel, err := scope.InitCallerRelation(relation.Field)
		if err != nil {
			return err
		}

		switch relation.Kind {
		case orm2.HasOne, orm2.BelongsTo:

			if scope.IsPolymorphic(relation) {
				c.Where(b.QuoteIdentifier(relation.Polymorphic.Field.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
				c.Where(b.QuoteIdentifier(relation.Polymorphic.Type.Information.Name)+" = ?", relation.Polymorphic.Value)
			} else {
				c.Where(b.QuoteIdentifier(relation.AssociationForeignKey.Information.Name)+" = ?", scope.CallerField(relation.ForeignKey.Name).Interface())
			}

			err = rel.First(c)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
		case orm2.HasMany, orm2.ManyToMany:

		}
	}

	return nil
}
func (e EagerLoading) All(interface{}, orm2.Scope, *sqlquery.Condition) error {
	return nil
}
func (e EagerLoading) Create(orm2.Scope) error {
	return nil
}
func (e EagerLoading) Update(orm2.Scope, *sqlquery.Condition) error {
	return nil
}
func (e EagerLoading) Delete(orm2.Scope, *sqlquery.Condition) error {
	return nil
}
