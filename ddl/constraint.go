package ddl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

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

// TableConstraint renders a table-level constraint inside CREATE or ALTER TABLE.
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

// Named returns a copy of the constraint with an explicit SQL name.
func (c Constraint[T]) Named(name string) Constraint[T] {
	c.name = &name
	return c
}

// Render writes the SQL representation of the constraint.
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
		genName := fmt.Sprintf("pk_%s_%s", table, strings.Join(c.columns, "_"))
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
	expr Predicater
}

// Check creates a table-level CHECK constraint.
func Check(expr Predicater) Constraint[check] {
	return Constraint[check]{c: check{expr: expr}}
}

func (c check) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	var tmp strings.Builder
	c.expr.Render(&tmp, d)
	sql := tmp.String()
	if name == nil {
		fields := strings.Fields(strings.ToLower(sql))
		normalized := strings.Join(fields, " ")
		sum := sha256.Sum256([]byte(normalized))
		hsh := hex.EncodeToString(sum[:])[:8]
		genName := fmt.Sprintf("chk_%s_%s", table, hsh)
		name = &genName
	}
	renderConstraint(*name, "CHECK", func() { w.WriteString(sql) }, w, d)
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

func (c foreignKey) Render(table string, name *string, w *strings.Builder, d dialect.Renderer) {
	if c.reference == nil {
		panic("foreign key reference is required")
	}
	if len(c.srcCols) != len(c.reference.columns) {
		panic("the number of columns on the left and right must match")
	}
	if name == nil {
		genName := fmt.Sprintf("fk_%s_%s_%s_%s", table, strings.Join(c.srcCols, "_"), c.reference.table, strings.Join(c.reference.columns, "_"))
		name = &genName
	}
	renderConstraint(*name, "FOREIGN KEY", func() { renderColumnList(w, d, c.srcCols) }, w, d)
	w.WriteString(" REFERENCES ")
	w.WriteString(d.QuoteIdent(c.reference.table))
	w.WriteString(" (")
	renderColumnList(w, d, c.reference.columns)
	w.WriteString(")")
	if c.options != nil && c.options.onDelete != nil {
		w.WriteString(" ON DELETE ")
		w.WriteString(string(*c.options.onDelete))
	}
	if c.options != nil && c.options.onUpdate != nil {
		w.WriteString(" ON UPDATE ")
		w.WriteString(string(*c.options.onUpdate))
	}
}

func renderConstraint(name, opType string, dataRender func(), w *strings.Builder, d dialect.Renderer) {
	fmt.Fprintf(w, "CONSTRAINT %s %s (", d.QuoteIdent(name), opType)
	dataRender()
	w.WriteString(")")
}
