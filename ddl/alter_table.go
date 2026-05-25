package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type alterOperationKind string

const (
	alterAddColumn      alterOperationKind = "add_column"
	alterDropColumn     alterOperationKind = "drop_column"
	alterRenameColumn   alterOperationKind = "rename_column"
	alterColumnType     alterOperationKind = "alter_column_type"
	alterSetNotNull     alterOperationKind = "set_not_null"
	alterDropNotNull    alterOperationKind = "drop_not_null"
	alterSetDefault     alterOperationKind = "set_default"
	alterDropDefault    alterOperationKind = "drop_default"
	alterAddConstraint  alterOperationKind = "add_constraint"
	alterDropConstraint alterOperationKind = "drop_constraint"
)

// AlterTableStmt builds ALTER TABLE statements.
type AlterTableStmt struct {
	table      string
	operations []alterOperation
}

type alterOperation struct {
	kind       alterOperationKind
	column     string
	newColumn  string
	columnDef  ColumnDef
	typ        Type
	value      any
	hasValue   bool
	expr       string
	constraint Constraint
	name       string
}

// AlterTable starts an ALTER TABLE statement.
func AlterTable(table any) AlterTableStmt {
	return AlterTableStmt{table: tableName(table)}
}

// AddColumn appends an ADD COLUMN operation.
func (s AlterTableStmt) AddColumn(column ColumnDef) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterAddColumn, columnDef: column})
}

// DropColumn appends a DROP COLUMN operation.
func (s AlterTableStmt) DropColumn(column any) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterDropColumn, column: columnName(column)})
}

// RenameColumn appends a RENAME COLUMN operation.
func (s AlterTableStmt) RenameColumn(from, to any) AlterTableStmt {
	return s.appendOperation(&alterOperation{
		kind:      alterRenameColumn,
		column:    columnName(from),
		newColumn: columnName(to),
	})
}

// AlterColumnType appends a column type change.
func (s AlterTableStmt) AlterColumnType(column any, typ Type) AlterTableStmt {
	return s.appendOperation(&alterOperation{
		kind:   alterColumnType,
		column: columnName(column),
		typ:    typ,
	})
}

// SetNotNull appends an ALTER COLUMN SET NOT NULL operation.
func (s AlterTableStmt) SetNotNull(column any) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterSetNotNull, column: columnName(column)})
}

// DropNotNull appends an ALTER COLUMN DROP NOT NULL operation.
func (s AlterTableStmt) DropNotNull(column any) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterDropNotNull, column: columnName(column)})
}

// SetDefault appends an ALTER COLUMN SET DEFAULT operation using a literal.
func (s AlterTableStmt) SetDefault(column, value any) AlterTableStmt {
	return s.appendOperation(&alterOperation{
		kind:     alterSetDefault,
		column:   columnName(column),
		value:    value,
		hasValue: true,
	})
}

// SetDefaultExpr appends an ALTER COLUMN SET DEFAULT operation using raw SQL.
func (s AlterTableStmt) SetDefaultExpr(column any, expr string) AlterTableStmt {
	return s.appendOperation(&alterOperation{
		kind:   alterSetDefault,
		column: columnName(column),
		expr:   requireName("default expression", expr),
	})
}

// DropDefault appends an ALTER COLUMN DROP DEFAULT operation.
func (s AlterTableStmt) DropDefault(column any) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterDropDefault, column: columnName(column)})
}

// AddConstraint appends an ADD CONSTRAINT operation.
func (s AlterTableStmt) AddConstraint(constraint Constraint) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterAddConstraint, constraint: constraint})
}

// DropConstraint appends a DROP CONSTRAINT operation.
func (s AlterTableStmt) DropConstraint(name string) AlterTableStmt {
	return s.appendOperation(&alterOperation{kind: alterDropConstraint, name: requireName("constraint", name)})
}

// Render renders the ALTER TABLE operations. Multiple operations are rendered
// as separate statements separated by semicolons.
func (s AlterTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s AlterTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(s.Render(d))
}

func (s AlterTableStmt) appendOperation(operation *alterOperation) AlterTableStmt {
	s.operations = append(s.operations, *operation)
	return s
}

