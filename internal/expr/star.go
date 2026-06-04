package expr

import (
	"github.com/SennovE/qrafter/internal/core"
)

type StarExpression struct{}

var _ core.Selecter = StarExpression{}

func (e StarExpression) Tables() core.TablesSet {
	return nil
}

func Star() StarExpression {
	return StarExpression{}
}
