package expr

import (
	"github.com/SennovE/qrafter/internal/core"
)

type ArgExpression struct {
	v any
}

var _ core.Selecter = ArgExpression{}

func (a ArgExpression) Tables() core.TablesSet {
	return nil
}

func (a ArgExpression) Value() any {
	return a.v
}

func Param(value any) ArgExpression {
	return ArgExpression{v: value}
}
