package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type FunctionExpression struct {
	name string
	args []CheckExperssion
}

var _ CheckExperssion = FunctionExpression{}

func (e FunctionExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(e.name)
	w.WriteString("(")
	core.RenderWithDelimiter(w, d, ", ", e.args)
	w.WriteString(")")
}

func (FunctionExpression) expression() {}

func Function(name string, args ...CheckExperssion) FunctionExpression {
	return FunctionExpression{
		name: name,
		args: args,
	}
}
