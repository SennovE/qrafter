package qrafter

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
	"github.com/SennovE/qrafter/internal/utils"
)

type WindowSpec struct {
	partitionBy []core.Selecter
	orderBy     []core.Selecter
	frame       *WindowFrame
}

type WindowFrame struct {
	mode  string
	start WindowFrameBound
	end   *WindowFrameBound
}

type WindowFrameMode struct {
	mode string
}

type WindowFrameBound struct {
	value string
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

func (s WindowSpec) Frame(frame WindowFrame) WindowSpec {
	s.frame = &frame
	return s
}

func Rows() WindowFrameMode {
	return WindowFrameMode{mode: "ROWS"}
}

func Range() WindowFrameMode {
	return WindowFrameMode{mode: "RANGE"}
}

func Groups() WindowFrameMode {
	return WindowFrameMode{mode: "GROUPS"}
}

func (m WindowFrameMode) Between(start, end WindowFrameBound) WindowFrame {
	return WindowFrame{
		mode:  m.mode,
		start: start,
		end:   &end,
	}
}

func (m WindowFrameMode) Bound(bound WindowFrameBound) WindowFrame {
	return WindowFrame{
		mode:  m.mode,
		start: bound,
	}
}

func (m WindowFrameMode) UnboundedPreceding() WindowFrame {
	return m.Bound(UnboundedPreceding())
}

func (m WindowFrameMode) CurrentRow() WindowFrame {
	return m.Bound(CurrentRow())
}

func (m WindowFrameMode) Preceding(v any) WindowFrame {
	return m.Bound(Preceding(v))
}

func (m WindowFrameMode) Following(v any) WindowFrame {
	return m.Bound(Following(v))
}

func UnboundedPreceding() WindowFrameBound {
	return WindowFrameBound{value: "UNBOUNDED PRECEDING"}
}

func UnboundedFollowing() WindowFrameBound {
	return WindowFrameBound{value: "UNBOUNDED FOLLOWING"}
}

func CurrentRow() WindowFrameBound {
	return WindowFrameBound{value: "CURRENT ROW"}
}

func Preceding(v any) WindowFrameBound {
	return WindowFrameBound{value: fmt.Sprint(v) + " PRECEDING"}
}

func Following(v any) WindowFrameBound {
	return WindowFrameBound{value: fmt.Sprint(v) + " FOLLOWING"}
}

func FrameBound(value string) WindowFrameBound {
	return WindowFrameBound{value: value}
}

func (f WindowFrame) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString(f.mode)
	if f.end != nil {
		w.WriteString(" BETWEEN ")
		f.start.Render(w, d)
		w.WriteString(" AND ")
		f.end.Render(w, d)
		return
	}

	w.WriteString(" ")
	f.start.Render(w, d)
}

func (b WindowFrameBound) Render(w *strings.Builder, d dialect.DialectRenderer) {
	w.WriteString(b.value)
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

	if s.frame != nil {
		if rendered {
			w.WriteString(" ")
		}
		s.frame.Render(w, d)
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
