package core

import "github.com/SennovE/qrafter/dialect"

type Renderer interface {
	Render(d dialect.DialectRenderer) string
}

type Selecter interface {
	Renderer
	Tables() TablesSet
}

type Predicater interface {
	Selecter
	Predicate()
}
