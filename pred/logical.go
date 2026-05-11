package pred

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/utils"
)

type LogicalPredicate struct {
	ps []Predicater
	op string
}

var _ = (Predicater)(LogicalPredicate{})

func (e LogicalPredicate) Predicate() {}

func (e LogicalPredicate) Render() string {
	var res strings.Builder
	for i, p := range e.ps {
		if i > 0 {
			fmt.Fprintf(&res, " %s ", e.op)
		}
		res.WriteString(p.Render())
	}
	return res.String()
}

func (e LogicalPredicate) Tables() qrafter.TablesSet {
	tables := make([]qrafter.TablesSet, len(e.ps))
	for i, p := range e.ps {
		tables[i] = p.Tables()
	}
	return utils.UnionSets(tables...)
}

func And(ps ...Predicater) LogicalPredicate {
	return LogicalPredicate{
		ps: ps,
		op: "AND",
	}
}
