package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/pred"
)

type HavingClause struct {
	Predicates []core.Predicater
}

var _ = (Clauser)(HavingClause{})

func (c HavingClause) Render(w *strings.Builder, d dialect.Renderer) {
	if len(c.Predicates) == 0 {
		return
	}

	w.WriteString(" HAVING ")
	if len(c.Predicates) == 1 {
		c.Predicates[0].Render(w, d)
		return
	}
	pred.Logical(pred.OpAnd, c.Predicates...).Render(w, d)
}
