package clauses

import (
	"github.com/SennovE/qrafter/internal/core"
)

type JoinClause struct {
	Type       string
	Table      core.TableRef
	Predicates []core.Predicater
}
