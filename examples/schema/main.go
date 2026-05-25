// Command schema renders a small PostgreSQL schema plan.
package main

import (
	"fmt"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

type OrganizationTable struct {
	q.Table `table:"organizations"`

	ID        q.Column[int64]     `db:"id"`
	Name      q.Column[string]    `db:"name"`
	CreatedAt q.Column[time.Time] `db:"created_at"`
}

type UserTable struct {
	q.Table `table:"users"`

	ID             q.Column[int64]      `db:"id"`
	OrganizationID q.Column[int64]      `db:"organization_id"`
	Email          q.Column[string]     `db:"email" ddl:"VARCHAR(320)"`
	DisplayName    q.Column[string]     `db:"display_name"`
	DeletedAt      q.Column[*time.Time] `db:"deleted_at"`
	CreatedAt      q.Column[time.Time]  `db:"created_at"`
}

func main() {
	orgs := q.MustNewTable[OrganizationTable]()
	users := q.MustNewTable[UserTable]()

	sqlText, err := ddl.Statements{
		ddl.CreateTable(orgs).
			IfNotExists().
			Columns(
				ddl.Column(orgs.ID, ddl.BigSerial()).PrimaryKey(),
				ddl.Column(orgs.Name, ddl.Text()).NotNull().Unique(),
				ddl.Column(orgs.CreatedAt, ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
			),

		ddl.CreateTable(users).
			IfNotExists().
			Columns(
				ddl.Column(users.ID, ddl.BigSerial()).PrimaryKey(),
				ddl.Column(users.OrganizationID, ddl.BigInt()).NotNull(),
				ddl.Column(users.Email, ddl.VarChar(320)).NotNull(),
				ddl.Column(users.DisplayName, ddl.Text()).NotNull().Default(""),
				ddl.Column(users.DeletedAt, ddl.TimestampTZ()),
				ddl.Column(users.CreatedAt, ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
			).
			Constraints(
				ddl.Unique(users.Email).Named("users_email_key"),
				ddl.ForeignKey(users.OrganizationID).
					References(orgs, orgs.ID).
					OnDelete(ddl.Cascade).
					Named("users_organization_id_fk"),
			),

		ddl.CreateIndex("users_active_email_idx").
			Unique().
			IfNotExists().
			On(users, users.Email).
			Where(users.DeletedAt.IsNull()),

		ddl.AlterTable(users).
			AddColumn(ddl.Column("last_seen_at", ddl.TimestampTZ())).
			DropDefault(users.DisplayName),
	}.Render(dialect.PostgreSQL{})
	if err != nil {
		panic(err)
	}

	fmt.Println(sqlText)
}
