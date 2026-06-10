package ddl

import "github.com/SennovE/qrafter/dialect"

// AlterTableStmt builds ALTER TABLE statements.
type AlterTableStmt struct {
	table      string
	operations []any
}

// AlterTable starts an ALTER TABLE statement.
func AlterTable(table string) AlterTableStmt {
	return AlterTableStmt{table: table}
}

type renameColumnStmt struct {
	column string
	name   string
}

// RenameColumn appends a RENAME COLUMN operation.
func (s AlterTableStmt) RenameColumn(column, name string) AlterTableStmt {
	s.operations = append(s.operations, renameColumnStmt{column: column, name: name})
	return s
}

type addColumnStmt struct {
	column      ColumnDef
	ifNotExists bool
}

// AddColumn appends an ADD COLUMN operation.
func (s AlterTableStmt) AddColumn(column ColumnDef) AlterTableStmt {
	s.operations = append(s.operations, addColumnStmt{column: column})
	return s
}

// AddColumnIfNotExists appends an ADD COLUMN IF NOT EXISTS operation.
func (s AlterTableStmt) AddColumnIfNotExists(column ColumnDef) AlterTableStmt {
	s.operations = append(s.operations, addColumnStmt{column: column, ifNotExists: true})
	return s
}

type dropColumnStmt struct {
	column   string
	ifExists bool
}

// DropColumn appends a DROP COLUMN operation.
func (s AlterTableStmt) DropColumn(column string) AlterTableStmt {
	s.operations = append(s.operations, dropColumnStmt{column: column})
	return s
}

// DropColumnIfExists appends a DROP COLUMN IF EXISTS operation.
func (s AlterTableStmt) DropColumnIfExists(column string) AlterTableStmt {
	s.operations = append(s.operations, dropColumnStmt{column: column, ifExists: true})
	return s
}

type alterColumnTypeStmt struct {
	column string
	typ    Type
}

// AlterColumnType appends a column type change.
func (s AlterTableStmt) AlterColumnType(column string, typ Type) AlterTableStmt {
	s.operations = append(s.operations, alterColumnTypeStmt{column: column, typ: typ})
	return s
}

type notNullOperation int

const (
	setNotNull notNullOperation = iota + 1
	dropNotNull
)

type changeNotNullStmt struct {
	column string
	op     notNullOperation
}

// SetNotNull appends an ALTER COLUMN SET NOT NULL operation.
func (s AlterTableStmt) SetNotNull(column string) AlterTableStmt {
	s.operations = append(s.operations, changeNotNullStmt{column: column, op: setNotNull})
	return s
}

// DropNotNull appends an ALTER COLUMN DROP NOT NULL operation.
func (s AlterTableStmt) DropNotNull(column string) AlterTableStmt {
	s.operations = append(s.operations, changeNotNullStmt{column: column, op: dropNotNull})
	return s
}

type setDefaultStmt struct {
	column string
	isExpr bool
	value  any
	expr   string
}

// SetDefault appends an ALTER COLUMN SET DEFAULT operation using a literal.
func (s AlterTableStmt) SetDefault(column string, value any) AlterTableStmt {
	s.operations = append(s.operations, setDefaultStmt{column: column, value: value})
	return s
}

// SetDefaultExpr appends an ALTER COLUMN SET DEFAULT operation using raw SQL.
func (s AlterTableStmt) SetDefaultExpr(column, expr string) AlterTableStmt {
	s.operations = append(s.operations, setDefaultStmt{column: column, isExpr: true, expr: expr})
	return s
}

type dropDefaultStmt struct {
	column string
}

// DropDefault appends an ALTER COLUMN DROP DEFAULT operation.
func (s AlterTableStmt) DropDefault(column string) AlterTableStmt {
	s.operations = append(s.operations, dropDefaultStmt{column: column})
	return s
}

type addConstraintStmt struct {
	table string
	c     TableConstraint
}

// AddConstraint appends an ADD CONSTRAINT operation.
func (s AlterTableStmt) AddConstraint(constraint TableConstraint) AlterTableStmt {
	s.operations = append(s.operations, addConstraintStmt{table: s.table, c: constraint})
	return s
}

type dropConstraintStmt struct {
	name     string
	ifExists bool
}

// DropConstraint appends a DROP CONSTRAINT operation.
func (s AlterTableStmt) DropConstraint(name string) AlterTableStmt {
	s.operations = append(s.operations, dropConstraintStmt{name: name})
	return s
}

// DropConstraintIfExists appends a DROP CONSTRAINT IF EXISTS operation.
func (s AlterTableStmt) DropConstraintIfExists(name string) AlterTableStmt {
	s.operations = append(s.operations, dropConstraintStmt{name: name, ifExists: true})
	return s
}

type renameConstraintStmt struct {
	column string
	name   string
}

// RenameConstraint appends a RENAME CONSTRAINT operation.
func (s AlterTableStmt) RenameConstraint(column, name string) AlterTableStmt {
	s.operations = append(s.operations, renameConstraintStmt{column: column, name: name})
	return s
}

// Render renders the ALTER TABLE operations.
func (s AlterTableStmt) Render(d dialect.Renderer) (string, error) {
	return Render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s AlterTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}
