package qrafter

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/pred"
)

func And(ps ...core.Predicater) pred.LogicalPredicate {
	return pred.Logical("AND", ps...)
}

func Or(ps ...core.Predicater) pred.LogicalPredicate {
	return pred.Logical("OR", ps...)
}

func Lt(a, b core.Selecter) pred.BinaryPredicate {
	return pred.Binary("<", a, b)
}

func Gt(a, b core.Selecter) pred.BinaryPredicate {
	return pred.Binary(">", a, b)
}

func Le(a, b core.Selecter) pred.BinaryPredicate {
	return pred.Binary("<=", a, b)
}

func Ge(a, b core.Selecter) pred.BinaryPredicate {
	return pred.Binary(">=", a, b)
}

func Eq(a, b core.Selecter) pred.BinaryPredicate {
	return pred.Binary("=", a, b)
}
