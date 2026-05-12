package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

type Order struct {
	expr      core.Selecter
	direction string
	nulls     string
}

var _ = (core.Selecter)(Order{})

func newOrder(v any, direction string) Order {
	return Order{
		expr:      asSelecter(v),
		direction: direction,
	}
}

func Asc(v any) Order {
	return newOrder(v, "ASC")
}

func Desc(v any) Order {
	return newOrder(v, "DESC")
}

func (o Order) NullsFirst() Order {
	o.nulls = "FIRST"
	return o
}

func (o Order) NullsLast() Order {
	o.nulls = "LAST"
	return o
}

func (o Order) Render(w *strings.Builder, d dialect.DialectRenderer) {
	o.expr.Render(w, d)
	if o.direction != "" {
		w.WriteString(" ")
		w.WriteString(o.direction)
	}
	if o.nulls != "" {
		w.WriteString(" NULLS ")
		w.WriteString(o.nulls)
	}
}

func (o Order) Tables() core.TablesSet {
	return o.expr.Tables()
}
