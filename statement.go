package qrafter

import (
	"fmt"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

type cteCollector struct {
	ctes []*core.CTERef
}

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

func renderStatement(d dialect.Renderer, ctes []*core.CTERef, node any) (sql string, args []any) {
	return renderStatementWithClause(d, clauses.WithClause{}, ctes, node)
}

func renderStatementWithClause(
	d dialect.Renderer,
	withCl clauses.WithClause,
	ctes []*core.CTERef,
	node any,
) (sql string, args []any) {
	compiler := newCompiler(d)

	withCl = withCl.WithClauseFor(cteCollector{ctes: ctes})
	compiler.Compile(withCl)
	compiler.Compile(node)

	return compiler.SQL(), compiler.Args()
}

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
