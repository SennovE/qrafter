package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
)

// UpdateQuery represents an UPDATE statement under construction.
type UpdateQuery struct {
	state *updateQueryState
}

type updateQueryState struct {
	table       core.TableRef
	assignments []updateAssignment
	from        []core.TableRef
	whereCl     clauses.WhereClause
	returning   []core.Selecter
}

type updateAssignment struct {
	column ColumnRef
	value  core.Selecter
}

// Update starts an UPDATE query for the given table.
func Update(table TableConfigProvider) UpdateQuery {
	return UpdateQuery{
		state: &updateQueryState{
			table: GetTableRef(table),
		},
	}
}

// Set appends a column assignment to the UPDATE SET clause. Plain Go values are
// rendered as bound arguments; use Literal for inline SQL literals and Default
// for the SQL DEFAULT keyword.
func (q UpdateQuery) Set(column ColumnRef, value any) UpdateQuery {
	q = q.cloneState()
	selecter := asSelecter(value)
	q.state.assignments = append(q.state.assignments, updateAssignment{
		column: column,
		value:  selecter,
	})
	q.state.addFromTables(sortedTablesFromSelecters([]core.Selecter{selecter}))
	return q
}

// SetFrom appends assignments from the current values stored in Column fields
// on a table model.
func (q UpdateQuery) SetFrom(row any) UpdateQuery {
	values := reflectColumnValues(row)
	if len(values) == 0 {
		return q
	}

	q = q.cloneState()
	for _, value := range values {
		selecter := asSelecter(value.value)
		q.state.assignments = append(q.state.assignments, updateAssignment{
			column: value.column,
			value:  selecter,
		})
		q.state.addFromTables(sortedTablesFromSelecters([]core.Selecter{selecter}))
	}
	return q
}

// From appends tables to the UPDATE FROM clause.
func (q UpdateQuery) From(tables ...TableConfigProvider) UpdateQuery {
	q = q.cloneState()
	for _, table := range tables {
		q.state.addFrom(GetTableRef(table))
	}
	return q
}

// Where appends predicates to the UPDATE WHERE clause.
func (q UpdateQuery) Where(predicates ...core.Predicater) UpdateQuery {
	q = q.cloneState()
	q.state.addFromTables(sortedTablesFromSelecters(predicates))
	q.state.whereCl.Predicates = append(q.state.whereCl.Predicates, predicates...)
	return q
}

// Returning appends expressions to a RETURNING clause.
func (q UpdateQuery) Returning(items ...core.Selecter) UpdateQuery {
	q = q.cloneState()
	q.state.returning = append(q.state.returning, items...)
	return q
}

// Render renders the query and returns SQL plus bound arguments.
func (q UpdateQuery) Render(d dialect.Renderer) (sql string, args []any) {
	return renderStatement(d, q.CTEs(), q.RenderStatement)
}

// RenderStatement writes the UPDATE query body.
func (q UpdateQuery) RenderStatement(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()

	w.WriteString("UPDATE ")
	state.table.Render(w, d)
	renderUpdateAssignments(w, d, state.assignments)
	renderUpdateFrom(w, d, state.from)
	state.whereCl.Render(w, d)
	renderReturning(w, d, state.returning)
}

// CTEs returns common table expressions referenced by the UPDATE query.
func (q UpdateQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	seen := make(map[string]struct{})
	ctes := make([]*core.CTERef, 0)

	ctes = appendCTEFromTable(ctes, seen, state.table)
	ctes = appendCTEsFromUpdateAssignments(ctes, seen, state.assignments)
	ctes = appendCTEsFromTables(ctes, seen, state.from)
	ctes = appendCTEsFromSelecters(ctes, seen, state.whereCl.Predicates)
	ctes = appendCTEsFromSelecters(ctes, seen, state.returning)

	return ctes
}

func (q UpdateQuery) currentState() updateQueryState {
	if q.state == nil {
		return updateQueryState{}
	}
	return *q.state
}

func (q UpdateQuery) cloneState() UpdateQuery {
	state := q.currentState()
	state.assignments = append([]updateAssignment(nil), state.assignments...)
	state.from = append([]core.TableRef(nil), state.from...)
	state.whereCl.Predicates = append([]core.Predicater(nil), state.whereCl.Predicates...)
	state.returning = append([]core.Selecter(nil), state.returning...)
	q.state = &state
	return q
}

func (s *updateQueryState) addFrom(table core.TableRef) {
	if table.Name == "" || table == s.table {
		return
	}
	for _, existing := range s.from {
		if existing == table {
			return
		}
	}
	s.from = append(s.from, table)
}

func (s *updateQueryState) addFromTables(tables []core.TableRef) {
	for _, table := range tables {
		s.addFrom(table)
	}
}

func renderUpdateAssignments(w *strings.Builder, d dialect.Renderer, assignments []updateAssignment) {
	w.WriteString(" SET ")
	for i, assignment := range assignments {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString(d.QuoteIdent(assignment.column.ColumnName()))
		w.WriteString(" = ")
		assignment.value.Render(w, d)
	}
}

func renderUpdateFrom(w *strings.Builder, d dialect.Renderer, from []core.TableRef) {
	if len(from) == 0 {
		return
	}

	w.WriteString(" FROM ")
	core.RenderWithDelimiter(w, d, ", ", from)
}

func appendCTEsFromUpdateAssignments(ctes []*core.CTERef, seen map[string]struct{}, assignments []updateAssignment) []*core.CTERef {
	for _, assignment := range assignments {
		ctes = appendCTEsFromTables(ctes, seen, sortedTablesFromSelecters([]core.Selecter{assignment.value}))
	}
	return ctes
}
