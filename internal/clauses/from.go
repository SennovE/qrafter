package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type FromClause struct {
	Tables core.TablesSet
	Joins  []JoinClause
}

func UpdateTables[T core.Selecter](c *FromClause, others []T) {
	tables := make([]core.TablesSet, len(others)+1)
	for i := 0; i < len(others); i++ {
		tables[i] = others[i].Tables()
	}
	tables[len(tables)-1] = c.Tables
	c.Tables = utils.UnionSets(tables...)
	c.removeJoinedTables()
}

func (c *FromClause) AddJoin(joinType string, table core.TableRef, predicates ...core.Predicater) {
	c.Joins = append(c.Joins, JoinClause{
		Type:       joinType,
		Table:      table,
		Predicates: predicates,
	})
	c.removeJoinedTables()
}

func (c *FromClause) removeJoinedTables() {
	for _, join := range c.Joins {
		delete(c.Tables, join.Table)
	}
}
