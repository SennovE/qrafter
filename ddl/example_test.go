package ddl_test

import (
	"fmt"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

func ExampleCreateTable() {
	sql := ddl.CreateTable("users").
		IfNotExists().
		Columns(
			ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
			ddl.Column("email", ddl.VarChar(320)).NotNull(),
			ddl.Column("org_id", ddl.BigInt()),
			ddl.Column("created_at", ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
		).
		Constraints(
			ddl.Unique("email").Named("users_email_key"),
			ddl.ForeignKey("org_id").
				References("orgs", "id").
				OnDelete(ddl.Cascade).
				Named("users_org_id_fk"),
		).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// CREATE TABLE IF NOT EXISTS "users" (
	//     "id" BIGSERIAL PRIMARY KEY,
	//     "email" VARCHAR(320) NOT NULL,
	//     "org_id" BIGINT,
	//     "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
	//     CONSTRAINT "users_email_key" UNIQUE ("email"),
	//     CONSTRAINT "users_org_id_fk" FOREIGN KEY ("org_id") REFERENCES "orgs" ("id") ON DELETE CASCADE
	// )
}

func ExampleColumn() {
	sql := ddl.CreateTable("email_archive").
		Columns(
			ddl.Column("email", ddl.Text()),
			ddl.Column("archived_at", ddl.TimestampTZ()).NotNull(),
		).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// CREATE TABLE "email_archive" (
	//     "email" TEXT,
	//     "archived_at" TIMESTAMPTZ NOT NULL
	// )
}

func ExampleCreateIndex() {
	sql := ddl.CreateIndex("users_email_active_idx").
		Unique().
		IfNotExists().
		On("users", ddl.KeyCol("email")).
		Where(ddl.Col("deleted_at").IsNull()).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// CREATE UNIQUE INDEX IF NOT EXISTS "users_email_active_idx" ON "users" ("email") WHERE "deleted_at" IS NULL
}

func ExampleAlterTable() {
	sql := ddl.AlterTable("users").
		AddColumn(ddl.Column("display_name", ddl.Text()).NotNull().Default("")).
		SetDefaultExpr("created_at", "now()").
		SetNotNull("email").
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	//     ADD COLUMN "display_name" TEXT NOT NULL DEFAULT '',
	//     ALTER COLUMN "created_at" SET DEFAULT now(),
	//     ALTER COLUMN "email" SET NOT NULL
}

func ExampleAlterTableStmt_RenameColumn() {
	sql := ddl.AlterTable("users").
		RenameColumn("name", "display_name").
		DropDefault("created_at").
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	//     RENAME COLUMN "name" TO "display_name",
	//     ALTER COLUMN "created_at" DROP DEFAULT
}

func ExampleAlterTableStmt_AddConstraint() {
	sql := ddl.AlterTable("users").
		AddConstraint(
			ddl.ForeignKey("org_id").
				References("orgs", "id").
				OnDelete(ddl.Cascade).
				Named("users_org_id_fk"),
		).
		AddConstraint(ddl.Unique("email").Named("users_email_key")).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	//     ADD CONSTRAINT "users_org_id_fk" FOREIGN KEY ("org_id") REFERENCES "orgs" ("id") ON DELETE CASCADE,
	//     ADD CONSTRAINT "users_email_key" UNIQUE ("email")
}

func ExampleAlterTableStmt_Render_unsupportedFeature() {
	_, err := ddl.AlterTable("users").
		AlterColumnType("email", ddl.Text()).
		Render(dialect.SQLite{})

	fmt.Println(err)

	// Output:
	// SQLite dialect does not support ALTER COLUMN TYPE
}

func ExampleCheck() {
	sql := ddl.AlterTable("users").
		AddConstraint(
			ddl.Check(ddl.Col("age").Ge(0)).Named("chk_users"),
		).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	//     ADD CONSTRAINT "chk_users" CHECK ("age" >= 0)
}
