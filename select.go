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

func (q SelectQuery) Union(other SelectQuery) CompoundQuery {
	return newCompoundQuery(q, "UNION", other)
}

func (q SelectQuery) UnionAll(other SelectQuery) CompoundQuery {
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

func (q SelectQuery) Render(d dialect.DialectRenderer) string {
	var w strings.Builder
	withCl := withClauseFor(q.withCl, q)

	withCl.Render(&w, d)
	q.RenderQueryExpression(&w, d)

	return w.String()
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

func withClauseFor(withCl clauses.WithClause, q core.QueryExpression) clauses.WithClause {
	seen := make(map[string]struct{}, len(withCl.CTEs))
	for _, cte := range withCl.CTEs {
		if cte == nil {
			continue
		}
		seen[cte.Name] = struct{}{}
		if cte.Recursive {
			withCl.Recursive = true
		}
	}

	for _, cte := range q.CTEs() {
		if cte == nil {
			continue
		}
		if cte.Recursive {
			withCl.Recursive = true
		}
		if _, ok := seen[cte.Name]; ok {
			continue
		}
		withCl.CTEs = append(withCl.CTEs, cte)
		seen[cte.Name] = struct{}{}
	}

	return withCl
}
