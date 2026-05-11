package dialect

type DialectSelectRenderer interface {
	RenderSelect() string
}

type BaseDialect struct {
}
