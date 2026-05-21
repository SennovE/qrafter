package qrafter

import (
	"strings"

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
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(dialect.UnsupportedFeatureError); ok {
				err = e
				return
			}
			panic(r)
		}
	}()
	sql, args = q.MustRender(d)
	return
}

// MustRender is like Render but panics if the query is invalid.
func (q CompoundQuery) MustRender(d dialect.Renderer) (sql string, args []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	withCl := clauses.WithClause{}.WithClauseFor(q)
	withCl.Render(&w, renderer)
	q.RenderQueryExpression(&w, renderer)

	return w.String(), renderer.Args()
}

// RenderQueryExpression writes the compound query body.
func (q CompoundQuery) RenderQueryExpression(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()
	state.left.RenderSetOperand(w, d)
	w.WriteString("\n")
	w.WriteString(state.operator.String())
	w.WriteString("\n")
	state.right.RenderSetOperand(w, d)
	state.orderByCl.Render(w, d)
	state.limitOffsetCl.Render(w, d)
}

// RenderSetOperand writes the compound query as a parenthesized set operand.
func (q CompoundQuery) RenderSetOperand(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("(")
	q.RenderQueryExpression(w, d)
	w.WriteString(")")
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
	q.state = &state
	return q
}
