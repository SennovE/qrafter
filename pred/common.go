package pred

import "github.com/SennovE/qrafter/expr"

type Predicater interface {
	expr.Selecter
	Predicate()
}
