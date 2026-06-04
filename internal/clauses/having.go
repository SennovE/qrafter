package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type HavingClause struct {
	Predicates []core.Predicater
}
