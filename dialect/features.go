package dialect

import "fmt"

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
