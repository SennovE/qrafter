package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type WhereClause struct {
	Predicates []core.Predicater
}

var _ = (Clauser)(WhereClause{})

func (c WhereClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	if len(c.Predicates) > 0 {
		w.WriteString(" WHERE ")
		for i, pred := range c.Predicates {
			if i > 0 {
				w.WriteString(" AND ")
			}
			w.WriteString(pred.Render(d))
		}
	}
}
