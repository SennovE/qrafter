package dialect

import "fmt"

// PostgreSQL renders qrafter queries using PostgreSQL placeholder syntax.
type PostgreSQL struct {
	BaseDialect
}

// Placeholder renders a PostgreSQL numbered placeholder.
func (PostgreSQL) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}
