package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

type SelectQuery struct {
	withCl        clauses.WithClause
	selectCl      clauses.SelectClause
	fromCl        clauses.FromClause
	whereCl       clauses.WhereClause
	groupByCl     clauses.GroupByClause
	havingCl      clauses.HavingClause
	orderByCl     clauses.OrderByClause
	limitOffsetCl clauses.LimitOffsetClause
}

func Select(cols ...core.Selecter) SelectQuery {
	q := SelectQuery{
		selectCl: clauses.SelectClause{Columns: cols},
	}
	clauses.UpdateTables(&q.fromCl, cols)
	return q
}

func (q SelectQuery) Where(predicates ...core.Predicater) SelectQuery {
	clauses.UpdateTables(&q.fromCl, predicates)
	q.whereCl.Predicates = append(q.whereCl.Predicates, predicates...)
	return q
}

func (q SelectQuery) Join(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("JOIN", table, predicates...)
}

func (q SelectQuery) LeftJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("LEFT JOIN", table, predicates...)
}

func (q SelectQuery) RightJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("RIGHT JOIN", table, predicates...)
}

func (q SelectQuery) FullJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("FULL JOIN", table, predicates...)
}

func (q SelectQuery) CrossJoin(table TableConfigProvider) SelectQuery {
	return q.join("CROSS JOIN", table)
}

func (q SelectQuery) join(joinType string, table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	clauses.UpdateTables(&q.fromCl, predicates)
	q.fromCl.AddJoin(joinType, GetTableRef(table), unwrapPredicates(predicates)...)
	return q
}

func (q SelectQuery) GroupBy(cols ...core.Selecter) SelectQuery {
	clauses.UpdateTables(&q.fromCl, cols)
	q.groupByCl.Columns = append(q.groupByCl.Columns, cols...)
	return q
}

func (q SelectQuery) Having(predicates ...core.Predicater) SelectQuery {
	clauses.UpdateTables(&q.fromCl, predicates)
	q.havingCl.Predicates = append(q.havingCl.Predicates, unwrapPredicates(predicates)...)
	return q
}

func (q SelectQuery) OrderBy(items ...core.Selecter) SelectQuery {
	clauses.UpdateTables(&q.fromCl, items)
	q.orderByCl.Items = append(q.orderByCl.Items, items...)
	return q
}

func (q SelectQuery) Limit(l int) SelectQuery {
	q.limitOffsetCl.Limit = l
	return q
}

func (q SelectQuery) Offset(o int) SelectQuery {
	q.limitOffsetCl.Offset = o
	return q
}

func (q SelectQuery) Union(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, "UNION", other)
}

func (q SelectQuery) UnionAll(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, "UNION ALL", other)
}

func (q SelectQuery) CTE(name string) CommonTableExpression {
	return CommonTableExpression{
		ref: &core.CTERef{
			Name:  name,
			Query: q,
		},
	}
}

func (q SelectQuery) RecursiveCTE(name string) CommonTableExpression {
	return q.CTE(name).Recursive()
}

func (q SelectQuery) Render(d dialect.DialectRenderer) (string, []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	withCl := q.withCl.WithClauseFor(q)
	withCl.Render(&w, renderer)
	q.RenderQueryExpression(&w, renderer)

	return w.String(), renderer.Args()
}

func (q SelectQuery) RenderQueryExpression(w *strings.Builder, d dialect.DialectRenderer) {
	clauses := []clauses.Clauser{
		q.selectCl,
		q.fromCl,
		q.whereCl,
		q.groupByCl,
		q.havingCl,
		q.orderByCl,
		q.limitOffsetCl,
	}

	for _, cl := range clauses {
		cl.Render(w, d)
	}
}

func (q SelectQuery) RenderSetOperand(w *strings.Builder, d dialect.DialectRenderer) {
	if len(q.orderByCl.Items) > 0 || q.limitOffsetCl.Limit != 0 || q.limitOffsetCl.Offset != 0 {
		w.WriteString("(")
		q.RenderQueryExpression(w, d)
		w.WriteString(")")
		return
	}
	q.RenderQueryExpression(w, d)
}

func (q SelectQuery) CTEs() []*core.CTERef {
	ctes := make([]*core.CTERef, 0)
	for _, table := range core.GetSortedTables(q.fromCl.Tables) {
		if table.CTE != nil {
			ctes = append(ctes, table.CTE)
		}
	}
	for _, join := range q.fromCl.Joins {
		if join.Table.CTE != nil {
			ctes = append(ctes, join.Table.CTE)
		}
	}
	return ctes
}
