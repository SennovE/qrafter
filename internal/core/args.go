package core

import "github.com/SennovE/qrafter/dialect"

type ArgRenderer interface {
	AddArg(value any) string
}

type ArgsRenderer struct {
	dialect.Renderer
	args []any
}

func NewArgsRenderer(d dialect.Renderer) *ArgsRenderer {
	return &ArgsRenderer{Renderer: d}
}

func (r *ArgsRenderer) AddArg(value any) string {
	r.args = append(r.args, value)
	return r.Placeholder(len(r.args))
}

func (r *ArgsRenderer) Args() []any {
	return r.args
}
