package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// CreateTableStmt builds a CREATE TABLE statement.
type CreateTableStmt struct {
	name        string
	ifNotExists bool
	columns     []ColumnDef
	constraints []TableConstraint
}

// CreateTable starts a CREATE TABLE statement.
func CreateTable(name string) CreateTableStmt {
	return CreateTableStmt{name: name}
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateTableStmt) IfNotExists() CreateTableStmt {
	s.ifNotExists = true
	return s
}

// Column appends a column definition.
func (s CreateTableStmt) Column(name string, typ Type) CreateTableStmt {
	return s.Columns(Column(name, typ))
}

// Columns appends column definitions.
func (s CreateTableStmt) Columns(columns ...ColumnDef) CreateTableStmt {
	s.columns = append(s.columns, columns...)
	return s
}

// Constraint appends a table-level constraint.
func (s CreateTableStmt) Constraint(constraint TableConstraint) CreateTableStmt {
	return s.Constraints(constraint)
}

// Constraints appends table-level constraints.
func (s CreateTableStmt) Constraints(constraints ...TableConstraint) CreateTableStmt {
	s.constraints = append(s.constraints, constraints...)
	return s
}

// Render renders the CREATE TABLE statement.
func (s CreateTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s CreateTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if len(s.columns) == 0 && len(s.constraints) == 0 {
		panic(fmt.Errorf("CREATE TABLE %q must include at least one column or constraint", s.name))
	}

	w.WriteString("CREATE TABLE ")
	if s.ifNotExists {
		w.WriteString("IF NOT EXISTS ")
	}
	w.WriteString(d.QuoteIdent(s.name))
	w.WriteString(" (\n")

	item := 0
	for _, column := range s.columns {
		if item > 0 {
			w.WriteString(",\n")
		}
		w.WriteString("    ")
		column.Render(w, d)
		item++
	}
	for i := range s.constraints {
		if item > 0 {
			w.WriteString(",\n")
		}
		w.WriteString("    ")
		s.constraints[i].Render(s.name, w, d)
		item++
	}

	w.WriteString("\n)")
}
