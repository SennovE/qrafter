package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type OrderByClause struct {
	Items []core.Selecter
}

var _ Clauser = OrderByClause{}

func (c OrderByClause) Render(w *strings.Builder, d dialect.Renderer) {
	if len(c.Items) == 0 {
		return
	}

	w.WriteString(" ORDER BY ")
	core.RenderWithDelimiter(w, d, ", ", c.Items)
}
