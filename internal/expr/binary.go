package expr

import (
	"fmt"

	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type BinaryExpression struct {
	a, b core.Selecter
	op   string
}

var _ = (core.Selecter)(BinaryExpression{})

func (e BinaryExpression) Render() string {
	return fmt.Sprintf("%s %s %s", e.a.Render(), e.op, e.b.Render())
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
