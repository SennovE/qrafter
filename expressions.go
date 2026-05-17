package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
)

// Expression represents a SQL value expression that can be selected or compared.
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

// Render writes the SQL representation of the expression.
func (e Expression) Render(w *strings.Builder, d dialect.Renderer) {
	e.selecter.Render(w, d)
}

// Tables returns table references used by the expression.
func (e Expression) Tables() core.TablesSet {
	return e.selecter.Tables()
}

// Precedence returns the expression precedence used for parenthesizing SQL.
func (e Expression) Precedence() int {
	if _, ok := e.selecter.(TableRefer); ok {
		return core.PrecedenceMultiplicative + 1
	}
	if p, ok := e.selecter.(core.Precedencer); ok {
		return p.Precedence()
	}
	return core.PrecedenceMultiplicative + 1
}

// As returns the expression with a SQL alias.
func (e Expression) As(alias string) Expression {
	return newExpression(expr.Alias(e.selecter, alias))
}

// Add returns an addition expression.
func (e Expression) Add(v any) Expression {
	return newExpression(expr.Binary("+", e.selecter, asSelecter(v)))
}

// Sub returns a subtraction expression.
func (e Expression) Sub(v any) Expression {
	return newExpression(expr.Binary("-", e.selecter, asSelecter(v)))
}

// Mul returns a multiplication expression.
func (e Expression) Mul(v any) Expression {
	return newExpression(expr.Binary("*", e.selecter, asSelecter(v)))
}

// Div returns a division expression.
func (e Expression) Div(v any) Expression {
	return newExpression(expr.Binary("/", e.selecter, asSelecter(v)))
}

// Asc returns an ascending ORDER BY expression.
func (e Expression) Asc() Order {
	return Asc(e)
}

// Desc returns a descending ORDER BY expression.
func (e Expression) Desc() Order {
	return Desc(e)
}

// Literal returns an expression rendered inline using the dialect's literal rules.
func Literal(v any) Expression {
	return newExpression(expr.Literal(v))
}

// Param returns an expression rendered as a placeholder with a bound argument.
func Param(v any) Expression {
	return newExpression(expr.Param(v))
}

// Star returns a '*' expression.
func Star() Expression {
	return newExpression(expr.Star())
}

// Distinct wraps an expression in DISTINCT.
func Distinct(v any) Expression {
	return newExpression(expr.Distinct(asSelecter(v)))
}

// Func builds a SQL function call expression.
func Func(name string, args ...any) Expression {
	return newExpression(expr.Function(name, asSelecters(args)...))
}
