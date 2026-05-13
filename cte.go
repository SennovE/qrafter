package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type CommonTableExpression struct {
	ref *core.CTERef
}

func (cte CommonTableExpression) TableConfig() TableConfig {
	return TableConfig{Name: cte.TableRef().Name}
}

func (cte CommonTableExpression) TableRef() core.TableRef {
	if cte.ref == nil {
		return core.TableRef{}
	}
	return core.TableRef{Name: cte.ref.Name, CTE: cte.ref}
}

func (cte CommonTableExpression) WithColumns(columns ...string) CommonTableExpression {
	if cte.ref == nil {
		cte.ref = &core.CTERef{}
	}
	cte.ref.Columns = append(cte.ref.Columns, columns...)
	return cte
}

func (cte CommonTableExpression) Recursive() CommonTableExpression {
	if cte.ref == nil {
		cte.ref = &core.CTERef{}
	}
	cte.ref.Recursive = true
	return cte
}

func (cte CommonTableExpression) Bind(table any) error {
	return bindWithTableRef(table, cte.TableRef())
}

func (cte CommonTableExpression) Column(name string) Column[any] {
	var col Column[any]
	col.Bind(name, cte.TableRef())
	return col
}

func (cte CommonTableExpression) Render(w *strings.Builder, d dialect.DialectRenderer) {
	cte.ref.Render(w, d)
}
