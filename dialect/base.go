package dialect

import (
	"fmt"

	"github.com/SennovE/qrafter/internal/utils"
)

// Renderer renders SQL syntax that differs across database dialects.
type Renderer interface {
	// QuoteIdent renders a SQL identifier.
	QuoteIdent(ident string) string
	// Literal renders an inline SQL literal.
	Literal(value any) string
	// Placeholder renders a bind placeholder for a one-based argument position.
	Placeholder(position int) string
	// LimitOffset renders dialect-specific LIMIT/OFFSET syntax.
	LimitOffset(limit, offset int) string
}

// BaseDialect renders ANSI-style identifiers, literals, placeholders, and limits.
type BaseDialect struct{}

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

// LimitOffset renders LIMIT and OFFSET clauses.
func (BaseDialect) LimitOffset(limit, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf("OFFSET %d", offset)
	default:
		return ""
	}
}
