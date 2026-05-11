package qrafter

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
)

func As(e core.Selecter, alias string) expr.AliasedExpression {
	return expr.Alias(e, alias)
}

func Const(v any) expr.ConstExpression {
	return expr.Const(v)
}

func Sum(a, b core.Selecter) expr.BinaryExpression {
	return expr.Binary("+", a, b)
}

func Sub(a, b core.Selecter) expr.BinaryExpression {
	return expr.Binary("-", a, b)
}

func Mul(a, b core.Selecter) expr.BinaryExpression {
	return expr.Binary("*", a, b)
}

func Div(a, b core.Selecter) expr.BinaryExpression {
	return expr.Binary("/", a, b)
}
