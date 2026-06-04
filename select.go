package qrafter

import (
	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// SelectQuery represents a SELECT statement under construction.
type SelectQuery struct {
	state *selectQueryState
}

var _ core.QueryRenderer = SelectQuery{}

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

// Render renders the query and returns SQL, bound arguments and an error if the query is invalid.
func (q SelectQuery) Render(d dialect.Renderer) (sql string, args []any, err error) {
	return renderQuery(func() (string, []any) {
		return q.MustRender(d)
	})
}

// MustRender is like Render but panics if the query is invalid.
func (q SelectQuery) MustRender(d dialect.Renderer) (sql string, args []any) {
	state := q.currentState()
	return renderStatementWithClause(d, state.withCl, q.CTEs(), q)
}

// CTEs returns common table expressions referenced by the query.
func (q SelectQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	ctes := make([]*core.CTERef, 0)
	seen := make(map[string]struct{})

	ctes = appendCTEsFromTables(ctes, seen, core.GetSortedTables(state.fromCl.Tables))
	for _, join := range state.fromCl.Joins {
		ctes = appendCTEFromTable(ctes, seen, join.Table)
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
	state.withCl.CTEs = append([]*core.CTERef(nil), state.withCl.CTEs...)
	state.selectCl.Columns = append([]core.Selecter(nil), state.selectCl.Columns...)
	state.fromCl = cloneFromClause(state.fromCl)
	state.whereCl.Predicates = append([]core.Predicater(nil), state.whereCl.Predicates...)
	state.groupByCl.Columns = append([]core.Selecter(nil), state.groupByCl.Columns...)
	state.havingCl.Predicates = append([]core.Predicater(nil), state.havingCl.Predicates...)
	state.orderByCl.Items = append([]core.Selecter(nil), state.orderByCl.Items...)
	q.state = &state
	return q
}

func cloneFromClause(cl clauses.FromClause) clauses.FromClause {
	cl.Tables = cloneTablesSet(cl.Tables)
	cl.Joins = append([]clauses.JoinClause(nil), cl.Joins...)
	for i := range cl.Joins {
		cl.Joins[i].Predicates = append([]core.Predicater(nil), cl.Joins[i].Predicates...)
	}
	return cl
}

func cloneTablesSet(tables core.TablesSet) core.TablesSet {
	if len(tables) == 0 {
		return nil
	}

	cloned := make(core.TablesSet, len(tables))
	for table := range tables {
		cloned[table] = struct{}{}
	}
	return cloned
}
