package clauses

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/pred"
)

type JoinClause struct {
	Type       string
	Table      core.TableRef
	Predicates []core.Predicater
}

var _ = (Clauser)(JoinClause{})

func (j JoinClause) Render(w *strings.Builder, d dialect.Renderer) {
	fmt.Fprintf(w, " %s ", j.Type)
	j.Table.Render(w, d)

	if len(j.Predicates) == 0 {
		return
	}

	w.WriteString(" ON ")
	if len(j.Predicates) == 1 {
		j.Predicates[0].Render(w, d)
		return
	}
	pred.Logical(pred.OpAnd, j.Predicates...).Render(w, d)
}
