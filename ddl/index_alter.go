package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type alterIndexOpRenderer interface {
	renderAlterIndexOp(w *strings.Builder, d dialect.Renderer)
}

// AlterIndexStmt builds an ALTER INDEX statement.
type AlterIndexStmt struct {
	name      string
	table     *string
	operation alterIndexOpRenderer
}

// AlterIndex starts an ALTER INDEX statement.
func AlterIndex(name string) AlterIndexStmt {
	return AlterIndexStmt{name: name}
}

type renameIndexStmt struct {
	oldName string
	newName string
}

// Rename changes the index name.
func (s AlterIndexStmt) Rename(name string) AlterIndexStmt {
	s.operation = renameIndexStmt{oldName: s.name, newName: name}
	return s
}

func (s renameIndexStmt) renderAlterIndexOp(w *strings.Builder, d dialect.Renderer) {
	if isMySQL(d) {
		fmt.Fprintf(w, "RENAME INDEX %s TO %s", d.QuoteIdent(s.oldName), d.QuoteIdent(s.newName))
	} else {
		fmt.Fprintf(w, "%s RENAME TO %s", d.QuoteIdent(s.oldName), d.QuoteIdent(s.newName))
	}
}

// OnTable sets the table name required by dialects such as MySQL.
func (s AlterIndexStmt) OnTable(name string) AlterIndexStmt {
	s.table = &name
	return s
}

// Render renders the ALTER INDEX operations.
func (s AlterIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s AlterIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s AlterIndexStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if s.operation == nil {
		panic(fmt.Errorf("ALTER INDEX %q must include an operation", s.name))
	}
	if isMySQL(d) {
		if s.table == nil {
			panic("MySQL requires table name")
		}
		w.WriteString("ALTER TABLE ")
		w.WriteString(d.QuoteIdent(*s.table))
		w.WriteString(" ")
	} else {
		w.WriteString("ALTER INDEX ")
	}
	s.operation.renderAlterIndexOp(w, d)
}
