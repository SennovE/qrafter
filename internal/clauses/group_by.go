package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type GroupByClause struct {
	Columns []core.Selecter
}

var _ Clauser = GroupByClause{}

func (c GroupByClause) Render(w *strings.Builder, d dialect.Renderer) {
	if len(c.Columns) == 0 {
		return
	}

	w.WriteString(" GROUP BY ")
	core.RenderWithDelimiter(w, d, ", ", c.Columns)
}
