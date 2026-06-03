package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

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
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s DropTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s DropTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) && s.behavior != dropDefault {
		unsupported(d, "DROP TABLE CASCADE/RESTRICT")
	}

	w.WriteString("DROP TABLE ")
	if s.ifExists {
		w.WriteString("IF EXISTS ")
	}
	for i, t := range s.tables {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString(d.QuoteIdent(t))
	}
	switch s.behavior {
	case dropCascade:
		w.WriteString(" CASCADE")
	case dropRestrict:
		w.WriteString(" RESTRICT")
	}
}
