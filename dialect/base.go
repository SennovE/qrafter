package dialect

import (
	"fmt"

	"github.com/SennovE/qrafter/internal/utils"
)

type DialectRenderer interface {
	QuoteIdent(ident string) string
	Literal(value any) string
	LimitOffset(limit, offset int) string
}

type BaseDialect struct{}

func (BaseDialect) QuoteIdent(ident string) string {
	return utils.QuoteWith(ident, `"`)
}

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

func (BaseDialect) LimitOffset(limit, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf(" LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf(" OFFSET %d", offset)
	default:
		return ""
	}
}
