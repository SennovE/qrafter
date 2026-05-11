package core

import "github.com/SennovE/qrafter/dialect"

const (
	PrecedenceOr = iota + 1
	PrecedenceAnd
	PrecedenceComparison
	PrecedenceAdditive
	PrecedenceMultiplicative
)

type Renderer interface {
	Render(d dialect.DialectRenderer) string
}

type Precedencer interface {
	Precedence() int
}

type Selecter interface {
	Renderer
	Tables() TablesSet
}

type Predicater interface {
	Selecter
	Predicate()
}

func RenderChild(r Renderer, parentPrecedence int, parenthesizeOnEqual bool, d dialect.DialectRenderer) string {
	rendered := r.Render(d)

	child, ok := r.(Precedencer)
	if !ok {
		return rendered
	}

	childPrecedence := child.Precedence()
	if childPrecedence < parentPrecedence || childPrecedence == parentPrecedence && parenthesizeOnEqual {
		return "(" + rendered + ")"
	}

	return rendered
}
