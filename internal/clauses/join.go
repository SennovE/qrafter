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

var _ Clauser = (*JoinClause)(nil)

func (c *JoinClause) Render(w *strings.Builder, d dialect.Renderer) {
	fmt.Fprintf(w, "\n%s ", c.Type)
	c.Table.Render(w, d)

	if len(c.Predicates) == 0 {
		return
	}

	w.WriteString(" ON ")
	if len(c.Predicates) == 1 {
		c.Predicates[0].Render(w, d)
		return
	}
	pred.Logical(pred.OpAnd, c.Predicates...).Render(w, d)
}
