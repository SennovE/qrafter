package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

type SelectQuery struct {
	selectCl      clauses.SelectClause
	fromCl        clauses.FromClause
	whereCl       clauses.WhereClause
	limitOffsetCl clauses.LimitOffsetClause
}

func Select(cols ...core.Selecter) SelectQuery {
	q := SelectQuery{
		selectCl: clauses.SelectClause{Colums: cols},
	}
	clauses.UpdateTables(&q.fromCl, cols)
	return q
}

func (q SelectQuery) Where(predicates ...core.Predicater) SelectQuery {
	clauses.UpdateTables(&q.fromCl, predicates)
	q.whereCl.Predicates = append(q.whereCl.Predicates, predicates...)
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

func (q SelectQuery) Render(d dialect.DialectRenderer) string {
	var w strings.Builder

	clauses := []clauses.Clauser{
		q.selectCl,
		q.fromCl,
		q.whereCl,
		q.limitOffsetCl,
	}

	for _, cl := range clauses {
		cl.Render(&w, d)
	}

	return w.String()
}
