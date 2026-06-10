package ddl

import "github.com/SennovE/qrafter/dialect"

// CreateTableStmt builds a CREATE TABLE statement.
type CreateTableStmt struct {
	TableName        string
	IfNotExistsFlag  bool
	ColumnDefs       []ColumnDef
	TableConstraints []TableConstraint
}

// CreateTable starts a CREATE TABLE statement.
func CreateTable(name string) CreateTableStmt {
	return CreateTableStmt{TableName: name}
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateTableStmt) IfNotExists() CreateTableStmt {
	s.IfNotExistsFlag = true
	return s
}

// Column appends a column definition.
func (s CreateTableStmt) Column(name string, typ Type) CreateTableStmt {
	return s.Columns(Column(name, typ))
}

// Columns appends column definitions.
func (s CreateTableStmt) Columns(columns ...ColumnDef) CreateTableStmt {
	s.ColumnDefs = append(s.ColumnDefs, columns...)
	return s
}

// Constraint appends a table-level constraint.
func (s CreateTableStmt) Constraint(constraint TableConstraint) CreateTableStmt {
	return s.Constraints(constraint)
}

// Constraints appends table-level constraints.
func (s CreateTableStmt) Constraints(constraints ...TableConstraint) CreateTableStmt {
	s.TableConstraints = append(s.TableConstraints, constraints...)
	return s
}

// Render renders the CREATE TABLE statement.
func (s CreateTableStmt) Render(d dialect.Renderer) (string, error) {
	return Render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}
