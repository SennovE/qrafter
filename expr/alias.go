package expr

import (
	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/utils"
)

type AliasedExpression struct {
	Expr  Selecter
	Alias string
}

var _ = (Selecter)(AliasedExpression{})

func (a AliasedExpression) Tables() qrafter.TablesSet {
	return a.Expr.Tables()
}

func (a AliasedExpression) Render() string {
	return a.Expr.Render() + " AS " + utils.QuoteIdent(a.Alias)
}

func As(expr Selecter, alias string) AliasedExpression {
	return AliasedExpression{
		Expr:  expr,
		Alias: alias,
	}
}
