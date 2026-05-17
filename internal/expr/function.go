package expr

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type FunctionExpression struct {
	name string
	args []core.Selecter
}

var _ = (core.Selecter)(FunctionExpression{})

func (e FunctionExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString(e.name)
	w.WriteString("(")
	core.RenderWithDelimiter(w, d, ", ", e.args)
	w.WriteString(")")
}

func (e FunctionExpression) Tables() core.TablesSet {
	tables := make([]core.TablesSet, len(e.args))
	for i, arg := range e.args {
		tables[i] = arg.Tables()
	}
	return utils.UnionSets(tables...)
}

func Function(name string, args ...core.Selecter) FunctionExpression {
	return FunctionExpression{
		name: name,
		args: args,
	}
}

type DistinctExpression struct {
	expr core.Selecter
}

var _ = (core.Selecter)(DistinctExpression{})

func (e DistinctExpression) Render(w *strings.Builder, d dialect.Renderer) {
	w.WriteString("DISTINCT ")
	e.expr.Render(w, d)
}

func (e DistinctExpression) Tables() core.TablesSet {
	return e.expr.Tables()
}

func Distinct(expr core.Selecter) DistinctExpression {
	return DistinctExpression{expr: expr}
}
