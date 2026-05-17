package qrafter

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
)

// Aggregate represents a SQL aggregate expression such as COUNT or SUM.
type Aggregate struct {
	Expression
}

var _ = (core.Aggregater)(Aggregate{})

func newAggregate(s core.Selecter) Aggregate {
	return Aggregate{Expression: newExpression(s)}
}

// Aggregate marks the value as an aggregate expression.
func (a Aggregate) Aggregate() {}

// As returns the aggregate expression with a SQL alias.
func (a Aggregate) As(alias string) Aggregate {
	return newAggregate(expr.Alias(a.selecter, alias))
}

// AggregateFunc builds an aggregate expression with the given function name.
func AggregateFunc(name string, args ...any) Aggregate {
	return newAggregate(expr.Function(name, asSelecters(args)...))
}

// Count builds a COUNT aggregate expression.
func Count(args ...any) Aggregate {
	if len(args) == 0 {
		return AggregateFunc("COUNT", Star())
	}
	return AggregateFunc("COUNT", args...)
}

// Sum builds a SUM aggregate expression.
func Sum(v any) Aggregate {
	return AggregateFunc("SUM", v)
}

// Avg builds an AVG aggregate expression.
func Avg(v any) Aggregate {
	return AggregateFunc("AVG", v)
}

// Min builds a MIN aggregate expression.
func Min(v any) Aggregate {
	return AggregateFunc("MIN", v)
}

// Max builds a MAX aggregate expression.
func Max(v any) Aggregate {
	return AggregateFunc("MAX", v)
}
