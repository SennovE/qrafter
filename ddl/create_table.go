package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// CreateTableStmt builds a CREATE TABLE statement.
type CreateTableStmt struct {
	state *createTableState
}

type createTableState struct {
	table       string
	model       any
	ifNotExists bool
	columns     []ColumnDef
	constraints []Constraint
}

// CreateTable starts a CREATE TABLE statement.
func CreateTable(table any) CreateTableStmt {
	return CreateTableStmt{
		state: &createTableState{
			table: tableName(table),
			model: table,
		},
	}
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateTableStmt) IfNotExists() CreateTableStmt {
	state := s.currentState()
	state.ifNotExists = true
	return CreateTableStmt{state: &state}
}

// Column appends a column definition.
func (s CreateTableStmt) Column(name any, typ Type) CreateTableStmt {
	return s.Columns(Column(name, typ))
}

// Columns appends column definitions.
func (s CreateTableStmt) Columns(columns ...ColumnDef) CreateTableStmt {
	state := s.currentState()
	state.columns = append(append([]ColumnDef(nil), state.columns...), columns...)
	return CreateTableStmt{state: &state}
}

// FromModel appends column definitions inferred from the table model's
// qrafter.Column fields. A field tag such as ddl:"VARCHAR(255)" overrides the
// type inferred from qrafter.Column[T]; ddl:"-" skips the field.
func (s CreateTableStmt) FromModel() CreateTableStmt {
	state := s.currentState()
	state.columns = append(append([]ColumnDef(nil), state.columns...), columnsFromModel(state.model)...)
	return CreateTableStmt{state: &state}
}

// Constraint appends a table-level constraint.
func (s CreateTableStmt) Constraint(constraint Constraint) CreateTableStmt {
	return s.Constraints(constraint)
}

// Constraints appends table-level constraints.
func (s CreateTableStmt) Constraints(constraints ...Constraint) CreateTableStmt {
	state := s.currentState()
	state.constraints = append(append([]Constraint(nil), state.constraints...), constraints...)
	return CreateTableStmt{state: &state}
}

// Render renders the CREATE TABLE statement.
func (s CreateTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(s.Render(d))
}

func (s CreateTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	state := s.currentState()
	if len(state.columns) == 0 && len(state.constraints) == 0 {
		panic(fmt.Errorf("CREATE TABLE %q must include at least one column or constraint", state.table))
	}

	w.WriteString("CREATE TABLE ")
	if state.ifNotExists {
		w.WriteString("IF NOT EXISTS ")
	}
	w.WriteString(d.QuoteIdent(state.table))
	w.WriteString(" (\n")

	item := 0
	for _, column := range state.columns {
		renderCreateTableDelimiter(w, item)
		column.render(w, d)
		item++
	}
	for i := range state.constraints {
		renderCreateTableDelimiter(w, item)
		state.constraints[i].render(w, d)
		item++
	}

	w.WriteString("\n)")
}

func (s CreateTableStmt) currentState() createTableState {
	if s.state == nil {
		return createTableState{}
	}
	return *s.state
}

func renderCreateTableDelimiter(w *strings.Builder, item int) {
	if item > 0 {
		w.WriteString(",\n")
	}
	w.WriteString("    ")
}
