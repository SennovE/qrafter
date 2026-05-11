package expr

import (
	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/utils"
)

type Column[T any] struct {
	Name  string
	Table qrafter.TableRef
}

var _ = (Selecter)(Column[int]{})

func (c *Column[T]) Bind(name string, table qrafter.TableRef) {
	c.Name = name
	c.Table = table
}

func (c Column[T]) Tables() qrafter.TablesSet {
	return qrafter.TablesSet{c.Table: struct{}{}}
}

func (c Column[T]) Render() string {
	return utils.QuoteIdent(c.Table.SQLName()) + "." + utils.QuoteIdent(c.Name)
}
