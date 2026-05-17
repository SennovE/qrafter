package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type SelectClause struct {
	Columns []core.Selecter
}

var _ = (Clauser)(SelectClause{})

func (c SelectClause) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("SELECT ")
	core.RenderWithDelimiter(w, d, ", ", c.Columns)
}
