package qrafter

import (
	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// CompoundQuery represents a set operation such as UNION or UNION ALL.
type CompoundQuery struct {
	state *compoundQueryState
}

var _ core.QueryRenderer = CompoundQuery{}

type compoundQueryState struct {
	left          core.QueryExpression
	right         core.QueryExpression
	operator      compoundOperator
	orderByCl     clauses.OrderByClause
	limitOffsetCl clauses.LimitOffsetClause
}

type compoundOperator uint8

const (
	compoundUnion compoundOperator = iota
	compoundUnionAll
)

func (o compoundOperator) String() string {
	switch o {
	case compoundUnionAll:
		return "UNION ALL"
	default:
		return "UNION"
	}
}

func newCompoundQuery(left core.QueryExpression, operator compoundOperator, right core.QueryExpression) CompoundQuery {
	return CompoundQuery{
		state: &compoundQueryState{
			left:     left,
			operator: operator,
			right:    right,
		},
	}
}

// OrderBy appends a final ORDER BY clause to the compound query.
func (q CompoundQuery) OrderBy(items ...core.Selecter) CompoundQuery {
	q = q.cloneState()
	q.state.orderByCl.Items = append(q.state.orderByCl.Items, items...)
	return q
}

// Limit sets a final LIMIT clause on the compound query.
func (q CompoundQuery) Limit(l int) CompoundQuery {
	q = q.cloneState()
	q.state.limitOffsetCl.Limit = l
	return q
}

// Offset sets a final OFFSET clause on the compound query.
func (q CompoundQuery) Offset(o int) CompoundQuery {
	q = q.cloneState()
	q.state.limitOffsetCl.Offset = o
	return q
}

// Union combines this query with another query using UNION.
func (q CompoundQuery) Union(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, compoundUnion, other)
}

// UnionAll combines this query with another query using UNION ALL.
func (q CompoundQuery) UnionAll(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, compoundUnionAll, other)
}

// CTE wraps the compound query as a common table expression.
func (q CompoundQuery) CTE(name string) CommonTableExpression {
	return CommonTableExpression{
		ref: &core.CTERef{
			Name:  name,
			Query: q,
		},
	}
}

// RecursiveCTE wraps the compound query as a recursive common table expression.
func (q CompoundQuery) RecursiveCTE(name string) CommonTableExpression {
	return q.CTE(name).Recursive()
}

// Render renders the query and returns SQL, bound arguments and an error if the query is invalid.
func (q CompoundQuery) Render(d dialect.Renderer) (sql string, args []any, err error) {
	return renderQuery(func() (string, []any) {
		return q.MustRender(d)
	})
}

// MustRender is like Render but panics if the query is invalid.
func (q CompoundQuery) MustRender(d dialect.Renderer) (sql string, args []any) {
	return renderStatementWithClause(d, clauses.WithClause{}, q.CTEs(), q)
}

// CTEs returns common table expressions referenced by the compound query.
func (q CompoundQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	ctes := state.left.CTEs()
	ctes = append(ctes, state.right.CTEs()...)
	return ctes
}

func (q CompoundQuery) currentState() compoundQueryState {
	if q.state == nil {
		return compoundQueryState{}
	}
	return *q.state
}

func (q CompoundQuery) cloneState() CompoundQuery {
	state := q.currentState()
	state.orderByCl.Items = append([]core.Selecter(nil), state.orderByCl.Items...)
	q.state = &state
	return q
}
