package pred

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type BinaryPredicate struct {
	a, b core.Selecter
	op   string
}

var _ core.Predicater = BinaryPredicate{}

func (e BinaryPredicate) Predicate() {}

func (e BinaryPredicate) Left() core.Selecter {
	return e.a
}

func (e BinaryPredicate) Right() core.Selecter {
	return e.b
}

func (e BinaryPredicate) Op() string {
	return e.op
}

func (e BinaryPredicate) Precedence() int {
	return core.PrecedenceComparison
}

func (e BinaryPredicate) Tables() core.TablesSet {
	return utils.UnionSets(e.a.Tables(), e.b.Tables())
}

func Binary(op string, a, b core.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: op,
	}
}
