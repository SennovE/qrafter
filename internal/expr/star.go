package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type StarExpression struct{}

var _ core.Selecter = StarExpression{}

func (e StarExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("*")
}

func (e StarExpression) Tables() core.TablesSet {
	return nil
}

func Star() StarExpression {
	return StarExpression{}
}
