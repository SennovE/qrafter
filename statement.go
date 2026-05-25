package qrafter

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

type cteCollector struct {
	ctes []*core.CTERef
}

type statementBodyRenderer func(w *strings.Builder, d dialect.Renderer)

type queryRenderer func() (sql string, args []any)

func renderQuery(fn queryRenderer) (sql string, args []any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			sql = ""
			args = nil
			err = panicToError(recovered)
		}
	}()

	sql, args = fn()
	return sql, args, nil
}

func panicToError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}
	return fmt.Errorf("%v", value)
}

func renderStatement(d dialect.Renderer, ctes []*core.CTERef, renderBody statementBodyRenderer) (sql string, args []any) {
	return renderStatementWithClause(d, clauses.WithClause{}, ctes, renderBody)
}

func renderStatementWithClause(
	d dialect.Renderer,
	withCl clauses.WithClause,
	ctes []*core.CTERef,
	renderBody statementBodyRenderer,
) (sql string, args []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	withCl = withCl.WithClauseFor(cteCollector{ctes: ctes})
	withCl.Render(&w, renderer)
	renderBody(&w, renderer)

	return w.String(), renderer.Args()
}

func renderReturning(w *strings.Builder, d dialect.Renderer, returning []core.Selecter) {
	if len(returning) == 0 {
		return
	}

	dialect.RenderReturning(w, d, func() {
		core.RenderWithDelimiter(w, d, ", ", returning)
	})
}

func (c cteCollector) RenderQueryExpression(_ *strings.Builder, _ dialect.Renderer) {}

func (c cteCollector) RenderSetOperand(_ *strings.Builder, _ dialect.Renderer) {}

func (c cteCollector) CTEs() []*core.CTERef {
	return c.ctes
}

func sortedTablesFromSelecters[T core.Selecter](items []T) []core.TableRef {
	tables := make(core.TablesSet)
	for _, item := range items {
		for table := range item.Tables() {
			tables[table] = struct{}{}
		}
	}
	return core.GetSortedTables(tables)
}

func appendCTEsFromSelecters[T core.Selecter](ctes []*core.CTERef, seen map[string]struct{}, items []T) []*core.CTERef {
	for _, table := range sortedTablesFromSelecters(items) {
		ctes = appendCTEFromTable(ctes, seen, table)
	}
	return ctes
}

func appendCTEsFromTables(ctes []*core.CTERef, seen map[string]struct{}, tables []core.TableRef) []*core.CTERef {
	for _, table := range tables {
		ctes = appendCTEFromTable(ctes, seen, table)
	}
	return ctes
}

func appendCTEFromTable(ctes []*core.CTERef, seen map[string]struct{}, table core.TableRef) []*core.CTERef {
	if table.CTE == nil {
		return ctes
	}
	if _, ok := seen[table.CTE.Name]; ok {
		return ctes
	}
	ctes = append(ctes, table.CTE)
	seen[table.CTE.Name] = struct{}{}
	return ctes
}
