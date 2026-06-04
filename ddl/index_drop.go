package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// DropIndexStmt builds a DROP INDEX statement.
type DropIndexStmt struct {
	name string

	table *string

	ifExists     bool
	concurrently bool
	behavior     dropBehavior

	online bool
}

// DropIndex starts a DROP INDEX statement.
func DropIndex(name string) DropIndexStmt {
	return DropIndexStmt{name: name}
}

// IfExists adds IF EXISTS.
func (s DropIndexStmt) IfExists() DropIndexStmt {
	s.ifExists = true
	return s
}

// Concurrently adds CONCURRENTLY for dialects that support it.
func (s DropIndexStmt) Concurrently() DropIndexStmt {
	s.concurrently = true
	return s
}

// Cascade adds CASCADE.
func (s DropIndexStmt) Cascade() DropIndexStmt {
	s.behavior = dropCascade
	return s
}

// Restrict adds RESTRICT.
func (s DropIndexStmt) Restrict() DropIndexStmt {
	s.behavior = dropRestrict
	return s
}

// OnTable adds the table name required by dialects such as MySQL.
func (s DropIndexStmt) OnTable(name string) DropIndexStmt {
	s.table = &name
	return s
}

// Online adds ONLINE for dialects that support it.
func (s DropIndexStmt) Online() DropIndexStmt {
	s.online = true
	return s
}

// Render renders the DROP INDEX operations.
func (s DropIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s DropIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s DropIndexStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("DROP INDEX ")
	if s.concurrently {
		w.WriteString("CONCURRENTLY ")
	}
	if s.ifExists {
		w.WriteString("IF EXISTS ")
	}
	w.WriteString(d.QuoteIdent(s.name))
	if s.table != nil {
		w.WriteString(" ON ")
		w.WriteString(d.QuoteIdent(*s.table))
	}
	if s.online {
		w.WriteString(" ONLINE")
	}
	switch s.behavior {
	case dropCascade:
		w.WriteString(" CASCADE")
	case dropRestrict:
		w.WriteString(" RESTRICT")
	}
}
