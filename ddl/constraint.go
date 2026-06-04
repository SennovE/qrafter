package ddl

// ReferenceAction is an ON DELETE or ON UPDATE action.
type ReferenceAction string

const (
	// NoAction renders NO ACTION.
	NoAction ReferenceAction = "NO ACTION"
	// Restrict renders RESTRICT.
	Restrict ReferenceAction = "RESTRICT"
	// Cascade renders CASCADE.
	Cascade ReferenceAction = "CASCADE"
	// SetNull renders SET NULL.
	SetNull ReferenceAction = "SET NULL"
	// SetDefault renders SET DEFAULT.
	SetDefault ReferenceAction = "SET DEFAULT"
)

// TableConstraint is a table-level constraint inside CREATE or ALTER TABLE.
type TableConstraint interface {
	tableConstraint()
}

type constraintKind interface {
	constraintKind()
}

// Constraint describes a table-level constraint.
type Constraint[T constraintKind] struct {
	name *string
	c    T
}

func (Constraint[T]) tableConstraint() {}

// Named returns a copy of the constraint with an explicit SQL name.
func (c Constraint[T]) Named(name string) Constraint[T] {
	c.name = &name
	return c
}

type primaryKey struct {
	columns []string
}

func (primaryKey) constraintKind() {}

// PrimaryKey creates a table-level PRIMARY KEY constraint.
func PrimaryKey(columns ...string) Constraint[primaryKey] {
	return Constraint[primaryKey]{c: primaryKey{columns: columns}}
}

type unique struct {
	columns []string
}

func (unique) constraintKind() {}

// Unique creates a table-level UNIQUE constraint.
func Unique(columns ...string) Constraint[unique] {
	return Constraint[unique]{c: unique{columns: columns}}
}

type check struct {
	expr Predicate
}

func (check) constraintKind() {}

// Check creates a table-level CHECK constraint.
func Check(expr Predicate) Constraint[check] {
	return Constraint[check]{c: check{expr: expr}}
}

// ForeignKeyConstraint builds a table-level FOREIGN KEY constraint.
type ForeignKeyConstraint struct {
	Constraint[foreignKey]
}

type foreignKey struct {
	srcCols   []string
	reference *foreignKeyReference
	options   *foreignKeyOptions
}

func (foreignKey) constraintKind() {}

type foreignKeyReference struct {
	table   string
	columns []string
}

type foreignKeyOptions struct {
	onDelete *ReferenceAction
	onUpdate *ReferenceAction
}

// ForeignKey creates a table-level FOREIGN KEY constraint.
func ForeignKey(columns ...string) ForeignKeyConstraint {
	return ForeignKeyConstraint{
		Constraint: Constraint[foreignKey]{
			c: foreignKey{
				srcCols: columns,
			},
		},
	}
}

// References sets the referenced table and columns for a FOREIGN KEY.
func (c ForeignKeyConstraint) References(table string, columns ...string) ForeignKeyConstraint {
	c.c.reference = &foreignKeyReference{
		table:   table,
		columns: append([]string(nil), columns...),
	}
	return c
}

// OnDelete sets the foreign key ON DELETE action.
func (c ForeignKeyConstraint) OnDelete(action ReferenceAction) ForeignKeyConstraint {
	options := c.c.cloneOptions()
	options.onDelete = &action
	c.c.options = options
	return c
}

// OnUpdate sets the foreign key ON UPDATE action.
func (c ForeignKeyConstraint) OnUpdate(action ReferenceAction) ForeignKeyConstraint {
	options := c.c.cloneOptions()
	options.onUpdate = &action
	c.c.options = options
	return c
}

func (c foreignKey) cloneOptions() *foreignKeyOptions {
	if c.options == nil {
		return &foreignKeyOptions{}
	}
	options := *c.options
	return &options
}
