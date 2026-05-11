package expr

import "github.com/SennovE/qrafter"

type Renderer interface {
	Render() string
}

type Selecter interface {
	Renderer
	Tables() qrafter.TablesSet
}
