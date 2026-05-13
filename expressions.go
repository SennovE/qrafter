package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
)

type Expression struct {
	selecter core.Selecter
}

var _ = (core.Selecter)(Expression{})

func newExpression(s core.Selecter) Expression {
	return Expression{selecter: s}
}

func asSelecter(v any) core.Selecter {
	switch v := v.(type) {
	case Expression:
		return v.selecter
	case core.Selecter:
		return v
	default:
		return expr.Param(v)
	}
}

func asSelecters(values []any) []core.Selecter {
	selecters := make([]core.Selecter, len(values))
	for i, value := range values {
		selecters[i] = asSelecter(value)
	}
	return selecters
}

func (e Expression) Render(w *strings.Builder, d dialect.DialectRenderer) {
	e.selecter.Render(w, d)
}

func (e Expression) Tables() core.TablesSet {
	return e.selecter.Tables()
}

func (e Expression) Precedence() int {
	if _, ok := e.selecter.(ColumnMarker); ok {
		return core.PrecedenceMultiplicative + 1
	}
	if p, ok := e.selecter.(core.Precedencer); ok {
		return p.Precedence()
	}
	return core.PrecedenceMultiplicative + 1
}

func (e Expression) As(alias string) Expression {
	return newExpression(expr.Alias(e.selecter, alias))
}

func (e Expression) Add(v any) Expression {
	return newExpression(expr.Binary("+", e.selecter, asSelecter(v)))
}

func (e Expression) Sub(v any) Expression {
	return newExpression(expr.Binary("-", e.selecter, asSelecter(v)))
}

func (e Expression) Mul(v any) Expression {
	return newExpression(expr.Binary("*", e.selecter, asSelecter(v)))
}

func (e Expression) Div(v any) Expression {
	return newExpression(expr.Binary("/", e.selecter, asSelecter(v)))
}

func (e Expression) Asc() Order {
	return Asc(e)
}

func (e Expression) Desc() Order {
	return Desc(e)
}

func Literal(v any) Expression {
	return newExpression(expr.Literal(v))
}

func Param(v any) Expression {
	return newExpression(expr.Param(v))
}

func Star() Expression {
	return newExpression(expr.Star())
}

func Distinct(v any) Expression {
	return newExpression(expr.Distinct(asSelecter(v)))
}

func Func(name string, args ...any) Expression {
	return newExpression(expr.Function(name, asSelecters(args)...))
}
