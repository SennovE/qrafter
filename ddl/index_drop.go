package ddl

// DropIndexStmt builds a DROP INDEX statement.
type DropIndexStmt struct {
	name     string
	table    *string
	ifExists bool
}
