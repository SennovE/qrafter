package core

type Renderer interface {
	Render() string
}

type Selecter interface {
	Renderer
	Tables() TablesSet
}

type Predicater interface {
	Selecter
	Predicate()
}
