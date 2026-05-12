package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

type withClause struct {
	recursive bool
	ctes      []CommonTableExpression
}

var _ = (clauses.Clauser)(withClause{})

func (c withClause) Render(w *strings.Builder, d dialect.DialectRenderer) {
	if len(c.ctes) == 0 {
		return
	}

	w.WriteString("WITH ")
	if c.recursive {
		w.WriteString("RECURSIVE ")
	}

	core.RenderWithDelimiter(w, d, ", ", c.ctes)
	w.WriteString(" ")
}

type CommonTableExpression struct {
	name    string
	columns []string
	query   core.QueryRenderer
}

func (cte CommonTableExpression) TableConfig() TableConfig {
	return TableConfig{Name: cte.name}
}

func (cte CommonTableExpression) WithColumns(columns ...string) CommonTableExpression {
	cte.columns = append(cte.columns, columns...)
	return cte
}

func (cte CommonTableExpression) Bind(table any) error {
	return bindWithTableRef(table, core.TableRef{Name: cte.name})
}

func (cte CommonTableExpression) Column(name string) Column[any] {
	var col Column[any]
	col.Bind(name, core.TableRef{Name: cte.name})
	return col
}

func (cte CommonTableExpression) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString(d.QuoteIdent(cte.name))
	if len(cte.columns) > 0 {
		w.WriteString(" (")
		for i, column := range cte.columns {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteString(d.QuoteIdent(column))
		}
		w.WriteString(")")
	}

	w.WriteString(" AS (")
	w.WriteString(cte.query.Render(d))
	w.WriteString(")")
}
