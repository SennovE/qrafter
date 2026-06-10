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
	Name *string
	Kind T
}

func (Constraint[T]) tableConstraint() {}

// Named returns a copy of the constraint with an explicit SQL name.
func (c Constraint[T]) Named(name string) Constraint[T] {
	c.Name = &name
	return c
}

// PrimaryKeyKind stores PRIMARY KEY constraint data.
type PrimaryKeyKind struct {
	Columns []string
}

func (PrimaryKeyKind) constraintKind() {}

// PrimaryKey creates a table-level PRIMARY KEY constraint.
func PrimaryKey(columns ...string) Constraint[PrimaryKeyKind] {
	return Constraint[PrimaryKeyKind]{Kind: PrimaryKeyKind{Columns: append([]string(nil), columns...)}}
}

// UniqueKind stores UNIQUE constraint data.
type UniqueKind struct {
	Columns []string
}

func (UniqueKind) constraintKind() {}

// Unique creates a table-level UNIQUE constraint.
func Unique(columns ...string) Constraint[UniqueKind] {
	return Constraint[UniqueKind]{Kind: UniqueKind{Columns: append([]string(nil), columns...)}}
}

// CheckKind stores CHECK constraint data.
type CheckKind struct {
	Expr Predicate
}

func (CheckKind) constraintKind() {}

// Check creates a table-level CHECK constraint.
func Check(expr Predicate) Constraint[CheckKind] {
	return Constraint[CheckKind]{Kind: CheckKind{Expr: expr}}
}

// ForeignKeyConstraint builds a table-level FOREIGN KEY constraint.
type ForeignKeyConstraint struct {
	Constraint[ForeignKeyKind]
}

// ForeignKeyKind stores FOREIGN KEY constraint data.
type ForeignKeyKind struct {
	SourceColumns []string
	Reference     *ForeignKeyReference
	Options       *ForeignKeyOptions
}

func (ForeignKeyKind) constraintKind() {}

// ForeignKeyReference stores the referenced table and columns.
type ForeignKeyReference struct {
	Table   string
	Columns []string
}

// ForeignKeyOptions stores ON DELETE and ON UPDATE actions.
type ForeignKeyOptions struct {
	OnDelete *ReferenceAction
	OnUpdate *ReferenceAction
}

// ForeignKey creates a table-level FOREIGN KEY constraint.
func ForeignKey(columns ...string) ForeignKeyConstraint {
	return ForeignKeyConstraint{
		Constraint: Constraint[ForeignKeyKind]{
			Kind: ForeignKeyKind{
				SourceColumns: append([]string(nil), columns...),
			},
		},
	}
}

// References sets the referenced table and columns for a FOREIGN KEY.
func (c ForeignKeyConstraint) References(table string, columns ...string) ForeignKeyConstraint {
	c.Kind.Reference = &ForeignKeyReference{
		Table:   table,
		Columns: append([]string(nil), columns...),
	}
	return c
}

// OnDelete sets the foreign key ON DELETE action.
func (c ForeignKeyConstraint) OnDelete(action ReferenceAction) ForeignKeyConstraint {
	options := c.Kind.cloneOptions()
	options.OnDelete = &action
	c.Kind.Options = options
	return c
}

// OnUpdate sets the foreign key ON UPDATE action.
func (c ForeignKeyConstraint) OnUpdate(action ReferenceAction) ForeignKeyConstraint {
	options := c.Kind.cloneOptions()
	options.OnUpdate = &action
	c.Kind.Options = options
	return c
}

// Named returns a copy of the constraint with an explicit SQL name.
func (c ForeignKeyConstraint) Named(name string) ForeignKeyConstraint {
	c.Constraint = c.Constraint.Named(name)
	return c
}

func (c ForeignKeyKind) cloneOptions() *ForeignKeyOptions {
	if c.Options == nil {
		return &ForeignKeyOptions{}
	}
	options := *c.Options
	return &options
}
