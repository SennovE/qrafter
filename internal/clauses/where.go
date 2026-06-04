package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type WhereClause struct {
	Predicates []core.Predicater
}
