package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type LimitOffsetClause struct {
	Limit, Offset int
}

var _ = (Clauser)(LimitOffsetClause{})

func (c LimitOffsetClause) Render(w *strings.Builder, d dialect.Renderer) {
	rendered := d.LimitOffset(c.Limit, c.Offset)
	if rendered == "" {
		return
	}
	w.WriteString(" ")
	w.WriteString(rendered)
}
