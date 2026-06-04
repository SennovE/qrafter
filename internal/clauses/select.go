package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type SelectClause struct {
	Columns []core.Selecter
}
