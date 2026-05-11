package query

import (
	"fmt"
	"sort"
	"strings"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/expr"
	"github.com/SennovE/qrafter/pred"
	"github.com/SennovE/qrafter/utils"
)

type SelectQuery struct {
	colums     []expr.Selecter
	tables     map[qrafter.TableRef]struct{}
	predicates []pred.Predicater
	limit      int
	offset     int
}

func Select(cols ...expr.Selecter) SelectQuery {
	q := SelectQuery{
		colums: cols,
		tables: make(map[qrafter.TableRef]struct{}, 0),
	}

	tables := make([]qrafter.TablesSet, len(cols))
	for i := 0; i < len(cols); i++ {
		tables[i] = cols[i].Tables()
	}
	q.tables = utils.UnionSets(tables...)

	return q
}

func (q SelectQuery) Where(predicates ...pred.Predicater) SelectQuery {
	tables := make([]qrafter.TablesSet, len(predicates))
	for i := 0; i < len(predicates); i++ {
		tables[i] = predicates[i].Tables()
	}
	tables = append(tables, q.tables)
	q.tables = utils.UnionSets(tables...)

	q.predicates = append(q.predicates, predicates...)
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
	for i, table := range getSortedTables(q.tables) {
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

func getSortedTables(tables map[qrafter.TableRef]struct{}) []qrafter.TableRef {
	sortedTables := make([]qrafter.TableRef, 0, len(tables))
	for table := range tables {
		sortedTables = append(sortedTables, table)
	}
	sort.Slice(sortedTables, func(i, j int) bool {
		if sortedTables[i].Name == sortedTables[j].Name {
			return sortedTables[i].Alias < sortedTables[j].Alias
		}
		return sortedTables[i].Name < sortedTables[j].Name
	})
	return sortedTables
}
