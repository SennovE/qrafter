package dialect

import (
	"fmt"
	"strings"
)

// RendererWrapper is implemented by renderers that wrap another dialect
// renderer, such as qrafter's argument-collecting renderer.
type RendererWrapper interface {
	UnwrapRenderer() Renderer
}

// UnsupportedFeatureError is raised when a query uses syntax that a dialect
// cannot render correctly.
type UnsupportedFeatureError struct {
	Dialect string
	Feature string
}

func (e UnsupportedFeatureError) Error() string {
	return fmt.Sprintf("%s dialect does not support %s", e.Dialect, e.Feature)
}

// RecoverFromUnsupportedFeatureError recovers from a panic with UnsupportedFeatureError
// and stores it in err; all other panics are re-thrown.
func RecoverFromUnsupportedFeatureError(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(UnsupportedFeatureError); ok {
			*err = e
		} else {
			panic(r)
		}
	}
}

// DefaultValuesRenderer customizes INSERT DEFAULT VALUES rendering.
type DefaultValuesRenderer interface {
	RenderDefaultValues(w *strings.Builder)
}

// ReturningRenderer customizes RETURNING clause rendering.
type ReturningRenderer interface {
	RenderReturning(w *strings.Builder, renderItems func())
}

// OrderRenderer customizes ORDER BY item rendering.
type OrderRenderer interface {
	RenderOrder(w *strings.Builder, renderExpr func(), direction, nulls string)
}

// JoinRenderer customizes JOIN clause rendering.
type JoinRenderer interface {
	RenderJoin(w *strings.Builder, joinType string, renderTable, renderPredicates func())
}

// UpdateRenderer customizes UPDATE target and FROM-clause rendering.
type UpdateRenderer interface {
	RenderUpdateTarget(w *strings.Builder, renderTarget func(), hasFrom bool, renderFrom func())
	RenderUpdateFrom(w *strings.Builder, renderFrom func())
}

// DeleteRenderer customizes DELETE target and USING-clause rendering.
type DeleteRenderer interface {
	RenderDeleteTarget(w *strings.Builder, renderTarget, renderTargetName func(), hasUsing bool, renderUsing func())
	RenderDeleteUsing(w *strings.Builder, renderUsing func())
}

// UnwrapRenderer returns the underlying dialect renderer.
func UnwrapRenderer(d Renderer) Renderer {
	for {
		wrapped, ok := d.(RendererWrapper)
		if !ok {
			return d
		}
		d = wrapped.UnwrapRenderer()
	}
}

// RenderDefaultValues writes dialect-specific INSERT default row syntax.
func RenderDefaultValues(w *strings.Builder, d Renderer) {
	if renderer, ok := UnwrapRenderer(d).(DefaultValuesRenderer); ok {
		renderer.RenderDefaultValues(w)
		return
	}
	w.WriteString("\nDEFAULT VALUES")
}

// RenderReturning writes a dialect-specific RETURNING clause.
func RenderReturning(w *strings.Builder, d Renderer, renderItems func()) {
	if renderer, ok := UnwrapRenderer(d).(ReturningRenderer); ok {
		renderer.RenderReturning(w, renderItems)
		return
	}
	w.WriteString("\nRETURNING ")
	renderItems()
}

// RenderOrder writes a dialect-specific ORDER BY item.
func RenderOrder(w *strings.Builder, d Renderer, renderExpr func(), direction, nulls string) {
	if renderer, ok := UnwrapRenderer(d).(OrderRenderer); ok {
		renderer.RenderOrder(w, renderExpr, direction, nulls)
		return
	}

	renderOrderDefault(w, renderExpr, direction, nulls)
}

// RenderJoin writes a dialect-specific JOIN clause.
func RenderJoin(w *strings.Builder, d Renderer, joinType string, renderTable, renderPredicates func()) {
	if renderer, ok := UnwrapRenderer(d).(JoinRenderer); ok {
		renderer.RenderJoin(w, joinType, renderTable, renderPredicates)
		return
	}

	renderJoinDefault(w, joinType, renderTable, renderPredicates)
}

// RenderUpdateTarget writes a dialect-specific UPDATE target.
func RenderUpdateTarget(w *strings.Builder, d Renderer, renderTarget func(), hasFrom bool, renderFrom func()) {
	if renderer, ok := UnwrapRenderer(d).(UpdateRenderer); ok {
		renderer.RenderUpdateTarget(w, renderTarget, hasFrom, renderFrom)
		return
	}

	w.WriteString("UPDATE ")
	renderTarget()
}

// RenderUpdateFrom writes a dialect-specific UPDATE source-table clause.
func RenderUpdateFrom(w *strings.Builder, d Renderer, renderFrom func()) {
	if renderer, ok := UnwrapRenderer(d).(UpdateRenderer); ok {
		renderer.RenderUpdateFrom(w, renderFrom)
		return
	}

	w.WriteString("\nFROM ")
	renderFrom()
}

// RenderDeleteTarget writes a dialect-specific DELETE target.
func RenderDeleteTarget(
	w *strings.Builder,
	d Renderer,
	renderTarget func(),
	renderTargetName func(),
	hasUsing bool,
	renderUsing func(),
) {
	if renderer, ok := UnwrapRenderer(d).(DeleteRenderer); ok {
		renderer.RenderDeleteTarget(w, renderTarget, renderTargetName, hasUsing, renderUsing)
		return
	}

	w.WriteString("DELETE FROM ")
	renderTarget()
}

// RenderDeleteUsing writes a dialect-specific DELETE source-table clause.
func RenderDeleteUsing(w *strings.Builder, d Renderer, renderUsing func()) {
	if renderer, ok := UnwrapRenderer(d).(DeleteRenderer); ok {
		renderer.RenderDeleteUsing(w, renderUsing)
		return
	}

	w.WriteString("\nUSING ")
	renderUsing()
}

func renderOrderDefault(w *strings.Builder, renderExpr func(), direction, nulls string) {
	renderExpr()
	if direction != "" {
		w.WriteString(" ")
		w.WriteString(direction)
	}
	if nulls != "" {
		w.WriteString(" NULLS ")
		w.WriteString(nulls)
	}
}

func renderJoinDefault(w *strings.Builder, joinType string, renderTable, renderPredicates func()) {
	w.WriteString("\n")
	w.WriteString(joinType)
	w.WriteString(" ")
	renderTable()
	renderPredicates()
}
