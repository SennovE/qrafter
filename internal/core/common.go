package core

import (
	"github.com/SennovE/qrafter/dialect"
)

const (
	PrecedenceOr = iota + 1
	PrecedenceAnd
	PrecedenceComparison
	PrecedenceAdditive
	PrecedenceMultiplicative
)

type QueryRenderer interface {
	Render(d dialect.Renderer) (sql string, args []any, err error)
	MustRender(d dialect.Renderer) (sql string, args []any)
}

type QueryExpression interface {
	CTEs() []*CTERef
}

type Precedencer interface {
	Precedence() int
}

type Selecter interface {
	Tables() TablesSet
}

type Aggregater interface {
	Selecter
	Aggregate()
}

type Predicater interface {
	Selecter
	Predicate()
}
