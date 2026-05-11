package pred

import (
	"fmt"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type BinaryPredicate struct {
	a, b core.Selecter
	op   string
}

var _ = (core.Predicater)(BinaryPredicate{})

func (e BinaryPredicate) Predicate() {}

func (e BinaryPredicate) Render(d dialect.DialectRenderer) string {
	return fmt.Sprintf("%s %s %s", e.a.Render(d), e.op, e.b.Render(d))
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
