package pred

import (
	"fmt"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/expr"
	"github.com/SennovE/qrafter/utils"
)

type BinaryPredicate struct {
	a, b expr.Selecter
	op   string
}

var _ = (Predicater)(BinaryPredicate{})

func (e BinaryPredicate) Predicate() {}

func (e BinaryPredicate) Render() string {
	return fmt.Sprintf("%s %s %s", e.a.Render(), e.op, e.b.Render())
}

func (e BinaryPredicate) Tables() qrafter.TablesSet {
	return utils.UnionSets(e.a.Tables(), e.b.Tables())
}

func Lt(a, b expr.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: "<",
	}
}

func Gt(a, b expr.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: ">",
	}
}

func Le(a, b expr.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: "<=",
	}
}

func Ge(a, b expr.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: ">=",
	}
}

func Eq(a, b expr.Selecter) BinaryPredicate {
	return BinaryPredicate{
		a:  a,
		b:  b,
		op: "=",
	}
}
