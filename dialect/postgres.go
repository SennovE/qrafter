package dialect

import "fmt"

const postgresDialectName = "PostgreSQL"

// PostgreSQL renders qrafter queries using PostgreSQL placeholder syntax.
type PostgreSQL struct {
	BaseDialect
}

// DialectName returns the dialect name.
func (PostgreSQL) DialectName() string {
	return postgresDialectName
}

// Placeholder renders a PostgreSQL numbered placeholder.
func (PostgreSQL) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}
