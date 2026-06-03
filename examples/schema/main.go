// Command schema renders a small PostgreSQL schema plan.
package main

import (
	"fmt"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

func main() {
	sqlText, err := ddl.Statements{
		ddl.CreateTable("organizations").
			IfNotExists().
			Columns(
				ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
				ddl.Column("name", ddl.Text()).NotNull().Unique(),
				ddl.Column("created_at", ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
			),

		ddl.CreateTable("users").
			IfNotExists().
			Columns(
				ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
				ddl.Column("organization_id", ddl.BigInt()).NotNull(),
				ddl.Column("email", ddl.VarChar(320)).NotNull(),
				ddl.Column("display_name", ddl.Text()).NotNull().Default(""),
				ddl.Column("deleted_at", ddl.TimestampTZ()),
				ddl.Column("created_at", ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
			).
			Constraints(
				ddl.Unique("email").Named("users_email_key"),
				ddl.ForeignKey("organization_id").
					References("organizations", "id").
					OnDelete(ddl.Cascade).
					Named("users_organization_id_fk"),
			),

		ddl.CreateIndex("users_active_email_idx").
			Unique().
			IfNotExists().
			On("users", "email").
			Where(),

		ddl.AlterTable("users").
			AddColumn(ddl.Column("last_seen_at", ddl.TimestampTZ())).
			DropDefault("display_name"),
	}.Render(dialect.PostgreSQL{})
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlText)
}
