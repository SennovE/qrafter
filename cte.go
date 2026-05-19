package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

// CommonTableExpression represents a named SQL common table expression.
type CommonTableExpression struct {
	ref *core.CTERef
}

// TableConfig returns the table configuration for the CTE reference.
func (cte CommonTableExpression) TableConfig() TableConfig {
	return TableConfig{Name: cte.TableRef().Name}
}

// TableRef returns the table reference used when querying the CTE.
func (cte CommonTableExpression) TableRef() core.TableRef {
	if cte.ref == nil {
		return core.TableRef{}
	}
	return core.TableRef{Name: cte.ref.Name, CTE: cte.ref}
}

// WithColumns adds an explicit column list to the CTE declaration.
func (cte CommonTableExpression) WithColumns(columns ...string) CommonTableExpression {
	if cte.ref == nil {
		cte.ref = &core.CTERef{}
	}
	cte.ref.Columns = append(cte.ref.Columns, columns...)
	return cte
}

// Recursive marks the CTE as recursive.
func (cte CommonTableExpression) Recursive() CommonTableExpression {
	if cte.ref == nil {
		cte.ref = &core.CTERef{}
	}
	cte.ref.Recursive = true
	return cte
}

// BindTableToCTE binds a table model to the CTE's table reference.
func BindTableToCTE[T any](cte CommonTableExpression) (T, error) {
	return bindWithTableRef[T](cte.TableRef())
}

// MustBindTableToCTE is like BindTableToCTE but panics on failure.
func MustBindTableToCTE[T any](cte CommonTableExpression) T {
	table, err := bindWithTableRef[T](cte.TableRef())
	if err != nil {
		panic(err)
	}
	return table
}

// Column returns an untyped column reference belonging to the CTE.
func (cte CommonTableExpression) Column(name string) Column[any] {
	var col Column[any]
	col.Bind(name, cte.TableRef())
	return col
}

// Render writes the CTE declaration.
func (cte CommonTableExpression) Render(w *strings.Builder, d dialect.Renderer) {
	cte.ref.Render(w, d)
}

// Union combines the CTE query with another query using UNION.
func (cte CommonTableExpression) Union(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(cte, compoundUnion, other)
}

// UnionAll combines the CTE query with another query using UNION ALL.
func (cte CommonTableExpression) UnionAll(other core.QueryExpression) CompoundQuery {
	return newCompoundQuery(cte, compoundUnionAll, other)
}

// RenderQueryExpression writes the CTE's underlying query expression.
func (cte CommonTableExpression) RenderQueryExpression(w *strings.Builder, d dialect.Renderer) {
	cte.ref.Query.RenderQueryExpression(w, d)
}

// RenderSetOperand writes the CTE's underlying query as a set operand.
func (cte CommonTableExpression) RenderSetOperand(w *strings.Builder, d dialect.Renderer) {
	cte.ref.Query.RenderSetOperand(w, d)
}

// CTEs returns common table expressions referenced by the CTE's query.
func (cte CommonTableExpression) CTEs() []*core.CTERef {
	return cte.ref.Query.CTEs()
}
