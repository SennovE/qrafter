package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type Column[T any] struct {
	Expression
	Name  string
	Table core.TableRef
	value T
}

var _ = (core.Selecter)(Column[int]{})

type ColumnMarker interface {
	TableRef() core.TableRef
}

func (c Column[T]) TableRef() core.TableRef {
	return c.Table
}

func (c *Column[T]) Bind(name string, table core.TableRef) {
	c.Name = name
	c.Table = table
	c.Expression = newExpression(c)
}

func (c Column[T]) Tables() core.TablesSet {
	return core.TablesSet{c.Table: struct{}{}}
}

func (c Column[T]) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString(d.QuoteIdent(c.Table.SQLName()))
	w.WriteString(".")
	w.WriteString(d.QuoteIdent(c.Name))
}
