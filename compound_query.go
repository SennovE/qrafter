package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// CompoundQuery represents a set operation such as UNION or UNION ALL.
type CompoundQuery struct {
	left          core.QueryExpression
	operator      string
	right         core.QueryExpression
	orderByCl     clauses.OrderByClause
	limitOffsetCl clauses.LimitOffsetClause
}

func newCompoundQuery(left core.QueryExpression, operator string, right core.QueryExpression) CompoundQuery {
	return CompoundQuery{
		left:     left,
		operator: operator,
		right:    right,
	}
}

// OrderBy appends a final ORDER BY clause to the compound query.
func (q CompoundQuery) OrderBy(items ...core.Selecter) CompoundQuery {
	q.orderByCl.Items = append(q.orderByCl.Items, items...)
	return q
}

// Limit sets a final LIMIT clause on the compound query.
func (q CompoundQuery) Limit(l int) CompoundQuery {
	q.limitOffsetCl.Limit = l
	return q
}

// Offset sets a final OFFSET clause on the compound query.
func (q CompoundQuery) Offset(o int) CompoundQuery {
	q.limitOffsetCl.Offset = o
	return q
}

// Union combines this query with another query using UNION.
func (q CompoundQuery) Union(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, "UNION", other)
}

// UnionAll combines this query with another query using UNION ALL.
func (q CompoundQuery) UnionAll(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(q, "UNION ALL", other)
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

// Render renders the compound query and returns SQL plus bound arguments.
func (q CompoundQuery) Render(d dialect.Renderer) (string, []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	withCl := clauses.WithClause{}.WithClauseFor(q)
	withCl.Render(&w, renderer)
	q.RenderQueryExpression(&w, renderer)

	return w.String(), renderer.Args()
}

// RenderQueryExpression writes the compound query body.
func (q CompoundQuery) RenderQueryExpression(w *strings.Builder, d dialect.Renderer) {
	q.left.RenderSetOperand(w, d)
	w.WriteString(" ")
	w.WriteString(q.operator)
	w.WriteString(" ")
	q.right.RenderSetOperand(w, d)
	q.orderByCl.Render(w, d)
	q.limitOffsetCl.Render(w, d)
}

// RenderSetOperand writes the compound query as a parenthesized set operand.
func (q CompoundQuery) RenderSetOperand(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("(")
	q.RenderQueryExpression(w, d)
	w.WriteString(")")
}

// CTEs returns common table expressions referenced by the compound query.
func (q CompoundQuery) CTEs() []*core.CTERef {
	ctes := q.left.CTEs()
	ctes = append(ctes, q.right.CTEs()...)
	return ctes
}
