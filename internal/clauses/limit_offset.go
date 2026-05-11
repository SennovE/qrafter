package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type LimitOffsetClause struct {
	Limit, Offset int
}

var _ = (Clauser)(LimitOffsetClause{})

func (c LimitOffsetClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString(d.LimitOffset(c.Limit, c.Offset))
}
