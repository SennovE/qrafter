package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type DefaultExpression struct{}

var _ core.Selecter = DefaultExpression{}

func (e DefaultExpression) Render(w *strings.Builder, _ dialect.Renderer) {
	w.WriteString("DEFAULT")
}

func (e DefaultExpression) Tables() core.TablesSet {
	return nil
}

func Default() DefaultExpression {
	return DefaultExpression{}
}
