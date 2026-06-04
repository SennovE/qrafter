package ddl

import "github.com/SennovE/qrafter/dialect"

type dropBehavior int

const (
	dropDefault dropBehavior = iota + 1
	dropRestrict
	dropCascade
)

// DropTableStmt builds a DROP TABLE statement.
type DropTableStmt struct {
	tables   []string
	ifExists bool
	behavior dropBehavior
}

// DropTable starts a DROP TABLE statement.
func DropTable(tables ...string) DropTableStmt {
	return DropTableStmt{tables: tables, behavior: dropDefault}
}

// IfExists adds IF EXISTS.
func (s DropTableStmt) IfExists() DropTableStmt {
	s.ifExists = true
	return s
}

// Cascade adds CASCADE.
func (s DropTableStmt) Cascade() DropTableStmt {
	s.behavior = dropCascade
	return s
}

// Restrict adds RESTRICT.
func (s DropTableStmt) Restrict() DropTableStmt {
	s.behavior = dropRestrict
	return s
}

// Render renders the DROP TABLE statement.
func (s DropTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s DropTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}
