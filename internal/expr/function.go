package expr

import (
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

type FunctionExpression struct {
	name string
	args []core.Selecter
}

var _ core.Selecter = FunctionExpression{}

func (e FunctionExpression) Name() string {
	return e.name
}

func (e FunctionExpression) Args() []core.Selecter {
	return e.args
}

func (e FunctionExpression) Tables() core.TablesSet {
	tables := make([]core.TablesSet, len(e.args))
	for i, arg := range e.args {
		tables[i] = arg.Tables()
	}
	return utils.UnionSets(tables...)
}

func Function(name string, args ...core.Selecter) FunctionExpression {
	return FunctionExpression{
		name: name,
		args: args,
	}
}

type DistinctExpression struct {
	expr core.Selecter
}

var _ core.Selecter = DistinctExpression{}

func (e DistinctExpression) Expr() core.Selecter {
	return e.expr
}

func (e DistinctExpression) Tables() core.TablesSet {
	return e.expr.Tables()
}

func Distinct(expr core.Selecter) DistinctExpression {
	return DistinctExpression{expr: expr}
}
