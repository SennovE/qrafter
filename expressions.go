package qrafter

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
)

func As(expression core.Selecter, alias string) expr.AliasedExpression {
	return expr.Alias(expression, alias)
}

func Const(v any) expr.ConstExpression {
	return expr.Const(v)
}

func Sum(a, b core.Selecter) expr.BinaryExpression {
	return expr.Binary("+", a, b)
}
