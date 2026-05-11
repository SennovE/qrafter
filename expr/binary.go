package expr

import (
	"fmt"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/utils"
)

type BinaryExpression struct {
	a, b Selecter
	op   string
}

var _ = (Selecter)(BinaryExpression{})

func (e BinaryExpression) Render() string {
	return fmt.Sprintf("%s %s %s", e.a.Render(), e.op, e.b.Render())
}

func (e BinaryExpression) Tables() qrafter.TablesSet {
	return utils.UnionSets(e.a.Tables(), e.b.Tables())
}

func Sum(a, b Selecter) BinaryExpression {
	return BinaryExpression{
		a:  a,
		b:  b,
		op: "+",
	}
}
