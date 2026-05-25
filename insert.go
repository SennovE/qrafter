package qrafter

import (
	"reflect"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

// InsertQuery represents an INSERT statement under construction.
type InsertQuery struct {
	state *insertQueryState
}

var _ core.QueryRenderer = InsertQuery{}

type insertQueryState struct {
	table         core.TableRef
	columns       []ColumnRef
	rows          [][]core.Selecter
	source        core.QueryExpression
	defaultValues bool
	returning     []core.Selecter
}

type columnValueProvider interface {
	ColumnRef
	insertValue() any
}

type reflectedColumnValue struct {
	column ColumnRef
	value  any
}

type columnValueKey struct {
	table  core.TableRef
	column string
}

// Insert starts an INSERT query for the given table.
func Insert(table TableConfigProvider) InsertQuery {
	return InsertQuery{
		state: &insertQueryState{
			table: GetTableRef(table),
		},
	}
}

// Columns appends the target column list for the INSERT.
func (q InsertQuery) Columns(columns ...ColumnRef) InsertQuery {
	q = q.cloneState()
	q.state.columns = append(q.state.columns, columns...)
	return q
}

// Values appends a VALUES row. Plain Go values are rendered as bound arguments;
// use Literal for inline SQL literals and Default for the SQL DEFAULT keyword.
func (q InsertQuery) Values(values ...any) InsertQuery {
	q = q.cloneState()
	q.state.defaultValues = false
	q.state.source = nil
	q.state.rows = append(q.state.rows, asSelecters(values))
	return q
}

// ValuesRows appends multiple VALUES rows. Plain Go values are rendered as
// bound arguments; use Literal for inline SQL literals and Default for the SQL
// DEFAULT keyword.
func (q InsertQuery) ValuesRows(rows [][]any) InsertQuery {
	if len(rows) == 0 {
		return q
	}

	q = q.cloneState()
	q.state.defaultValues = false
	q.state.source = nil
	for _, row := range rows {
		q.state.rows = append(q.state.rows, asSelecters(row))
	}
	return q
}

// Set appends a target column and value to the current VALUES row.
func (q InsertQuery) Set(column ColumnRef, value any) InsertQuery {
	q = q.cloneState()
	q.state.defaultValues = false
	q.state.source = nil
	q.state.columns = append(q.state.columns, column)
	if len(q.state.rows) == 0 {
		q.state.rows = [][]core.Selecter{{}}
	}
	last := len(q.state.rows) - 1
	q.state.rows[last] = append(q.state.rows[last], asSelecter(value))
	return q
}

// ValuesFrom appends a VALUES row from the current values stored in Column
// fields on a table model. If Columns was already called, values are read in
// that column order.
func (q InsertQuery) ValuesFrom(row any) InsertQuery {
	values := reflectColumnValues(row)
	if len(values) == 0 {
		return q
	}

	q = q.cloneState()
	q.state.defaultValues = false
	q.state.source = nil

	if len(q.state.columns) == 0 {
		for _, value := range values {
			q.state.columns = append(q.state.columns, value.column)
		}
	}

	q.state.rows = append(q.state.rows, selectColumnValues(q.state.columns, values))
	return q
}

// ValuesRowsFrom appends VALUES rows from the current values stored in Column
// fields on a slice of table models. If Columns was already called, values are
// read in that column order.
func (q InsertQuery) ValuesRowsFrom(rows any) InsertQuery {
	valueRows := reflectColumnValueRows(rows)
	if len(valueRows) == 0 {
		return q
	}

	q = q.cloneState()
	q.state.defaultValues = false
	q.state.source = nil

	if len(q.state.columns) == 0 {
		for _, value := range valueRows[0] {
			q.state.columns = append(q.state.columns, value.column)
		}
	}

	for _, values := range valueRows {
		q.state.rows = append(q.state.rows, selectColumnValues(q.state.columns, values))
	}
	return q
}

// DefaultValues makes the INSERT use DEFAULT VALUES.
func (q InsertQuery) DefaultValues() InsertQuery {
	q = q.cloneState()
	q.state.columns = nil
	q.state.rows = nil
	q.state.source = nil
	q.state.defaultValues = true
	return q
}

// FromSelect makes the INSERT read rows from a SELECT or compound query.
func (q InsertQuery) FromSelect(source core.QueryExpression) InsertQuery {
	q = q.cloneState()
	q.state.rows = nil
	q.state.defaultValues = false
	q.state.source = source
	return q
}

// Returning appends expressions to a RETURNING clause.
func (q InsertQuery) Returning(items ...core.Selecter) InsertQuery {
	q = q.cloneState()
	q.state.returning = append(q.state.returning, items...)
	return q
}

// Render renders the query and returns SQL, bound arguments and an error if the query is invalid.
func (q InsertQuery) Render(d dialect.Renderer) (sql string, args []any, err error) {
	return renderQuery(func() (string, []any) {
		return q.MustRender(d)
	})
}

// MustRender is like Render but panics if the query is invalid.
func (q InsertQuery) MustRender(d dialect.Renderer) (sql string, args []any) {
	return renderStatement(d, q.CTEs(), q.RenderStatement)
}

// RenderStatement writes the INSERT query body.
func (q InsertQuery) RenderStatement(w *strings.Builder, d dialect.Renderer) {
	state := q.currentState()

	w.WriteString("INSERT INTO ")
	state.table.Render(w, d)
	renderInsertColumns(w, d, state.columns)

	switch {
	case state.defaultValues || state.source == nil && len(state.rows) == 0:
		dialect.RenderDefaultValues(w, d)
	case state.source != nil:
		w.WriteString("\n")
		state.source.RenderQueryExpression(w, d)
	default:
		w.WriteString("\nVALUES ")
		renderInsertRows(w, d, state.rows)
	}

	renderReturning(w, d, state.returning)
}

// CTEs returns common table expressions referenced by the INSERT source.
func (q InsertQuery) CTEs() []*core.CTERef {
	state := q.currentState()
	if state.source == nil {
		return nil
	}
	return state.source.CTEs()
}

func (q InsertQuery) currentState() insertQueryState {
	if q.state == nil {
		return insertQueryState{}
	}
	return *q.state
}

func (q InsertQuery) cloneState() InsertQuery {
	state := q.currentState()
	state.columns = append([]ColumnRef(nil), state.columns...)
	state.rows = cloneInsertRows(state.rows)
	state.returning = append([]core.Selecter(nil), state.returning...)
	q.state = &state
	return q
}

func cloneInsertRows(rows [][]core.Selecter) [][]core.Selecter {
	if len(rows) == 0 {
		return nil
	}
	cloned := make([][]core.Selecter, len(rows))
	for i, row := range rows {
		cloned[i] = append([]core.Selecter(nil), row...)
	}
	return cloned
}

func renderInsertColumns(w *strings.Builder, d dialect.Renderer, columns []ColumnRef) {
	if len(columns) == 0 {
		return
	}

	w.WriteString(" (")
	for i, column := range columns {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString(d.QuoteIdent(column.ColumnName()))
	}
	w.WriteString(")")
}

func renderInsertRows(w *strings.Builder, d dialect.Renderer, rows [][]core.Selecter) {
	for i, row := range rows {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString("(")
		core.RenderWithDelimiter(w, d, ", ", row)
		w.WriteString(")")
	}
}

func reflectColumnValueRows(rows any) [][]reflectedColumnValue {
	v := reflect.ValueOf(rows)
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		if elemKind := v.Elem().Kind(); elemKind == reflect.Slice || elemKind == reflect.Array {
			v = v.Elem()
		}
	}

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil
	}

	valueRows := make([][]reflectedColumnValue, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !elem.CanInterface() {
			continue
		}

		values := reflectColumnValues(elem.Interface())
		if len(values) == 0 {
			continue
		}
		valueRows = append(valueRows, values)
	}

	return valueRows
}

func reflectColumnValues(row any) []reflectedColumnValue {
	v := reflect.ValueOf(row)
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	values := make([]reflectedColumnValue, 0, v.NumField())
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		f := v.Field(i)
		if !f.CanInterface() {
			continue
		}

		provider, ok := f.Interface().(columnValueProvider)
		if !ok {
			continue
		}

		values = append(values, reflectedColumnValue{
			column: provider,
			value:  provider.insertValue(),
		})
	}

	return values
}

func selectColumnValues(columns []ColumnRef, values []reflectedColumnValue) []core.Selecter {
	byKey := make(map[columnValueKey]any, len(values))
	byName := make(map[string]any, len(values))
	for _, value := range values {
		key := columnKey(value.column)
		byKey[key] = value.value
		byName[key.column] = value.value
	}

	row := make([]core.Selecter, 0, len(columns))
	for _, column := range columns {
		if value, ok := byKey[columnKey(column)]; ok {
			row = append(row, asSelecter(value))
			continue
		}
		row = append(row, asSelecter(byName[column.ColumnName()]))
	}

	return row
}

func columnKey(column ColumnRef) columnValueKey {
	return columnValueKey{
		table:  column.TableRef(),
		column: column.ColumnName(),
	}
}
