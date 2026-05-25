package ddl_test

import (
	"fmt"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

type exampleUser struct {
	q.Table `table:"users"`

	ID        q.Column[int64]      `db:"id"`
	Email     q.Column[string]     `db:"email" ddl:"VARCHAR(320)"`
	OrgID     q.Column[int64]      `db:"org_id"`
	DeletedAt q.Column[*time.Time] `db:"deleted_at"`
	CreatedAt q.Column[time.Time]  `db:"created_at"`
	Ignored   q.Column[string]     `ddl:"-"`
}

type exampleOrg struct {
	q.Table `table:"orgs"`

	ID q.Column[int64] `db:"id"`
}

func ExampleCreateTable() {
	users := q.MustNewTable[exampleUser]()
	orgs := q.MustNewTable[exampleOrg]()

	sql := ddl.CreateTable(users).
		IfNotExists().
		Columns(
			ddl.Column(users.ID, ddl.BigSerial()).PrimaryKey(),
			ddl.Column(users.Email, ddl.VarChar(320)).NotNull(),
			ddl.Column(users.OrgID, ddl.BigInt()),
			ddl.Column(users.CreatedAt, ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
		).
		Constraints(
			ddl.Unique(users.Email).Named("users_email_key"),
			ddl.ForeignKey(users.OrgID).
				References(orgs, orgs.ID).
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

func ExampleCreateTableStmt_FromModel() {
	users := q.MustNewTable[exampleUser]()

	sql := ddl.CreateTable(users).
		FromModel().
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// CREATE TABLE "users" (
	//     "id" BIGINT,
	//     "email" VARCHAR(320),
	//     "org_id" BIGINT,
	//     "deleted_at" TIMESTAMPTZ,
	//     "created_at" TIMESTAMPTZ
	// )
}

func ExampleColumn() {
	users := q.MustNewTable[exampleUser]()

	sql := ddl.CreateTable("email_archive").
		Columns(
			ddl.Column(users.Email, ddl.Text()),
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
	users := q.MustNewTable[exampleUser]()

	sql := ddl.CreateIndex("users_email_active_idx").
		Unique().
		IfNotExists().
		On(users, users.Email).
		Where(users.DeletedAt.IsNull()).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// CREATE UNIQUE INDEX IF NOT EXISTS "users_email_active_idx" ON "users" ("email") WHERE "deleted_at" IS NULL
}

func ExampleAlterTable() {
	users := q.MustNewTable[exampleUser]()

	sql := ddl.AlterTable(users).
		AddColumn(ddl.Column("display_name", ddl.Text()).NotNull().Default("")).
		SetDefaultExpr(users.CreatedAt, "now()").
		SetNotNull(users.Email).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	// ADD COLUMN "display_name" TEXT NOT NULL DEFAULT '';
	// ALTER TABLE "users"
	// ALTER COLUMN "created_at" SET DEFAULT now();
	// ALTER TABLE "users"
	// ALTER COLUMN "email" SET NOT NULL
}

func ExampleAlterTableStmt_RenameColumn() {
	users := q.MustNewTable[exampleUser]()

	sql := ddl.AlterTable(users).
		RenameColumn("name", "display_name").
		DropDefault(users.CreatedAt).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	// RENAME COLUMN "name" TO "display_name";
	// ALTER TABLE "users"
	// ALTER COLUMN "created_at" DROP DEFAULT
}

func ExampleAlterTableStmt_AddConstraint() {
	users := q.MustNewTable[exampleUser]()
	orgs := q.MustNewTable[exampleOrg]()

	sql := ddl.AlterTable(users).
		AddConstraint(
			ddl.ForeignKey(users.OrgID).
				References(orgs, orgs.ID).
				OnDelete(ddl.Cascade).
				Named("users_org_id_fk"),
		).
		AddConstraint(ddl.Unique(users.Email).Named("users_email_key")).
		MustRender(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// ALTER TABLE "users"
	// ADD CONSTRAINT "users_org_id_fk" FOREIGN KEY ("org_id") REFERENCES "orgs" ("id") ON DELETE CASCADE;
	// ALTER TABLE "users"
	// ADD CONSTRAINT "users_email_key" UNIQUE ("email")
}

func ExampleAlterTableStmt_Render_unsupportedFeature() {
	users := q.MustNewTable[exampleUser]()

	_, err := ddl.AlterTable(users).
		AlterColumnType(users.Email, ddl.Text()).
		Render(dialect.SQLite{})

	fmt.Println(err)

	// Output:
	// SQLite dialect does not support ALTER COLUMN TYPE
}
