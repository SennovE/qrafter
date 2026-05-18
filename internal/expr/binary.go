package expr

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type BinaryExpression struct {
	a, b core.Selecter
	op   string
}

var _ core.Selecter = BinaryExpression{}

func (e BinaryExpression) Render(w *strings.Builder, d dialect.Renderer) {
	core.RenderChild(e.a, e.Precedence(), false, w, d)
	fmt.Fprintf(w, " %s ", e.op)
	core.RenderChild(e.b, e.Precedence(), e.parenthesizeRightPeer(), w, d)
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

func (e BinaryExpression) parenthesizeRightPeer() bool {
	switch e.op {
	case "-", "/", "%":
		return true
	default:
		return false
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
