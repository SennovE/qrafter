package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/pred"
)

type WhereClause struct {
	Predicates []core.Predicater
}

var _ = (Clauser)(WhereClause{})

func (c WhereClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	if len(c.Predicates) > 0 {
		w.WriteString(" WHERE ")
		if len(c.Predicates) == 1 {
			w.WriteString(c.Predicates[0].Render(d))
			return
		}
		w.WriteString(pred.Logical(pred.OpAnd, c.Predicates...).Render(d))
	}
}
