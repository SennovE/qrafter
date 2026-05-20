package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// DeleteQuery represents a DELETE statement under construction.
type DeleteQuery struct {
	state *deleteQueryState
}

type deleteQueryState struct {
	table     core.TableRef
	using     []core.TableRef
	whereCl   clauses.WhereClause
	returning []core.Selecter
}

// Delete starts a DELETE query for the given table.
func Delete(table TableConfigProvider) DeleteQuery {
	return DeleteFrom(table)
}

// DeleteFrom starts a DELETE query for the given table.
func DeleteFrom(table TableConfigProvider) DeleteQuery {
	return DeleteQuery{
		state: &deleteQueryState{
			table: GetTableRef(table),
		},
	}
}

// Using appends tables to the DELETE USING clause.
func (q DeleteQuery) Using(tables ...TableConfigProvider) DeleteQuery {
	q = q.cloneState()
	for _, table := range tables {
		q.state.addUsing(GetTableRef(table))
	}
	return q
}

// Where appends predicates to the DELETE WHERE clause.
func (q DeleteQuery) Where(predicates ...core.Predicater) DeleteQuery {
	q = q.cloneState()
	q.state.addUsingTablesFrom(predicates)
	q.state.whereCl.Predicates = append(q.state.whereCl.Predicates, predicates...)
	return q
}

// Returning appends expressions to a RETURNING clause.
func (q DeleteQuery) Returning(items ...core.Selecter) DeleteQuery {
	q = q.cloneState()
	q.state.returning = append(q.state.returning, items...)
	return q
}

// Render renders the query and returns SQL plus bound arguments.
func (q DeleteQuery) Render(d dialect.Renderer) (sql string, args []any) {
	renderer := core.NewArgsRenderer(d)
	var w strings.Builder

	withCl := clauses.WithClause{}.WithClauseFor(cteCollector{ctes: q.CTEs()})
	withCl.Render(&w, renderer)
	q.RenderStatement(&w, renderer)

	return w.String(), renderer.Args()
}

// RenderStatement writes the DELETE query body.
func (q DeleteQuery) RenderStatement(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()

	w.WriteString("DELETE FROM ")
	state.table.Render(w, d)
	renderDeleteUsing(w, d, state.using)
	state.whereCl.Render(w, d)
	renderReturning(w, d, state.returning)
}

// CTEs returns common table expressions referenced by the DELETE query.
func (q DeleteQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	seen := make(map[string]struct{})
	ctes := make([]*core.CTERef, 0)

	ctes = appendCTEFromTable(ctes, seen, state.table)
	ctes = appendCTEsFromTables(ctes, seen, state.using)
	ctes = appendCTEsFromSelecters(ctes, seen, state.whereCl.Predicates)
	ctes = appendCTEsFromSelecters(ctes, seen, state.returning)

	return ctes
}

func (q DeleteQuery) currentState() deleteQueryState {
	if q.state == nil {
		return deleteQueryState{}
	}
	return *q.state
}

func (q DeleteQuery) cloneState() DeleteQuery {
	state := q.currentState()
	state.using = append([]core.TableRef(nil), state.using...)
	state.whereCl.Predicates = append([]core.Predicater(nil), state.whereCl.Predicates...)
	state.returning = append([]core.Selecter(nil), state.returning...)
	q.state = &state
	return q
}

func (s *deleteQueryState) addUsing(table core.TableRef) {
	if table.Name == "" || table == s.table {
		return
	}
	for _, existing := range s.using {
		if existing == table {
			return
		}
	}
	s.using = append(s.using, table)
}

func (s *deleteQueryState) addUsingTablesFrom(items []core.Predicater) {
	for _, table := range sortedTablesFromSelecters(items) {
		s.addUsing(table)
	}
}

func renderDeleteUsing(w *strings.Builder, d dialect.Renderer, using []core.TableRef) {
	if len(using) == 0 {
		return
	}

	w.WriteString(" USING ")
	core.RenderWithDelimiter(w, d, ", ", using)
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
