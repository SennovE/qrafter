package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/pred"
)

// Predicate represents a SQL boolean predicate.
type Predicate struct {
	predicater core.Predicater
}

var _ core.Predicater = Predicate{}

func newPredicate(p core.Predicater) Predicate {
	return Predicate{predicater: p}
}

func unwrapPredicates(ps []core.Predicater) []core.Predicater {
	res := make([]core.Predicater, len(ps))
	for i, p := range ps {
		if wrapped, ok := p.(Predicate); ok {
			res[i] = wrapped.predicater
			continue
		}
		res[i] = p
	}
	return res
}

// Predicate marks the value as a predicate.
func (p Predicate) Predicate() {}

// Render writes the SQL representation of the predicate.
func (p Predicate) Render(w *strings.Builder, d dialect.Renderer) {
	p.predicater.Render(w, d)
}

// Tables returns table references used by the predicate.
func (p Predicate) Tables() core.TablesSet {
	return p.predicater.Tables()
}

// Precedence returns the predicate precedence used for parenthesizing SQL.
func (p Predicate) Precedence() int {
	if prec, ok := p.predicater.(core.Precedencer); ok {
		return prec.Precedence()
	}
	return core.PrecedenceComparison
}

// And combines predicates with SQL AND.
func And(ps ...core.Predicater) Predicate {
	return newPredicate(pred.Logical(pred.OpAnd, unwrapPredicates(ps)...))
}

// Or combines predicates with SQL OR.
func Or(ps ...core.Predicater) Predicate {
	return newPredicate(pred.Logical(pred.OpOr, unwrapPredicates(ps)...))
}

// Lt returns a less-than predicate.
func (e Expression) Lt(v any) Predicate {
	return newPredicate(pred.Binary("<", e.selecter, asSelecter(v)))
}

// Gt returns a greater-than predicate.
func (e Expression) Gt(v any) Predicate {
	return newPredicate(pred.Binary(">", e.selecter, asSelecter(v)))
}

// Le returns a less-than-or-equal predicate.
func (e Expression) Le(v any) Predicate {
	return newPredicate(pred.Binary("<=", e.selecter, asSelecter(v)))
}

// Ge returns a greater-than-or-equal predicate.
func (e Expression) Ge(v any) Predicate {
	return newPredicate(pred.Binary(">=", e.selecter, asSelecter(v)))
}

// Eq returns an equality predicate.
func (e Expression) Eq(v any) Predicate {
	return newPredicate(pred.Binary("=", e.selecter, asSelecter(v)))
}
