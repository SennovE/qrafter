package core

import (
	"sort"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type ColumnBinder interface {
	Bind(name string, table TableRef)
}

type TableRef struct {
	Name  string
	Alias string
	CTE   *CTERef
}

var _ Renderer = TableRef{}

type TablesSet = map[TableRef]struct{}

type CTERef struct {
	Name      string
	Columns   []string
	Query     QueryExpression
	Recursive bool
}

var _ Renderer = (*CTERef)(nil)

func (t TableRef) SQLName() string {
	if t.Alias == "" {
		return t.Name
	}
	return t.Alias
}

func (t TableRef) Render(w *strings.Builder, d dialect.Renderer) {
	if t.Alias == "" {
		w.WriteString(d.QuoteIdent(t.Name))
	} else {
		w.WriteString(d.QuoteIdent(t.Name))
		w.WriteString(" AS ")
		w.WriteString(d.QuoteIdent(t.Alias))
	}
}

func (cte *CTERef) Render(w *strings.Builder, d dialect.Renderer) {
	if cte == nil {
		return
	}

	w.WriteString(d.QuoteIdent(cte.Name))
	if len(cte.Columns) > 0 {
		w.WriteString(" (")
		for i, column := range cte.Columns {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteString(d.QuoteIdent(column))
		}
		w.WriteString(")")
	}

	var body strings.Builder
	cte.Query.RenderQueryExpression(&body, d)

	w.WriteString(" AS (\n")
	writeIndentedLines(w, body.String(), "    ")
	w.WriteString("\n)")
}

func writeIndentedLines(w *strings.Builder, s, indent string) {
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			w.WriteString("\n")
		}
		if line == "" {
			continue
		}
		w.WriteString(indent)
		w.WriteString(line)
	}
}

func GetSortedTables(tables TablesSet) []TableRef {
	sortedTables := make([]TableRef, 0, len(tables))
	for table := range tables {
		sortedTables = append(sortedTables, table)
	}
	sort.Slice(sortedTables, func(i, j int) bool {
		if sortedTables[i].Name == sortedTables[j].Name {
			return sortedTables[i].Alias < sortedTables[j].Alias
		}
		return sortedTables[i].Name < sortedTables[j].Name
	})
	return sortedTables
}
