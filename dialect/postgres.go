package dialect

import "fmt"

type PostgreSQL struct {
	BaseDialect
}

func (PostgreSQL) Placeholder(position int) string {
	return fmt.Sprintf("$%d", position)
}
