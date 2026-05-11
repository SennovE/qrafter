package expr

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type AliasedExpression struct {
	expr  core.Selecter
	alias string
}

var _ = (core.Selecter)(AliasedExpression{})

func (a AliasedExpression) Tables() core.TablesSet {
	return a.expr.Tables()
}

func (a AliasedExpression) Render() string {
	return a.expr.Render() + " AS " + utils.QuoteIdent(a.alias)
}

func Alias(expr core.Selecter, alias string) AliasedExpression {
	return AliasedExpression{
		expr:  expr,
		alias: alias,
	}
}
