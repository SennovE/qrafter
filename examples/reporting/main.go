// Command reporting renders a larger analytical query.
package main

import (
	"fmt"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type CustomerTable struct {
	q.Table `table:"customers"`

	ID        q.Column[int64]      `db:"id"`
	Name      q.Column[string]     `db:"name"`
	DeletedAt q.Column[*time.Time] `db:"deleted_at"`
}

type OrderTable struct {
	q.Table `table:"orders"`

	ID         q.Column[int64]     `db:"id"`
	CustomerID q.Column[int64]     `db:"customer_id"`
	Status     q.Column[string]    `db:"status"`
	CreatedAt  q.Column[time.Time] `db:"created_at"`
}

type OrderItemTable struct {
	q.Table `table:"order_items"`

	ID        q.Column[int64] `db:"id"`
	OrderID   q.Column[int64] `db:"order_id"`
	Quantity  q.Column[int64] `db:"quantity"`
	UnitPrice q.Column[int64] `db:"unit_price_cents"`
}

func main() {
	customers := q.MustNewTable[CustomerTable]()
	orders := q.MustNewTable[OrderTable]()
	items := q.MustNewTable[OrderItemTable]()

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lineTotal := items.Quantity.Mul(items.UnitPrice)

	customerSpend := q.Select(
		orders.CustomerID,
		q.Count(orders.ID).As("orders_count"),
		q.Max(orders.CreatedAt).As("last_order_at"),
		q.Sum(lineTotal).As("total_spend_cents"),
	).
		Join(items, orders.ID.Eq(items.OrderID)).
		Where(orders.Status.Eq("paid"), orders.CreatedAt.Ge(since)).
		GroupBy(orders.CustomerID).
		CTE("customer_spend").
		WithColumns("customer_id", "orders_count", "last_order_at", "total_spend_cents")

	spend := customerSpend.Column("total_spend_cents")
	rank := q.Rank().
		Over(q.Window().OrderBy(spend.Desc())).
		As("spend_rank")

	sqlText, args, err := q.Select(
		customers.ID,
		customers.Name,
		customerSpend.Column("orders_count"),
		customerSpend.Column("last_order_at"),
		spend,
		rank,
	).
		Join(customerSpend, customers.ID.Eq(customerSpend.Column("customer_id"))).
		Where(customers.DeletedAt.IsNull()).
		OrderBy(spend.Desc()).
		Limit(20).
		Render(dialect.PostgreSQL{})
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlText)
	fmt.Println(args)
}
