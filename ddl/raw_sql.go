package ddl

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// RawSQLStmt renders a SQL statement verbatim.
type RawSQLStmt struct {
	sql string
}

// RawSQL creates a statement rendered verbatim.
func RawSQL(sql string) RawSQLStmt {
	return RawSQLStmt{sql: sql}
}

// Render renders the raw SQL statement.
func (s RawSQLStmt) Render(_ dialect.Renderer) (string, error) {
	return strings.TrimSpace(s.sql), nil
}

// MustRender renders the raw SQL statement and never returns an error.
func (s RawSQLStmt) MustRender(d dialect.Renderer) string {
	sql, _ := s.Render(d)
	return sql
}
