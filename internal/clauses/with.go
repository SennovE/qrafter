package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type WithClause struct {
	Recursive bool
	CTEs      []*core.CTERef
}

func (c WithClause) WithClauseFor(q core.QueryExpression) WithClause {
	seen := make(map[string]struct{}, len(c.CTEs))
	c.indexExistingCTEs(seen)
	c.appendCTEs(q.CTEs(), seen)

	return c
}

func (c *WithClause) indexExistingCTEs(seen map[string]struct{}) {
	for _, cte := range c.CTEs {
		if cte == nil {
			continue
		}
		seen[cte.Name] = struct{}{}
		if cte.Recursive {
			c.Recursive = true
		}
	}
}

func (c *WithClause) appendCTEs(ctes []*core.CTERef, seen map[string]struct{}) {
	for _, cte := range ctes {
		if !c.appendCTE(cte, seen) {
			continue
		}
		c.appendCTEs(cte.Query.CTEs(), seen)
	}
}

func (c *WithClause) appendCTE(cte *core.CTERef, seen map[string]struct{}) bool {
	if cte == nil {
		return false
	}
	if cte.Recursive {
		c.Recursive = true
	}
	if _, ok := seen[cte.Name]; ok {
		return false
	}

	c.CTEs = append(c.CTEs, cte)
	seen[cte.Name] = struct{}{}
	return true
}
