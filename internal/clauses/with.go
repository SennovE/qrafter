package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type WithClause struct {
	Recursive bool
	CTEs      []*core.CTERef
}

var _ = (Clauser)(WithClause{})

func (c WithClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	if len(c.CTEs) == 0 {
		return
	}

	w.WriteString("WITH ")
	if c.Recursive {
		w.WriteString("RECURSIVE ")
	}

	core.RenderWithDelimiter(w, d, ", ", c.CTEs)
	w.WriteString(" ")
}