func (s AlterTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if len(s.operations) == 0 {
		panic(fmt.Errorf("ALTER TABLE %q must include at least one operation", s.table))
	}

	for i := range s.operations {
		if i > 0 {
			w.WriteString(";\n")
		}
		renderAlterOperation(w, d, s.table, &s.operations[i])
	}
}

func renderAlterOperation(w *strings.Builder, d dialect.Renderer, table string, operation *alterOperation) {
	w.WriteString("ALTER TABLE ")
	w.WriteString(d.QuoteIdent(table))
	w.WriteString("\n")

	switch operation.kind {
	case alterAddColumn:
		w.WriteString("ADD COLUMN ")
		operation.columnDef.render(w, d)
	case alterDropColumn:
		w.WriteString("DROP COLUMN ")
		w.WriteString(d.QuoteIdent(operation.column))
	case alterRenameColumn:
		w.WriteString("RENAME COLUMN ")
		w.WriteString(d.QuoteIdent(operation.column))
		w.WriteString(" TO ")
		w.WriteString(d.QuoteIdent(operation.newColumn))
	case alterColumnType:
		renderAlterColumnType(w, d, operation)
	case alterSetNotNull:
		renderAlterColumnNullability(w, d, operation.column, "SET NOT NULL")
	case alterDropNotNull:
		renderAlterColumnNullability(w, d, operation.column, "DROP NOT NULL")
	case alterSetDefault:
		renderAlterColumnDefault(w, d, operation)
	case alterDropDefault:
		renderAlterColumnDropDefault(w, d, operation.column)
	case alterAddConstraint:
		renderAlterAddConstraint(w, d, &operation.constraint)
	case alterDropConstraint:
		renderAlterDropConstraint(w, d, operation.name)
	default:
		panic(fmt.Errorf("unsupported ALTER TABLE operation %q", operation.kind))
	}
}

func renderAlterColumnType(w *strings.Builder, d dialect.Renderer, operation *alterOperation) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN TYPE")
	}
	if isMySQL(d) {
		w.WriteString("MODIFY COLUMN ")
		w.WriteString(d.QuoteIdent(operation.column))
		w.WriteString(" ")
		w.WriteString(operation.typ.render(d))
		return
	}

	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(operation.column))
	w.WriteString(" TYPE ")
	w.WriteString(operation.typ.render(d))
}

func renderAlterColumnNullability(w *strings.Builder, d dialect.Renderer, column, action string) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN NULLABILITY")
	}
	if isMySQL(d) {
		unsupported(d, "ALTER COLUMN NULLABILITY")
	}

	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(column))
	w.WriteString(" ")
	w.WriteString(action)
}

func renderAlterColumnDefault(w *strings.Builder, d dialect.Renderer, operation *alterOperation) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN DEFAULT")
	}

	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(operation.column))
	w.WriteString(" SET DEFAULT ")
	if operation.expr != "" {
		w.WriteString(operation.expr)
		return
	}
	if operation.hasValue {
		w.WriteString(d.Literal(operation.value))
	}
}

func renderAlterColumnDropDefault(w *strings.Builder, d dialect.Renderer, column string) {
	if isSQLite(d) {
		unsupported(d, "ALTER COLUMN DEFAULT")
	}

	w.WriteString("ALTER COLUMN ")
	w.WriteString(d.QuoteIdent(column))
	w.WriteString(" DROP DEFAULT")
}

func renderAlterAddConstraint(w *strings.Builder, d dialect.Renderer, constraint *Constraint) {
	if isSQLite(d) {
		unsupported(d, "ALTER TABLE ADD CONSTRAINT")
	}

	w.WriteString("ADD ")
	constraint.render(w, d)
}

func renderAlterDropConstraint(w *strings.Builder, d dialect.Renderer, name string) {
	if isSQLite(d) {
		unsupported(d, "ALTER TABLE DROP CONSTRAINT")
	}
	if isMySQL(d) {
		unsupported(d, "ALTER TABLE DROP CONSTRAINT")
	}

	w.WriteString("DROP CONSTRAINT ")
	w.WriteString(d.QuoteIdent(name))
}
