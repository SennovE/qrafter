package ddl

import "github.com/SennovE/qrafter/dialect"

// AlterIndexStmt builds an ALTER INDEX statement.
type AlterIndexStmt struct {
	name      string
	table     *string
	operation any
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

// OnTable sets the table name required by dialects such as MySQL.
func (s AlterIndexStmt) OnTable(name string) AlterIndexStmt {
	s.table = &name
	return s
}

// Render renders the ALTER INDEX operations.
func (s AlterIndexStmt) Render(d dialect.Renderer) (string, error) {
	return Render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s AlterIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}
