package expr

import (
	"github.com/SennovE/qrafter/internal/core"
)

type AliasedExpression struct {
	expr  core.Selecter
	alias string
}

var _ core.Selecter = AliasedExpression{}

func (a AliasedExpression) Expr() core.Selecter {
	return a.expr
}

func (a AliasedExpression) Alias() string {
	return a.alias
}

func (a AliasedExpression) Tables() core.TablesSet {
	return a.expr.Tables()
}

func Alias(expr core.Selecter, alias string) AliasedExpression {
	return AliasedExpression{
		expr:  expr,
		alias: alias,
	}
}
