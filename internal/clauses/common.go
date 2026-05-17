package clauses

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type Clauser interface {
	Render(w *strings.Builder, d dialect.Renderer)
}
