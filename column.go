package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

// Column represents a typed SQL table column and a scan destination for values of type T.
type Column[T any] struct {
	Expression
	name  string
	table core.TableRef
	value T
}

var _ core.Selecter = Column[int]{}

// TableRefer identifies values that carry table reference information.
type TableRefer interface {
	TableRef() core.TableRef
}

// TableRef returns the table reference associated with the column.
func (c Column[T]) TableRef() core.TableRef {
	return c.table
}

// Bind attaches the column to a SQL name and table reference.
func (c *Column[T]) Bind(name string, table core.TableRef) {
	c.name = name
	c.table = table
	c.Expression = newExpression(c)
}

// Tables returns the set containing the column's table reference.
func (c Column[T]) Tables() core.TablesSet {
	return core.TablesSet{c.table: struct{}{}}
}

// Render writes the fully qualified column name.
func (c Column[T]) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(d.QuoteIdent(c.table.SQLName()))
	w.WriteString(".")
	w.WriteString(d.QuoteIdent(c.name))
}
