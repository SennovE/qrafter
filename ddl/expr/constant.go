package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type ConstExpression struct {
	v any
}

var _ CheckExperssion = ConstExpression{}

func (c ConstExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(d.Literal(c.v))
}

func (ConstExpression) expression() {}

func Literal(value any) ConstExpression {
	return ConstExpression{v: value}
}
