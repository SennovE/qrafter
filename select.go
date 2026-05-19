package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// SelectQuery represents a SELECT statement under construction.
type SelectQuery struct {
	state *selectQueryState
}

type selectQueryState struct {
	withCl        clauses.WithClause
	selectCl      clauses.SelectClause
	fromCl        clauses.FromClause
	whereCl       clauses.WhereClause
	groupByCl     clauses.GroupByClause
	havingCl      clauses.HavingClause
	orderByCl     clauses.OrderByClause
	limitOffsetCl clauses.LimitOffsetClause
}

// Select starts a SELECT query for the given expressions.
func Select(cols ...core.Selecter) SelectQuery {
	q := SelectQuery{
		state: &selectQueryState{
			selectCl: clauses.SelectClause{Columns: cols},
		},
	}
	clauses.UpdateTables(&q.state.fromCl, cols)
	return q
}

// Where appends predicates to the query's WHERE clause.
func (q SelectQuery) Where(predicates ...core.Predicater) SelectQuery {
	q = q.cloneState()
	clauses.UpdateTables(&q.state.fromCl, predicates)
	q.state.whereCl.Predicates = append(q.state.whereCl.Predicates, predicates...)
	return q
}

// Join adds an INNER JOIN to the query.
func (q SelectQuery) Join(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("JOIN", table, predicates...)
}

// LeftJoin adds a LEFT JOIN to the query.
func (q SelectQuery) LeftJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("LEFT JOIN", table, predicates...)
}

// RightJoin adds a RIGHT JOIN to the query.
func (q SelectQuery) RightJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("RIGHT JOIN", table, predicates...)
}

// FullJoin adds a FULL JOIN to the query.
func (q SelectQuery) FullJoin(table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	return q.join("FULL JOIN", table, predicates...)
}

// CrossJoin adds a CROSS JOIN to the query.
func (q SelectQuery) CrossJoin(table TableConfigProvider) SelectQuery {
	return q.join("CROSS JOIN", table)
}

func (q SelectQuery) join(joinType string, table TableConfigProvider, predicates ...core.Predicater) SelectQuery {
	q = q.cloneState()
	clauses.UpdateTables(&q.state.fromCl, predicates)
	q.state.fromCl.AddJoin(joinType, GetTableRef(table), unwrapPredicates(predicates)...)
	return q
}

// GroupBy appends expressions to the GROUP BY clause.
func (q SelectQuery) GroupBy(cols ...core.Selecter) SelectQuery {
	q = q.cloneState()
	clauses.UpdateTables(&q.state.fromCl, cols)
	q.state.groupByCl.Columns = append(q.state.groupByCl.Columns, cols...)
	return q
}

// Having appends predicates to the HAVING clause.
func (q SelectQuery) Having(predicates ...core.Predicater) SelectQuery {
	q = q.cloneState()
	clauses.UpdateTables(&q.state.fromCl, predicates)
	q.state.havingCl.Predicates = append(q.state.havingCl.Predicates, unwrapPredicates(predicates)...)
	return q
}

// OrderBy appends expressions to the ORDER BY clause.
func (q SelectQuery) OrderBy(items ...core.Selecter) SelectQuery {
	q = q.cloneState()
	clauses.UpdateTables(&q.state.fromCl, items)
	q.state.orderByCl.Items = append(q.state.orderByCl.Items, items...)
	return q
}

// Limit sets a LIMIT clause on the query.
func (q SelectQuery) Limit(l int) SelectQuery {
	q = q.cloneState()
	q.state.limitOffsetCl.Limit = l
	return q
}

// Offset sets an OFFSET clause on the query.
func (q SelectQuery) Offset(o int) SelectQuery {
	q = q.cloneState()
	q.state.limitOffsetCl.Offset = o
	return q
}

// Union combines this query with another query using UNION.
func (q SelectQuery) Union(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, compoundUnion, other)
}

// UnionAll combines this query with another query using UNION ALL.
func (q SelectQuery) UnionAll(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, compoundUnionAll, other)
}

// CTE wraps the query as a common table expression.
func (q SelectQuery) CTE(name string) CommonTableExpression {
	return CommonTableExpression{
		ref: &core.CTERef{
			Name:  name,
			Query: q,
		},
	}
}

// RecursiveCTE wraps the query as a recursive common table expression.
func (q SelectQuery) RecursiveCTE(name string) CommonTableExpression {
	return q.CTE(name).Recursive()
}

// Render renders the query and returns SQL plus bound arguments.
func (q SelectQuery) Render(d dialect.Renderer) (sql string, args []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	state := q.currentState()
	withCl := state.withCl.WithClauseFor(q)
	withCl.Render(&w, renderer)
	q.RenderQueryExpression(&w, renderer)

	return w.String(), renderer.Args()
}

// RenderQueryExpression writes the SELECT query body.
func (q SelectQuery) RenderQueryExpression(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()
	cls := []clauses.Clauser{
		state.selectCl,
		state.fromCl,
		state.whereCl,
		state.groupByCl,
		state.havingCl,
		state.orderByCl,
		state.limitOffsetCl,
	}

	for _, cl := range cls {
		cl.Render(w, d)
	}
}

// RenderSetOperand writes the query as an operand in a set operation.
func (q SelectQuery) RenderSetOperand(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()
	if len(state.orderByCl.Items) > 0 || state.limitOffsetCl.Limit != 0 || state.limitOffsetCl.Offset != 0 {
		w.WriteString("(")
		q.RenderQueryExpression(w, d)
		w.WriteString(")")
		return
	}
	q.RenderQueryExpression(w, d)
}

// CTEs returns common table expressions referenced by the query.
func (q SelectQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	ctes := make([]*core.CTERef, 0)
	for _, table := range core.GetSortedTables(state.fromCl.Tables) {
		if table.CTE != nil {
			ctes = append(ctes, table.CTE)
		}
	}
	for _, join := range state.fromCl.Joins {
		if join.Table.CTE != nil {
			ctes = append(ctes, join.Table.CTE)
		}
	}
	return ctes
}

func (q SelectQuery) currentState() selectQueryState {
	if q.state == nil {
		return selectQueryState{}
	}
	return *q.state
}

func (q SelectQuery) cloneState() SelectQuery {
	state := q.currentState()
	q.state = &state
	return q
}
