package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Orders struct {
	UserID q.Column[int]
	Amount q.Column[int]
	Status q.Column[string]
}

func (Orders) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "orders",
	}
}

type Users struct {
	ID   q.Column[int]
	Name q.Column[string]
}

func (Users) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "users",
	}
}

func TestSelectRender_WithCTE(t *testing.T) {
	OrdersTable := Orders{}
	require.NoError(t, q.Bind(&OrdersTable))
	UsersTable := Users{}
	require.NoError(t, q.Bind(&UsersTable))

	cte := q.
		Select(OrdersTable.UserID, q.Sum(OrdersTable.Amount)).
		Where(OrdersTable.Status.Eq("paid")).
		GroupBy(OrdersTable.UserID).
		CTE("total_amounts").
		WithColumns("user_id", "total")

	t.Run("Query with CTE binded to struct", func(t *testing.T) {
		type TotalAmountsTable struct {
			UserID q.Column[int]
			Total  q.Column[int]
		}

		TotalAmountsCTE := TotalAmountsTable{}
		err := cte.Bind(&TotalAmountsCTE)

		require.NoError(t, err)

		query := q.
			Select(UsersTable.Name, TotalAmountsCTE.Total).
			Join(cte, UsersTable.ID.Eq(TotalAmountsCTE.UserID)).
			Where(TotalAmountsCTE.Total.Gt(100))

		assert.Equal(
			t,
			`WITH "total_amounts" ("user_id", "total") AS (`+
				`SELECT "orders"."user_id", SUM("orders"."amount") FROM "orders" `+
				`WHERE "orders"."status" = 'paid'`+
				`) `+
				`SELECT "users"."name" FROM "users" `+
				`JOIN "total_amounts" ON "users"."id" = "total_amounts"."user_id" `+
				`WHERE "total_amounts"."total" > 100`,
			query.Render(dialect.PostgreSQL{}),
		)
	})

	t.Run("Query with CTE with obtaining columns by string name", func(t *testing.T) {
		query := q.
			Select(UsersTable.Name, cte.Column("total")).
			Join(cte, UsersTable.ID.Eq(cte.Column("user_id"))).
			Where(cte.Column("total").Gt(100))

		assert.Equal(
			t,
			`WITH "total_amounts" ("user_id", "total") AS (`+
				`SELECT "orders"."user_id", SUM("orders"."amount") FROM "orders" `+
				`WHERE "orders"."status" = 'paid'`+
				`) `+
				`SELECT "users"."name" FROM "users" `+
				`JOIN "total_amounts" ON "users"."id" = "total_amounts"."user_id" `+
				`WHERE "total_amounts"."total" > 100`,
			query.Render(dialect.PostgreSQL{}),
		)
	})
}
