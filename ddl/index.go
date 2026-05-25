package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

// CreateIndexStmt builds a CREATE INDEX statement.
type CreateIndexStmt struct {
	state *createIndexState
}

type createIndexState struct {
	name        string
	table       string
	columns     []string
	unique      bool
	ifNotExists bool
	where       string
	whereExpr   core.Renderer
}

// DropIndexStmt builds a DROP INDEX statement.
type DropIndexStmt struct {
	state *dropIndexState
}

type dropIndexState struct {
	name     string
	table    string
	ifExists bool
	cascade  bool
	restrict bool
}

// CreateIndex starts a CREATE INDEX statement.
func CreateIndex(name string) CreateIndexStmt {
	return CreateIndexStmt{state: &createIndexState{name: requireName("index", name)}}
}

// Unique makes the index unique.
func (s CreateIndexStmt) Unique() CreateIndexStmt {
	state := s.currentState()
	state.unique = true
	return CreateIndexStmt{state: &state}
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateIndexStmt) IfNotExists() CreateIndexStmt {
	state := s.currentState()
	state.ifNotExists = true
	return CreateIndexStmt{state: &state}
}

// On sets the indexed table and columns.
func (s CreateIndexStmt) On(table any, columns ...any) CreateIndexStmt {
	state := s.currentState()
	state.table = tableName(table)
	state.columns = columnNames(columns)
	return CreateIndexStmt{state: &state}
}

// Where adds a partial-index predicate. The predicate can be raw SQL or a
// qrafter predicate such as users.DeletedAt.IsNull().
func (s CreateIndexStmt) Where(predicate any) CreateIndexStmt {
	state := s.currentState()
	switch v := predicate.(type) {
	case string:
		state.where = requireName("index predicate", v)
		state.whereExpr = nil
	case core.Renderer:
		state.where = ""
		state.whereExpr = v
	default:
		panic(fmt.Errorf("unsupported index predicate %T", predicate))
	}
	return CreateIndexStmt{state: &state}
}

// Render renders the CREATE INDEX statement.
func (s CreateIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(s.Render(d))
}

func (s CreateIndexStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	state := s.currentState()
	if state.table == "" {
		panic(fmt.Errorf("CREATE INDEX %q must specify a table", state.name))
	}
	if len(state.columns) == 0 {
		panic(fmt.Errorf("CREATE INDEX %q must include at least one column", state.name))
	}
	if isMySQL(d) && state.ifNotExists {
		unsupported(d, "CREATE INDEX IF NOT EXISTS")
	}
	if isMySQL(d) && state.hasWhere() {
		unsupported(d, "PARTIAL INDEX")
	}

	w.WriteString("CREATE ")
	if state.unique {
		w.WriteString("UNIQUE ")
	}
	w.WriteString("INDEX ")
	if state.ifNotExists {
		w.WriteString("IF NOT EXISTS ")
	}
	w.WriteString(d.QuoteIdent(state.name))
	w.WriteString(" ON ")
	w.WriteString(d.QuoteIdent(state.table))
	w.WriteString(" (")
	renderColumnList(w, d, state.columns)
	w.WriteString(")")
	if state.hasWhere() {
		w.WriteString(" WHERE ")
		renderIndexWhere(w, d, &state)
	}
}

func (s CreateIndexStmt) currentState() createIndexState {
	if s.state == nil {
		return createIndexState{}
	}
	return *s.state
}

func (s *createIndexState) hasWhere() bool {
	return s.where != "" || s.whereExpr != nil
}

func renderIndexWhere(w *strings.Builder, d dialect.Renderer, state *createIndexState) {
	if state.where != "" {
		w.WriteString(state.where)
		return
	}

	var predicate strings.Builder
	state.whereExpr.Render(&predicate, d)
	w.WriteString(unqualifyIndexPredicate(predicate.String(), d, state.table))
}

func unqualifyIndexPredicate(predicate string, d dialect.Renderer, table string) string {
	prefix := d.QuoteIdent(table) + "."
	return strings.ReplaceAll(predicate, prefix, "")
}

// DropIndex starts a DROP INDEX statement.
func DropIndex(name string) DropIndexStmt {
	return DropIndexStmt{state: &dropIndexState{name: requireName("index", name)}}
}

// IfExists adds IF EXISTS.
func (s DropIndexStmt) IfExists() DropIndexStmt {
	state := s.currentState()
	state.ifExists = true
	return DropIndexStmt{state: &state}
}

// On sets the table for dialects that require it, such as MySQL.
func (s DropIndexStmt) On(table any) DropIndexStmt {
	state := s.currentState()
	state.table = tableName(table)
	return DropIndexStmt{state: &state}
}

// Cascade adds CASCADE.
func (s DropIndexStmt) Cascade() DropIndexStmt {
	state := s.currentState()
	state.cascade = true
	state.restrict = false
	return DropIndexStmt{state: &state}
}

// Restrict adds RESTRICT.
func (s DropIndexStmt) Restrict() DropIndexStmt {
	state := s.currentState()
	state.restrict = true
	state.cascade = false
	return DropIndexStmt{state: &state}
}

// Render renders the DROP INDEX statement.
func (s DropIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s DropIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(s.Render(d))
}

func (s DropIndexStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	state := s.currentState()
	if isMySQL(d) {
		renderMySQLDropIndex(w, d, &state)
		return
	}
	if isSQLite(d) && (state.cascade || state.restrict) {
		unsupported(d, "DROP INDEX CASCADE/RESTRICT")
	}

	w.WriteString("DROP INDEX ")
	if state.ifExists {
		w.WriteString("IF EXISTS ")
	}
	w.WriteString(d.QuoteIdent(state.name))
	if state.cascade {
		w.WriteString(" CASCADE")
	}
	if state.restrict {
		w.WriteString(" RESTRICT")
	}
}

func (s DropIndexStmt) currentState() dropIndexState {
	if s.state == nil {
		return dropIndexState{}
	}
	return *s.state
}

func renderMySQLDropIndex(w *strings.Builder, d dialect.Renderer, s *dropIndexState) {
	if s.ifExists {
		unsupported(d, "DROP INDEX IF EXISTS")
	}
	if s.cascade || s.restrict {
		unsupported(d, "DROP INDEX CASCADE/RESTRICT")
	}
	if s.table == "" {
		panic(fmt.Errorf("DROP INDEX %q must specify a table for MySQL", s.name))
	}

	w.WriteString("DROP INDEX ")
	w.WriteString(d.QuoteIdent(s.name))
	w.WriteString(" ON ")
	w.WriteString(d.QuoteIdent(s.table))
}
