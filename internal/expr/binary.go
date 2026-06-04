package expr

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type BinaryExpression struct {
	a, b core.Selecter
	op   string
}

var _ core.Selecter = BinaryExpression{}

func (e BinaryExpression) Left() core.Selecter {
	return e.a
}

func (e BinaryExpression) Right() core.Selecter {
	return e.b
}

func (e BinaryExpression) Op() string {
	return e.op
}

func (e BinaryExpression) Precedence() int {
	switch e.op {
	case "*", "/", "%":
		return core.PrecedenceMultiplicative
	case "+", "-":
		return core.PrecedenceAdditive
	default:
		return core.PrecedenceComparison
	}
}

func (e BinaryExpression) Tables() core.TablesSet {
	return utils.UnionSets(e.a.Tables(), e.b.Tables())
}

func Binary(op string, a, b core.Selecter) BinaryExpression {
	return BinaryExpression{
		a:  a,
		b:  b,
		op: op,
	}
}
