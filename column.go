package qrafter

import (
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

type ddlColumn struct {
	Name string
}

// TableRefer identifies values that carry table reference information.
type TableRefer interface {
	TableRef() core.TableRef
}

// ColumnRef identifies a concrete SQL column.
type ColumnRef interface {
	TableRefer
	Name() string
}

// DDLKey returns struct that can be used in TableConfig.Columns
func (c Column[T]) DDLKey() ddlColumn {
	return ddlColumn{Name: c.name}
}

// TableRef returns the table reference associated with the column.
func (c Column[T]) TableRef() core.TableRef {
	return c.table
}

// Name returns the SQL column name associated with the column.
func (c Column[T]) Name() string {
	return c.name
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

func (c Column[T]) insertValue() any {
	return c.value
}
