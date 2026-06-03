package ddl

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/ddl/expr"
	"github.com/SennovE/qrafter/dialect"
)

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

type TableConstraint interface {
	Render(table string, w *strings.Builder, d dialect.Renderer)
}

type constraintRenderer interface {
	Render(table string, name *string, w *strings.Builder, d dialect.Renderer)
}

// Constraint describes a table-level constraint.
type Constraint[T constraintRenderer] struct {
	name *string
	c    T
}

func (c Constraint[T]) Named(name string) Constraint[T] {
	c.name = &name
	return c
}

func (c Constraint[T]) Render(table string, w *strings.Builder, d dialect.Renderer) {
	c.c.Render(table, c.name, w, d)
}

type primaryKey struct {
	columns []string
}

// PrimaryKey creates a table-level PRIMARY KEY constraint.
func PrimaryKey(columns ...string) Constraint[primaryKey] {
	return Constraint[primaryKey]{c: primaryKey{columns: columns}}
}

func (c primaryKey) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	if name == nil {
		genName := fmt.Sprintf("fk_%s_%s", table, strings.Join(c.columns, "_"))
		name = &genName
	}
	renderConstraint(*name, "PRIMARY KEY", func() { renderColumnList(w, d, c.columns) }, w, d)
}

type unique struct {
	columns []string
}

// Unique creates a table-level UNIQUE constraint.
func Unique(columns ...string) Constraint[unique] {
	return Constraint[unique]{c: unique{columns: columns}}
}

func (c unique) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	if name == nil {
		genName := fmt.Sprintf("uq_%s_%s", table, strings.Join(c.columns, "_"))
		name = &genName
	}
	renderConstraint(*name, "UNIQUE", func() { renderColumnList(w, d, c.columns) }, w, d)
}

type check struct {
	expr expr.CheckExperssion
}

// Check creates a table-level CHECK constraint.
func Check(expr expr.CheckExperssion) Constraint[check] {
	return Constraint[check]{c: check{expr: expr}}
}

func (c check) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	var tmp strings.Builder
	c.expr.Render(&tmp, d)
	sql := tmp.String()
	if name == nil {
		fields := strings.Fields(strings.ToLower(sql))
		normalized := strings.Join(fields, " ")
		sum := sha1.Sum([]byte(normalized))
		hsh := hex.EncodeToString(sum[:])[:8]
		genName := fmt.Sprintf("chk_%s_%s", table, hsh)
		name = &genName
	}
	renderConstraint(*name, "CHECK", func() { w.WriteString(sql) }, w, d)
}

type ForeignKeyConstraint struct {
	Constraint[foreignKey]
}

type foreignKey struct {
	srcCols  []string
	refTable string
	refCols  []string
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
	c.c.refTable = table
	c.c.refCols = columns
	return c
}

// OnDelete sets the foreign key ON DELETE action.
func (c ForeignKeyConstraint) OnDelete(action ReferenceAction) ForeignKeyConstraint {
	c.c.onDelete = &action
	return c
}

// OnUpdate sets the foreign key ON UPDATE action.
func (c ForeignKeyConstraint) OnUpdate(action ReferenceAction) ForeignKeyConstraint {
	c.c.onUpdate = &action
	return c
}

func (c foreignKey) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	if len(c.srcCols) != len(c.refCols) {
		panic("the number of columns on the left and right must match")
	}
	if name == nil {
		genName := fmt.Sprintf("fk_%s_%s_%s_%s", table, strings.Join(c.srcCols, "_"), c.refTable, strings.Join(c.refCols, "_"))
		name = &genName
	}
	renderConstraint(*name, "FOREIGN KEY", func() { renderColumnList(w, d, c.srcCols) }, w, d)
	w.WriteString(" REFERENCES")
	w.WriteString(c.refTable)
	w.WriteString("(")
	renderColumnList(w, d, c.refCols)
	w.WriteString(")")
	if c.onDelete != nil {
		w.WriteString(" ON DELETE ")
		w.WriteString(string(*c.onDelete))
	}
	if c.onUpdate != nil {
		w.WriteString(" ON UPDATE ")
		w.WriteString(string(*c.onUpdate))
	}
}

func renderConstraint(name string, typ string, dataRender func(), w *strings.Builder, d dialect.Renderer) {
	fmt.Fprintf(w, "CONSTRAINT %s %s (", name, typ)
	dataRender()
	w.WriteString(")")
}
