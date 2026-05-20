package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type cteCollector struct {
	ctes []*core.CTERef
}

func renderReturning(w *strings.Builder, d dialect.Renderer, returning []core.Selecter) {
	if len(returning) == 0 {
		return
	}

	w.WriteString(" RETURNING ")
	core.RenderWithDelimiter(w, d, ", ", returning)
}

func (c cteCollector) RenderQueryExpression(_ *strings.Builder, _ dialect.Renderer) {}

func (c cteCollector) RenderSetOperand(_ *strings.Builder, _ dialect.Renderer) {}

func (c cteCollector) CTEs() []*core.CTERef {
	return c.ctes
}
