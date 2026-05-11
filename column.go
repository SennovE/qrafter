package qrafter

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
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

func (c Column[T]) Render() string {
	return utils.QuoteIdent(c.Table.SQLName()) + "." + utils.QuoteIdent(c.Name)
}
