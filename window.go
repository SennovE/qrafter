package qrafter

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
	"github.com/SennovE/qrafter/internal/utils"
)

// WindowSpec represents a SQL window specification.
type WindowSpec struct {
	partitionBy []core.Selecter
	orderBy     []core.Selecter
	frame       *WindowFrame
}

// WindowFrame represents a SQL window frame clause.
type WindowFrame struct {
	mode  string
	start WindowFrameBound
	end   *WindowFrameBound
}

// WindowFrameMode represents a SQL window frame mode such as ROWS or RANGE.
type WindowFrameMode struct {
	mode string
}

// WindowFrameBound represents a SQL window frame bound.
type WindowFrameBound struct {
	value string
}

type windowExpression struct {
	expr core.Selecter
	spec WindowSpec
}

var _ core.Selecter = windowExpression{}

// Window starts an empty window specification.
func Window() WindowSpec {
	return WindowSpec{}
}

// PartitionBy starts a window specification with PARTITION BY expressions.
func PartitionBy(cols ...any) WindowSpec {
	return Window().PartitionBy(cols...)
}

// PartitionBy appends PARTITION BY expressions to the window specification.
func (s WindowSpec) PartitionBy(cols ...any) WindowSpec {
	s.partitionBy = append(append([]core.Selecter(nil), s.partitionBy...), asSelecters(cols)...)
	return s
}

// OrderBy appends ORDER BY expressions to the window specification.
func (s WindowSpec) OrderBy(items ...any) WindowSpec {
	s.orderBy = append(append([]core.Selecter(nil), s.orderBy...), asSelecters(items)...)
	return s
}

// Frame sets the frame clause for the window specification.
func (s WindowSpec) Frame(frame WindowFrame) WindowSpec {
	s.frame = &frame
	return s
}

// Rows returns a ROWS window frame mode.
func Rows() WindowFrameMode {
	return WindowFrameMode{mode: "ROWS"}
}

// Range returns a RANGE window frame mode.
func Range() WindowFrameMode {
	return WindowFrameMode{mode: "RANGE"}
}

// Groups returns a GROUPS window frame mode.
func Groups() WindowFrameMode {
	return WindowFrameMode{mode: "GROUPS"}
}

// Between returns a frame with BETWEEN start AND end bounds.
func (m WindowFrameMode) Between(start, end WindowFrameBound) WindowFrame {
	return WindowFrame{
		mode:  m.mode,
		start: start,
		end:   &end,
	}
}

// Bound returns a frame with a single bound.
func (m WindowFrameMode) Bound(bound WindowFrameBound) WindowFrame {
	return WindowFrame{
		mode:  m.mode,
		start: bound,
	}
}

// UnboundedPreceding returns a frame ending at UNBOUNDED PRECEDING.
func (m WindowFrameMode) UnboundedPreceding() WindowFrame {
	return m.Bound(UnboundedPreceding())
}

// CurrentRow returns a frame bound to CURRENT ROW.
func (m WindowFrameMode) CurrentRow() WindowFrame {
	return m.Bound(CurrentRow())
}

// Preceding returns a frame with a PRECEDING bound.
func (m WindowFrameMode) Preceding(v any) WindowFrame {
	return m.Bound(Preceding(v))
}

// Following returns a frame with a FOLLOWING bound.
func (m WindowFrameMode) Following(v any) WindowFrame {
	return m.Bound(Following(v))
}

// UnboundedPreceding returns an UNBOUNDED PRECEDING frame bound.
func UnboundedPreceding() WindowFrameBound {
	return WindowFrameBound{value: "UNBOUNDED PRECEDING"}
}

// UnboundedFollowing returns an UNBOUNDED FOLLOWING frame bound.
func UnboundedFollowing() WindowFrameBound {
	return WindowFrameBound{value: "UNBOUNDED FOLLOWING"}
}

// CurrentRow returns a CURRENT ROW frame bound.
func CurrentRow() WindowFrameBound {
	return WindowFrameBound{value: "CURRENT ROW"}
}

// Preceding returns a PRECEDING frame bound.
func Preceding(v any) WindowFrameBound {
	return WindowFrameBound{value: fmt.Sprint(v) + " PRECEDING"}
}

// Following returns a FOLLOWING frame bound.
func Following(v any) WindowFrameBound {
	return WindowFrameBound{value: fmt.Sprint(v) + " FOLLOWING"}
}

// FrameBound returns a custom frame bound.
func FrameBound(value string) WindowFrameBound {
	return WindowFrameBound{value: value}
}

// Render writes the SQL window frame.
func (f WindowFrame) Render(w *strings.Builder, d dialect.Renderer) {
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

// Render writes the SQL window frame bound.
func (b WindowFrameBound) Render(w *strings.Builder, _ dialect.Renderer) {
	w.WriteString(b.value)
}

// Tables returns table references used by the window specification.
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

// Render writes the SQL window specification.
func (s WindowSpec) Render(w *strings.Builder, d dialect.Renderer) {
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

func (e windowExpression) Render(w *strings.Builder, d dialect.Renderer) {
	e.expr.Render(w, d)
	w.WriteString(" OVER ")
	e.spec.Render(w, d)
}

// Over returns an expression with an OVER clause.
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

// WindowFunc builds a SQL window function expression.
func WindowFunc(name string, args ...any) Expression {
	return newExpression(expr.Function(name, asSelecters(args)...))
}

// RowNumber builds a ROW_NUMBER window function expression.
func RowNumber() Expression {
	return WindowFunc("ROW_NUMBER")
}

// Rank builds a RANK window function expression.
func Rank() Expression {
	return WindowFunc("RANK")
}

// DenseRank builds a DENSE_RANK window function expression.
func DenseRank() Expression {
	return WindowFunc("DENSE_RANK")
}

// Lag builds a LAG window function expression.
func Lag(v any, args ...any) Expression {
	return WindowFunc("LAG", append([]any{v}, args...)...)
}

// Lead builds a LEAD window function expression.
func Lead(v any, args ...any) Expression {
	return WindowFunc("LEAD", append([]any{v}, args...)...)
}
