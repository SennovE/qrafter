package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type AliasedExpression struct {
	expr  core.Selecter
	alias string
}

var _ core.Selecter = AliasedExpression{}

func (a AliasedExpression) Tables() core.TablesSet {
	return a.expr.Tables()
}

func (a AliasedExpression) Render(w *strings.Builder, d dialect.Renderer) {
	a.expr.Render(w, d)
	w.WriteString(" AS ")
	w.WriteString(d.QuoteIdent(a.alias))
}

func Alias(expr core.Selecter, alias string) AliasedExpression {
	return AliasedExpression{
		expr:  expr,
		alias: alias,
	}
}
