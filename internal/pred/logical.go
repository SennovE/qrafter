package pred

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type LogicalPredicate struct {
	ps []core.Predicater
	op string
}

const (
	OpAnd = "AND"
	OpOr  = "OR"
)

var _ = (core.Predicater)(LogicalPredicate{})

func (e LogicalPredicate) Predicate() {}

func (e LogicalPredicate) Render(d dialect.DialectRenderer) string {
	var res strings.Builder
	for i, p := range e.ps {
		if i > 0 {
			fmt.Fprintf(&res, " %s ", e.op)
		}
		res.WriteString(core.RenderChild(p, e.Precedence(), false, d))
	}
	return res.String()
}

func (e LogicalPredicate) Precedence() int {
	switch e.op {
	case OpOr:
		return core.PrecedenceOr
	case OpAnd:
		return core.PrecedenceAnd
	default:
		return core.PrecedenceComparison
	}
}

func (e LogicalPredicate) Tables() core.TablesSet {
	tables := make([]core.TablesSet, len(e.ps))
	for i, p := range e.ps {
		tables[i] = p.Tables()
	}
	return utils.UnionSets(tables...)
}

func Logical(op string, ps ...core.Predicater) LogicalPredicate {
	return LogicalPredicate{
		op: op,
		ps: ps,
	}
}
