package expr

import (
	"github.com/SennovE/qrafter/internal/core"
)

type DefaultExpression struct{}

var _ core.Selecter = DefaultExpression{}

func (e DefaultExpression) Tables() core.TablesSet {
	return nil
}

func Default() DefaultExpression {
	return DefaultExpression{}
}
