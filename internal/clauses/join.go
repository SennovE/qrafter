package clauses

import (
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

var _ Clauser = (*JoinClause)(nil)

func (c *JoinClause) Render(w *strings.Builder, d dialect.Renderer) {
	dialect.RenderJoin(w, d, c.Type, func() {
		c.Table.Render(w, d)
	}, func() {
		renderJoinPredicates(w, d, c.Predicates)
	})
}

func renderJoinPredicates(w *strings.Builder, d dialect.Renderer, predicates []core.Predicater) {
	if len(predicates) == 0 {
		return
	}

	w.WriteString(" ON ")
	if len(predicates) == 1 {
		predicates[0].Render(w, d)
		return
	}
	pred.Logical(pred.OpAnd, predicates...).Render(w, d)
}
