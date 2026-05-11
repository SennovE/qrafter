package qrafter

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type SelectQuery struct {
	colums     []core.Selecter
	tables     map[core.TableRef]struct{}
	predicates []core.Predicater
	limit      int
	offset     int
}

func Select(cols ...core.Selecter) SelectQuery {
	q := SelectQuery{
		colums: cols,
		tables: make(map[core.TableRef]struct{}, 0),
	}

	tables := make([]core.TablesSet, len(cols))
	for i := 0; i < len(cols); i++ {
		tables[i] = cols[i].Tables()
	}
	q.tables = utils.UnionSets(tables...)

	return q
}

func (q SelectQuery) Where(predicates ...core.Predicater) SelectQuery {
	tables := make([]core.TablesSet, len(predicates)+1)
	for i := 0; i < len(predicates); i++ {
		tables[i] = predicates[i].Tables()
	}
	tables[len(tables)-1] = q.tables
	q.tables = utils.UnionSets(tables...)

	q.predicates = append(q.predicates, predicates...)
	return q
}

func (q SelectQuery) Limit(l int) SelectQuery {
	q.limit = l
	return q
}

func (q SelectQuery) Offset(o int) SelectQuery {
	q.offset = o
	return q
}

func (q SelectQuery) Render() string {
	var builder strings.Builder
	builder.WriteString("SELECT ")

	for i, col := range q.colums {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(col.Render())
	}

	builder.WriteString(" FROM ")
	for i, table := range core.GetSortedTables(q.tables) {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(table.BuildSQL())
	}

	if len(q.predicates) > 0 {
		builder.WriteString(" WHERE ")
		for i, pred := range q.predicates {
			if i > 0 {
				builder.WriteString(" AND ")
			}
			builder.WriteString(pred.Render())
		}
	}

	if q.limit > 0 {
		fmt.Fprintf(&builder, " LIMIT %d", q.limit)
	}

	if q.offset > 0 {
		fmt.Fprintf(&builder, " OFFSET %d", q.limit)
	}

	return builder.String()
}
