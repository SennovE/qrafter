package qrafter

import (
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/core"
)

// Order represents an ORDER BY item.
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

// Asc returns an ascending ORDER BY item.
func Asc(v any) Order {
	return newOrder(v, "ASC")
}

// Desc returns a descending ORDER BY item.
func Desc(v any) Order {
	return newOrder(v, "DESC")
}

// NullsFirst returns an order item with NULLS FIRST.
func (o Order) NullsFirst() Order {
	o.nulls = "FIRST"
	return o
}

// NullsLast returns an order item with NULLS LAST.
func (o Order) NullsLast() Order {
	o.nulls = "LAST"
	return o
}

// Render writes the SQL representation of the order item.
func (o Order) Render(w *strings.Builder, d dialect.Renderer) {
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

// Tables returns table references used by the order item.
func (o Order) Tables() core.TablesSet {
	return o.expr.Tables()
}
