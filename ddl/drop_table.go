package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// DropTableStmt builds a DROP TABLE statement.
type DropTableStmt struct {
	table    string
	ifExists bool
	cascade  bool
	restrict bool
}

// DropTable starts a DROP TABLE statement.
func DropTable(table any) DropTableStmt {
	return DropTableStmt{table: tableName(table)}
}

// IfExists adds IF EXISTS.
func (s DropTableStmt) IfExists() DropTableStmt {
	s.ifExists = true
	return s
}

// Cascade adds CASCADE.
func (s DropTableStmt) Cascade() DropTableStmt {
	s.cascade = true
	s.restrict = false
	return s
}

// Restrict adds RESTRICT.
func (s DropTableStmt) Restrict() DropTableStmt {
	s.restrict = true
	s.cascade = false
	return s
}

// Render renders the DROP TABLE statement.
func (s DropTableStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s DropTableStmt) MustRender(d dialect.Renderer) string {
	return mustRender(s.Render(d))
}

func (s DropTableStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if isSQLite(d) && (s.cascade || s.restrict) {
		unsupported(d, "DROP TABLE CASCADE/RESTRICT")
	}

	w.WriteString("DROP TABLE ")
	if s.ifExists {
		w.WriteString("IF EXISTS ")
	}
	w.WriteString(d.QuoteIdent(s.table))
	if s.cascade {
		w.WriteString(" CASCADE")
	}
	if s.restrict {
		w.WriteString(" RESTRICT")
	}
}
