package dialect

import (
	"fmt"

	"github.com/SennovE/qrafter/internal/utils"
)

// Renderer renders SQL syntax that differs across database dialects.
type Renderer interface {
	// DialectName returns a user-facing dialect name.
	DialectName() string
	// QuoteIdent renders a SQL identifier.
	QuoteIdent(ident string) string
	// Literal renders an inline SQL literal.
	Literal(value any) string
	// Placeholder renders a bind placeholder for a one-based argument position.
	Placeholder(position int) string
	// CompileNode lets a dialect override rendering for a specific compiler node.
	CompileNode(c Compiler, node any) bool
}

// Compiler is the minimal surface dialects use to override node rendering.
type Compiler interface {
	Write(s string)
	WriteInt(n int)
	Compile(node any)
	CompileList(nodes []any, delimiter string)
	Unsupported(feature string)
	Renderer() Renderer
}

// BaseDialect renders ANSI-style identifiers, literals, placeholders, and limits.
type BaseDialect struct{}

// DialectName returns the dialect name.
func (BaseDialect) DialectName() string {
	return "BaseDialect"
}

// QuoteIdent renders a double-quoted identifier.
func (BaseDialect) QuoteIdent(ident string) string {
	return utils.QuoteWith(ident, `"`)
}

// Literal renders a basic SQL literal.
func (BaseDialect) Literal(value any) string {
	switch v := value.(type) {
	case nil:
		return "NULL"
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case string:
		return utils.QuoteWith(v, `'`)
	default:
		return fmt.Sprint(v)
	}
}

// Placeholder renders a question-mark placeholder.
func (BaseDialect) Placeholder(_ int) string {
	return "?"
}

// CompileNode returns false to use the default compiler implementation.
func (BaseDialect) CompileNode(_ Compiler, _ any) bool {
	return false
}
