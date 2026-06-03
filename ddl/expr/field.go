package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type FieldExpression struct {
	name string
}

var _ CheckExperssion = FieldExpression{}

func (e FieldExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(d.Literal(e.name))
}

func (FieldExpression) expression() {}

func Field(name string) FieldExpression {
	return FieldExpression{name: name}
}
