package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

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

func (q CompoundQuery) OrderBy(items ...core.Selecter) CompoundQuery {
	q.orderByCl.Items = append(q.orderByCl.Items, items...)
	return q
}

func (q CompoundQuery) Limit(l int) CompoundQuery {
	q.limitOffsetCl.Limit = l
	return q
}

func (q CompoundQuery) Offset(o int) CompoundQuery {
	q.limitOffsetCl.Offset = o
	return q
}

func (q CompoundQuery) Union(other SelectQuery) CompoundQuery {
	return newCompoundQuery(q, "UNION", other)
}

func (q CompoundQuery) UnionAll(other SelectQuery) CompoundQuery {
	return newCompoundQuery(q, "UNION ALL", other)
}

func (q CompoundQuery) CTE(name string) CommonTableExpression {
	return CommonTableExpression{
		ref: &core.CTERef{
			Name:  name,
			Query: q,
		},
	}
}

func (q CompoundQuery) RecursiveCTE(name string) CommonTableExpression {
	return q.CTE(name).Recursive()
}

func (q CompoundQuery) Render(d dialect.DialectRenderer) string {
	var w strings.Builder
	withCl := clauses.WithClause{}.WithClauseFor(q)

	withCl.Render(&w, d)
	q.RenderQueryExpression(&w, d)

	return w.String()
}

func (q CompoundQuery) RenderQueryExpression(w *strings.Builder, d dialect.DialectRenderer) {
	q.left.RenderSetOperand(w, d)
	w.WriteString(" ")
	w.WriteString(q.operator)
	w.WriteString(" ")
	q.right.RenderSetOperand(w, d)
	q.orderByCl.Render(w, d)
	q.limitOffsetCl.Render(w, d)
}

func (q CompoundQuery) RenderSetOperand(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString("(")
	q.RenderQueryExpression(w, d)
	w.WriteString(")")
}

func (q CompoundQuery) CTEs() []*core.CTERef {
	ctes := q.left.CTEs()
	ctes = append(ctes, q.right.CTEs()...)
	return ctes
}
