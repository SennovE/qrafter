package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type alterOperationRenderer interface {
	renderAlterOperation(w *strings.Builder, d dialect.Renderer)
}

// AlterTableStmt builds ALTER TABLE statements.
type AlterTableStmt struct {
	table      string
	operations []alterOperationRenderer
}

// AlterTable starts an ALTER TABLE statement.
func AlterTable(table any) AlterTableStmt {
	return AlterTableStmt{table: tableName(table)}
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

func (s renameColumnStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("RENAME COLUMN ")
	w.WriteString(d.QuoteIdent(s.column))
	w.WriteString(" TO ")
	w.WriteString(d.QuoteIdent(s.name))
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

func (s addColumnStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("ADD COLUMN ")
	if s.ifNotExists {
		w.WriteString("IF NOT EXISTS ")
	}
	s.column.Render(w, d)
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

func (s dropColumnStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("DROP COLUMN ")
	if s.ifExists {
		w.WriteString("IF EXISTS ")
	}
	w.WriteString(d.QuoteIdent(s.column))
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

func (s alterColumnTypeStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN TYPE")
	} else if isMySQL(d) {
		w.WriteString("MODIFY COLUMN ")
		w.WriteString(d.QuoteIdent(s.column))
		w.WriteString(" ")
	} else {
		w.WriteString("ALTER COLUMN ")
		w.WriteString(d.QuoteIdent(s.column))
		w.WriteString(" TYPE ")
	}
	s.typ.render(d)
}

type notNullOperation int

const (
	setNotNull notNullOperation = iota + 1
	dropNotNull
)

type changeNotNullStmt struct {
	column string
	op notNullOperation
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

func (s changeNotNullStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) || isMySQL(d) {
		unsupported(d, "ALTER COLUMN NULLABILITY")
	}
	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(s.column))
	switch s.op {
	case setNotNull:
		w.WriteString(" SET NOT NULL")
	case dropNotNull:
		w.WriteString(" DROP NOT NULL")
	}
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
func (s AlterTableStmt) SetDefaultExpr(column string, expr string) AlterTableStmt {
	s.operations = append(s.operations, setDefaultStmt{column: column, isExpr: true, expr: expr})
	return s
}

func (s setDefaultStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN DEFAULT")
	}
	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(s.column))
	w.WriteString(" SET DEFAULT ")
	if s.isExpr {
		w.WriteString(s.expr)
	} else {
		w.WriteString(d.Literal(s.value))
	}
}

type dropDefaultStmt struct {
	column string
}

// DropDefault appends an ALTER COLUMN DROP DEFAULT operation.
func (s AlterTableStmt) DropDefault(column string) AlterTableStmt {
	s.operations = append(s.operations, dropDefaultStmt{column: column})
	return s
}

func (s dropDefaultStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN DEFAULT")
	}
	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(s.column))
	w.WriteString(" DROP DEFAULT")
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

func (s addConstraintStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) {
		unsupported(d, "ALTER TABLE ADD CONSTRAINT")
	}
	w.WriteString("ADD ")
	s.c.Render(s.table, w, d)
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

// DropConstraint appends a DROP CONSTRAINT operation.
func (s AlterTableStmt) DropConstraintIfExists(name string) AlterTableStmt {
	s.operations = append(s.operations, dropConstraintStmt{name: name, ifExists: true})
	return s
}

func (s dropConstraintStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) {
		unsupported(d, "ALTER TABLE DROP CONSTRAINT")
	} else if isMySQL(d) {
		w.WriteString("DROP ")
	} else {
		w.WriteString("DROP CONSTRAINT ")
		if s.ifExists {
			w.WriteString("IF EXISTS ")
		}
	}
	w.WriteString(d.QuoteIdent(s.name))
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

func (s renameConstraintStmt) renderAlterOperation(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("RENAME CONSTRAINT ")
	w.WriteString(d.QuoteIdent(s.column))
	w.WriteString(" TO ")
	w.WriteString(d.QuoteIdent(s.name))
}

// Render renders the ALTER TABLE operations.
func (s AlterTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s AlterTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s AlterTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if len(s.operations) == 0 {
		panic(fmt.Errorf("ALTER TABLE %q must include at least one operation", s.table))
	}
	w.WriteString("ALTER TABLE ")

	for i, op := range s.operations {
		if i > 0 {
			if isSQLite(d) {
				w.WriteString(";\nALTER TABLE ")
			} else {
				w.WriteString(",\n    ")
			}
		}
		op.renderAlterOperation(w, d)
	}
}
