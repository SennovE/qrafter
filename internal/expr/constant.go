package expr

import (
	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type ConstExpression struct {
	v any
}

var _ = (core.Selecter)(ConstExpression{})

func (c ConstExpression) Tables() core.TablesSet {
	return nil
}

func (c ConstExpression) Render(d dialect.DialectRenderer) string {
	return d.Literal(c.v)
}

func Const(value any) ConstExpression {
	return ConstExpression{v: value}
}
