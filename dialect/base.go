package dialect

type DialectRenderer interface {
	QuoteIdent(ident string) string
	Literal(value any) string
	LimitOffset(limit, offset int) string
}

type BaseDialect struct{}
