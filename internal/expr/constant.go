package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type ConstExpression struct {
	v any
}

var _ core.Selecter = ConstExpression{}

func (c ConstExpression) Tables() core.TablesSet {
	return nil
}

func (c ConstExpression) Value() any {
	return c.v
}

func (c ConstExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(d.Literal(c.v))
}

func Literal(value any) ConstExpression {
	return ConstExpression{v: value}
}
