package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

const (
	precedenceOr = iota + 1
	precedenceAnd
	precedenceComparison
	precedenceAdditive
	precedenceMultiplicative
	precedenceValue
)

type renderer interface {
	Render(w *strings.Builder, d dialect.Renderer)
}

type precedencer interface {
	precedence() int
}

func renderChild(r renderer, parentPrecedence int, parenthesizeOnEqual bool, w *strings.Builder, d dialect.Renderer) {
	prec := precedenceValue
	if p, ok := r.(precedencer); ok {
		prec = p.precedence()
	}
	if prec < parentPrecedence || prec == parentPrecedence && parenthesizeOnEqual {
		w.WriteString("(")
		r.Render(w, d)
		w.WriteString(")")
		return
	}
	r.Render(w, d)
}

func infix(w *strings.Builder, d dialect.Renderer, a renderer, op string, b renderer, prec int, rightParenEqual bool) {
	renderChild(a, prec, false, w, d)
	w.WriteString(" ")
	w.WriteString(op)
	w.WriteString(" ")
	renderChild(b, prec, rightParenEqual, w, d)
}

// Expression represents a SQL value expression inside a DDL predicate.
type Expression struct {
	render func(w *strings.Builder, d dialect.Renderer)
	prec   int
}

func expression(prec int, render func(w *strings.Builder, d dialect.Renderer)) Expression {
	return Expression{render: render, prec: prec}
}

func asExpression(v any) renderer {
	switch v := v.(type) {
	case Expression:
		return v
	case renderer:
		return v
	default:
		return Literal(v)
	}
}

// Render writes the SQL representation of the expression.
func (e Expression) Render(w *strings.Builder, d dialect.Renderer) {
	e.render(w, d)
}

func (e Expression) precedence() int {
	return e.prec
}

func (e Expression) binary(op string, v any, prec int, rightParenEqual bool) Expression {
	r := asExpression(v)
	return expression(prec, func(w *strings.Builder, d dialect.Renderer) {
		infix(w, d, e, op, r, prec, rightParenEqual)
	})
}

// Add returns an addition expression.
func (e Expression) Add(v any) Expression { return e.binary("+", v, precedenceAdditive, false) }

// Sub returns a subtraction expression.
func (e Expression) Sub(v any) Expression { return e.binary("-", v, precedenceAdditive, true) }

// Mul returns a multiplication expression.
func (e Expression) Mul(v any) Expression { return e.binary("*", v, precedenceMultiplicative, false) }

// Div returns a division expression.
func (e Expression) Div(v any) Expression { return e.binary("/", v, precedenceMultiplicative, true) }

// Literal returns an expression rendered inline using the dialect's literal rules.
func Literal(v any) Expression {
	return expression(precedenceValue, func(w *strings.Builder, d dialect.Renderer) {
		w.WriteString(d.Literal(v))
	})
}

// Func builds a SQL function call expression.
func Func(name string, args ...any) Expression {
	renderers := make([]renderer, len(args))
	for i, arg := range args {
		renderers[i] = asExpression(arg)
	}
	return expression(precedenceValue, func(w *strings.Builder, d dialect.Renderer) {
		w.WriteString(name)
		w.WriteString("(")
		for i, arg := range renderers {
			if i > 0 {
				w.WriteString(", ")
			}
			arg.Render(w, d)
		}
		w.WriteString(")")
	})
}

// Col creates an unqualified column reference for DDL predicates.
func Col(name string) Expression {
	return expression(precedenceValue, func(w *strings.Builder, d dialect.Renderer) {
		w.WriteString(d.QuoteIdent(name))
	})
}

// Predicate represents a SQL boolean predicate inside DDL.
type Predicate struct {
	render func(w *strings.Builder, d dialect.Renderer)
	prec   int
}

// Predicater is implemented by SQL boolean expressions used in DDL predicates.
type Predicater interface {
	Render(w *strings.Builder, d dialect.Renderer)
	Predicate()
}

func predicate(prec int, render func(w *strings.Builder, d dialect.Renderer)) Predicate {
	return Predicate{render: render, prec: prec}
}

// Predicate marks the value as a predicate.
func (p Predicate) Predicate() {}

// Render writes the SQL representation of the predicate.
func (p Predicate) Render(w *strings.Builder, d dialect.Renderer) {
	p.render(w, d)
}

func (p Predicate) precedence() int {
	return p.prec
}

// And combines predicates with SQL AND.
func And(ps ...Predicater) Predicate { return logical("AND", precedenceAnd, ps) }

// Or combines predicates with SQL OR.
func Or(ps ...Predicater) Predicate { return logical("OR", precedenceOr, ps) }

func logical(op string, prec int, ps []Predicater) Predicate {
	return predicate(prec, func(w *strings.Builder, d dialect.Renderer) {
		for i, p := range ps {
			if i > 0 {
				w.WriteString(" ")
				w.WriteString(op)
				w.WriteString(" ")
			}
			renderChild(p, prec, false, w, d)
		}
	})
}

func (e Expression) compare(op string, v any) Predicate {
	r := asExpression(v)
	return predicate(precedenceComparison, func(w *strings.Builder, d dialect.Renderer) {
		infix(w, d, e, op, r, precedenceComparison, false)
	})
}

// Lt returns a less-than predicate.
func (e Expression) Lt(v any) Predicate { return e.compare("<", v) }

// Gt returns a greater-than predicate.
func (e Expression) Gt(v any) Predicate { return e.compare(">", v) }

// Le returns a less-than-or-equal predicate.
func (e Expression) Le(v any) Predicate { return e.compare("<=", v) }

// Ge returns a greater-than-or-equal predicate.
func (e Expression) Ge(v any) Predicate { return e.compare(">=", v) }

// Eq returns an equality predicate.
func (e Expression) Eq(v any) Predicate { return e.compare("=", v) }

// Like returns a LIKE predicate.
func (e Expression) Like(v any) Predicate { return e.compare("LIKE", v) }

// NotLike returns a NOT LIKE predicate.
func (e Expression) NotLike(v any) Predicate { return e.compare("NOT LIKE", v) }

// IsNull returns an IS NULL predicate.
func (e Expression) IsNull() Predicate { return e.compare("IS", Literal(nil)) }

// IsNotNull returns an IS NOT NULL predicate.
func (e Expression) IsNotNull() Predicate { return e.compare("IS NOT", Literal(nil)) }

func RawExpr(sql string) Expression {
	return expression(precedenceValue, func(w *strings.Builder, _ dialect.Renderer) {
		w.WriteString(sql)
	})
}

func RawPred(sql string) Predicate {
	return predicate(precedenceValue, func(w *strings.Builder, _ dialect.Renderer) {
		w.WriteString(sql)
	})
}
