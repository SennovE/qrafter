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

func (c WithClause) WithClauseFor(q core.QueryExpression) WithClause {
	seen := make(map[string]struct{}, len(c.CTEs))

	for _, cte := range c.CTEs {
		if cte == nil {
			continue
		}
		seen[cte.Name] = struct{}{}
		if cte.Recursive {
			c.Recursive = true
		}
	}

	for _, cte := range q.CTEs() {
		if cte == nil {
			continue
		}
		if cte.Recursive {
			c.Recursive = true
		}
		if _, ok := seen[cte.Name]; !ok {
			c.CTEs = append(c.CTEs, cte)
			seen[cte.Name] = struct{}{}
			for _, cte := range cte.Query.CTEs() {
				if cte == nil {
					continue
				}
				if cte.Recursive {
					c.Recursive = true
				}
				if _, ok := seen[cte.Name]; !ok {
					c.CTEs = append(c.CTEs, cte)
					seen[cte.Name] = struct{}{}
				}
			}
		}
	}

	return c
}
