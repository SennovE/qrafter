package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type columnConstraintKind string

const (
	columnConstraintPrimaryKey columnConstraintKind = "primary_key"
	columnConstraintNotNull    columnConstraintKind = "not_null"
	columnConstraintNull       columnConstraintKind = "null"
	columnConstraintUnique     columnConstraintKind = "unique"
	columnConstraintDefault    columnConstraintKind = "default"
	columnConstraintCheck      columnConstraintKind = "check"
	columnConstraintReferences columnConstraintKind = "references"
)

// ColumnDef describes a column inside CREATE TABLE or ALTER TABLE ADD COLUMN.
type ColumnDef struct {
	name        string
	typ         Type
	constraints []columnConstraint
}

type columnConstraint struct {
	kind       columnConstraintKind
	expr       string
	value      any
	hasValue   bool
	refTable   string
	refColumns []string
}

// Column creates a column definition. Name can be a string or qrafter.Column.
func Column(name any, typ Type) ColumnDef {
	return ColumnDef{
		name: columnName(name),
		typ:  typ,
	}
}

// PrimaryKey marks the column as a primary key.
func (c ColumnDef) PrimaryKey() ColumnDef {
	return c.appendConstraint(&columnConstraint{kind: columnConstraintPrimaryKey})
}

// NotNull adds a NOT NULL constraint.
func (c ColumnDef) NotNull() ColumnDef {
	return c.appendConstraint(&columnConstraint{kind: columnConstraintNotNull})
}

// Null adds an explicit NULL marker.
func (c ColumnDef) Null() ColumnDef {
	return c.appendConstraint(&columnConstraint{kind: columnConstraintNull})
}

// Unique adds a UNIQUE constraint.
func (c ColumnDef) Unique() ColumnDef {
	return c.appendConstraint(&columnConstraint{kind: columnConstraintUnique})
}

// Default adds a literal DEFAULT value rendered through the dialect.
func (c ColumnDef) Default(value any) ColumnDef {
	return c.appendConstraint(&columnConstraint{
		kind:     columnConstraintDefault,
		value:    value,
		hasValue: true,
	})
}

// DefaultExpr adds a raw SQL DEFAULT expression.
func (c ColumnDef) DefaultExpr(expr string) ColumnDef {
	return c.appendConstraint(&columnConstraint{
		kind: columnConstraintDefault,
		expr: requireName("default expression", expr),
	})
}

// Check adds a column-level CHECK expression.
func (c ColumnDef) Check(expr string) ColumnDef {
	return c.appendConstraint(&columnConstraint{
		kind: columnConstraintCheck,
		expr: requireName("check expression", expr),
	})
}

// References adds a column-level foreign key reference.
func (c ColumnDef) References(table any, columns ...any) ColumnDef {
	return c.appendConstraint(&columnConstraint{
		kind:       columnConstraintReferences,
		refTable:   tableName(table),
		refColumns: columnNames(columns),
	})
}

func (c ColumnDef) appendConstraint(constraint *columnConstraint) ColumnDef {
	c.constraints = append(append([]columnConstraint(nil), c.constraints...), *constraint)
	return c
}

func (c ColumnDef) render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(d.QuoteIdent(c.name))
	w.WriteString(" ")
	w.WriteString(c.typ.render(d))
	for i := range c.constraints {
		renderColumnConstraint(w, d, &c.constraints[i])
	}
}

func renderColumnConstraint(w *strings.Builder, d dialect.Renderer, constraint *columnConstraint) {
	switch constraint.kind {
	case columnConstraintPrimaryKey:
		w.WriteString(" PRIMARY KEY")
	case columnConstraintNotNull:
		w.WriteString(" NOT NULL")
	case columnConstraintNull:
		w.WriteString(" NULL")
	case columnConstraintUnique:
		w.WriteString(" UNIQUE")
	case columnConstraintDefault:
		w.WriteString(" DEFAULT ")
		if constraint.expr != "" {
			w.WriteString(constraint.expr)
			return
		}
		if constraint.hasValue {
			w.WriteString(d.Literal(constraint.value))
		}
	case columnConstraintCheck:
		w.WriteString(" CHECK (")
		w.WriteString(constraint.expr)
		w.WriteString(")")
	case columnConstraintReferences:
		w.WriteString(" REFERENCES ")
		w.WriteString(d.QuoteIdent(constraint.refTable))
		if len(constraint.refColumns) > 0 {
			w.WriteString(" (")
			renderColumnList(w, d, constraint.refColumns)
			w.WriteString(")")
		}
	}
}
