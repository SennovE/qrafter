package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type FromClause struct {
	Tables core.TablesSet
	Joins  []JoinClause
}

var _ = (Clauser)(FromClause{})

func (c FromClause) Render(w *strings.Builder, d dialect.Renderer) {
	if len(c.Tables) == 0 && len(c.Joins) == 0 {
		return
	}

	w.WriteString(" FROM ")

	tables := core.GetSortedTables(c.Tables)
	joins := c.Joins
	if len(tables) == 0 {
		joins[0].Table.Render(w, d)
		joins = joins[1:]
	} else {
		core.RenderWithDelimiter(w, d, ", ", tables)
	}

	for _, join := range joins {
		join.Render(w, d)
	}
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
