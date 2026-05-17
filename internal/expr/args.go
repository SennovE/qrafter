package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type ArgExpression struct {
	v any
}

var _ = (core.Selecter)(ArgExpression{})

func (a ArgExpression) Tables() core.TablesSet {
	return nil
}

func (a ArgExpression) Render(w *strings.Builder, d dialect.Renderer) {
	if renderer, ok := d.(core.ArgRenderer); ok {
		w.WriteString(renderer.AddArg(a.v))
		return
	}
	w.WriteString(d.Literal(a.v))
}

func Param(value any) ArgExpression {
	return ArgExpression{v: value}
}
