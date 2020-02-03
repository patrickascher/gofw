package orm

// Association is holding a relation information
type Association struct {
	Type             string
	StructTable      *Column
	AssociationTable *Column
	JunctionTable    *JunctionTable
	Polymorphic      string
}

// JunctionTable is holding all information about the database junction table of a many-to-many relation.
type JunctionTable struct {
	Table             string
	StructColumn      string
	AssociationColumn string
}

// Associations is holding all relations in a map.
// The map key will be the struct field which is holding the relation.
type Associations map[string]*Association
