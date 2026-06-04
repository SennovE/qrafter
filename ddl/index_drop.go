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

func DropIndex(name string) DropIndexStmt {
	return DropIndexStmt{name: name}
}

func (s DropIndexStmt) IfExists() DropIndexStmt {
	s.ifExists = true
	return s
}

func (s DropIndexStmt) Concurrently() DropIndexStmt {
	s.concurrently = true
	return s
}

func (s DropIndexStmt) Cascade() DropIndexStmt {
	s.behavior = dropCascade
	return s
}

func (s DropIndexStmt) Restrict() DropIndexStmt {
	s.behavior = dropRestrict
	return s
}

func (s DropIndexStmt) OnTable(name string) DropIndexStmt {
	s.table = &name
	return s
}

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
