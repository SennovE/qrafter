package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
	"github.com/SennovE/qrafter/internal/utils"
)

type WindowSpec struct {
	partitionBy []core.Selecter
	orderBy     []core.Selecter
	frame       string
}

type windowExpression struct {
	expr core.Selecter
	spec WindowSpec
}

var _ = (core.Selecter)(windowExpression{})

func Window() WindowSpec {
	return WindowSpec{}
}

func PartitionBy(cols ...any) WindowSpec {
	return Window().PartitionBy(cols...)
}

func (s WindowSpec) PartitionBy(cols ...any) WindowSpec {
	s.partitionBy = append(s.partitionBy, asSelecters(cols)...)
	return s
}

func (s WindowSpec) OrderBy(items ...any) WindowSpec {
	s.orderBy = append(s.orderBy, asSelecters(items)...)
	return s
}

func (s WindowSpec) Frame(frame string) WindowSpec {
	s.frame = frame
	return s
}

func (s WindowSpec) Tables() core.TablesSet {
	tables := make([]core.TablesSet, 0, len(s.partitionBy)+len(s.orderBy))
	for _, expr := range s.partitionBy {
		tables = append(tables, expr.Tables())
	}
	for _, expr := range s.orderBy {
		tables = append(tables, expr.Tables())
	}
	return utils.UnionSets(tables...)
}

func (s WindowSpec) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString("(")

	rendered := false
	if len(s.partitionBy) > 0 {
		w.WriteString("PARTITION BY ")
		core.RenderWithDelimiter(w, d, ", ", s.partitionBy)
		rendered = true
	}

	if len(s.orderBy) > 0 {
		if rendered {
			w.WriteString(" ")
		}
		w.WriteString("ORDER BY ")
		core.RenderWithDelimiter(w, d, ", ", s.orderBy)
		rendered = true
	}

	if s.frame != "" {
		if rendered {
			w.WriteString(" ")
		}
		w.WriteString(s.frame)
	}

	w.WriteString(")")
}

func (e windowExpression) Tables() core.TablesSet {
	return utils.UnionSets(e.expr.Tables(), e.spec.Tables())
}

func (e windowExpression) Render(w *strings.Builder, d dialect.DialectRenderer) {
	e.expr.Render(w, d)
	w.WriteString(" OVER ")
	e.spec.Render(w, d)
}

func (e Expression) Over(specs ...WindowSpec) Expression {
	spec := WindowSpec{}
	if len(specs) > 0 {
		spec = specs[0]
	}
	return newExpression(windowExpression{
		expr: e.selecter,
		spec: spec,
	})
}

func WindowFunc(name string, args ...any) Expression {
	return newExpression(expr.Function(name, asSelecters(args)...))
}

func RowNumber() Expression {
	return WindowFunc("ROW_NUMBER")
}

func Rank() Expression {
	return WindowFunc("RANK")
}

func DenseRank() Expression {
	return WindowFunc("DENSE_RANK")
}

func Lag(v any, args ...any) Expression {
	return WindowFunc("LAG", append([]any{v}, args...)...)
}

func Lead(v any, args ...any) Expression {
	return WindowFunc("LEAD", append([]any{v}, args...)...)
}
