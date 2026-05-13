package core

import "github.com/SennovE/qrafter/dialect"

type ArgRenderer interface {
	AddArg(value any) string
}

type ArgsRenderer struct {
	dialect.DialectRenderer
	args []any
}

func NewArgsRenderer(d dialect.DialectRenderer) *ArgsRenderer {
	return &ArgsRenderer{DialectRenderer: d}
}

func (r *ArgsRenderer) AddArg(value any) string {
	r.args = append(r.args, value)
	return r.Placeholder(len(r.args))
}

func (r *ArgsRenderer) Args() []any {
	return r.args
}
