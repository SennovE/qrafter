package ddl

// CreateIndexStmt builds a CREATE INDEX statement.
type CreateIndexStmt struct {
	name   string
	table  string

	unique      bool
	ifNotExists bool
}
