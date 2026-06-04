package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type GroupByClause struct {
	Columns []core.Selecter
}
