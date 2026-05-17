package pred

import (
	"fmt"
	"strings"

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

func (e BinaryPredicate) Render(w *strings.Builder, d dialect.Renderer) {
	core.RenderChild(e.a, e.Precedence(), false, w, d)
	fmt.Fprintf(w, " %s ", e.op)
	core.RenderChild(e.b, e.Precedence(), false, w, d)
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
