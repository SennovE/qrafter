package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type FromClause struct {
	Tables core.TablesSet
}

var _ = (Clauser)(FromClause{})

func (c FromClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	if len(c.Tables) > 0 {
		w.WriteString(" FROM ")

		for i, table := range core.GetSortedTables(c.Tables) {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteString(table.Render(d))
		}
	}
}

func UpdateTables[T core.Selecter](c *FromClause, others []T) {
	tables := make([]core.TablesSet, len(others)+1)
	for i := 0; i < len(others); i++ {
		tables[i] = others[i].Tables()
	}
	tables[len(tables)-1] = c.Tables
	c.Tables = utils.UnionSets(tables...)
}
