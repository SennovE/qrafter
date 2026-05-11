package qrafter

import (
	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type Column[T any] struct {
	Name  string
	Table core.TableRef
}

var _ = (core.Selecter)(Column[int]{})

func (c *Column[T]) Bind(name string, table core.TableRef) {
	c.Name = name
	c.Table = table
}

func (c Column[T]) Tables() core.TablesSet {
	return core.TablesSet{c.Table: struct{}{}}
}

func (c Column[T]) Render(d dialect.DialectRenderer) string {
	return d.QuoteIdent(c.Table.SQLName()) + "." + d.QuoteIdent(c.Name)
}
