package dialect

// DefaultValues is the empty-row INSERT source node.
type DefaultValues struct{}

// Returning is a DML RETURNING clause node.
type Returning struct {
	Items []any
}

// OrderItem is an ORDER BY item node.
type OrderItem struct {
	Expr      any
	Direction string
	Nulls     string
}

// Join is a SELECT join node.
type Join struct {
	Type       string
	Table      any
	Predicates []any
}

// LimitOffset is a LIMIT/OFFSET node.
type LimitOffset struct {
	Limit  int
	Offset int
}

// UpdateTarget is an UPDATE target node.
type UpdateTarget struct {
	Target any
	From   []any
}

// UpdateFrom is an UPDATE source-table clause node.
type UpdateFrom struct {
	From []any
}

// DeleteTarget is a DELETE target node.
type DeleteTarget struct {
	Target     any
	TargetName string
	Using      []any
}

// DeleteUsing is a DELETE source-table clause node.
type DeleteUsing struct {
	Using []any
}

// PartialIndexPredicate is a CREATE INDEX WHERE predicate node.
type PartialIndexPredicate struct {
	Predicate any
}

// DropTableBehavior is a DROP TABLE CASCADE/RESTRICT behavior node.
type DropTableBehavior struct {
	Behavior string
}

// AlterIndexRename is an index rename node.
type AlterIndexRename struct {
	OldName  string
	NewName  string
	Table    string
	HasTable bool
}

// AlterColumnType is an ALTER TABLE column type node.
type AlterColumnType struct {
	Column string
	Type   string
}

// AlterColumnNullability is an ALTER TABLE column nullability node.
type AlterColumnNullability struct {
	Column string
	Set    bool
}

// AlterColumnDefault is an ALTER TABLE column default node.
type AlterColumnDefault struct {
	Column string
	Drop   bool
	IsExpr bool
	Expr   string
	Value  any
}

// AlterTableAddConstraint is an ALTER TABLE ADD CONSTRAINT node.
type AlterTableAddConstraint struct {
	Constraint any
}

// AlterTableDropConstraint is an ALTER TABLE DROP CONSTRAINT node.
type AlterTableDropConstraint struct {
	Name     string
	IfExists bool
}

// AlterTableOperationSeparator is emitted before each ALTER TABLE operation.
type AlterTableOperationSeparator struct {
	Table string
	Index int
}
