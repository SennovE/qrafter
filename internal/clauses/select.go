package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type SelectClause struct {
	Colums []core.Selecter
}

var _ = (Clauser)(SelectClause{})

func (c SelectClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString("SELECT ")

	for i, col := range c.Colums {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString(col.Render(d))
	}
}
