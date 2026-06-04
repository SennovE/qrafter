package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type alterIndexOpRenderer interface {
	renderAlterIndexOp(w *strings.Builder, d dialect.Renderer)
}

type AlterIndexStmt struct {
	name      string
	table     *string
	operation alterIndexOpRenderer
}

func AlterIndex(name string) AlterIndexStmt {
	return AlterIndexStmt{name: name}
}

type renameIndexStmt struct {
	oldName string
	newName string
}

func (s AlterIndexStmt) Rename(name string) AlterIndexStmt {
	s.operation = renameIndexStmt{oldName: s.name, newName: name}
	return s
}

func (s renameIndexStmt) renderAlterIndexOp(w *strings.Builder, d dialect.Renderer) {
	if isMySQL(d) {
		fmt.Fprintf(w, "RENAME INDEX %s TO %s", s.oldName, s.newName)
	} else {
		fmt.Fprintf(w, "%s RENAME TO %s", s.oldName, s.newName)
	}
}

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
	if isMySQL(d) {
		if s.table == nil {
			panic("MySQL require table name")
		}
		w.WriteString("ALTER TABLE")
		w.WriteString(d.QuoteIdent(*s.table))
	} else {
		w.WriteString("ALTER INDEX")
	}
	s.operation.renderAlterIndexOp(w, d)
}
