package migrations

import (
	"context"
	"database/sql"
	"errors"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

var errUnexpectedMigrationQuery = errors.New("unexpected migration query")

func TestSchemaDiffDDLBuildsMigration(t *testing.T) {
	diff := migrationDiffFixture()
	migration := diff.DDL()

	upSQL, err := migration.Up.Render(dialect.PostgreSQL{})
	if err != nil {
		t.Fatalf("render up migration: %v", err)
	}
	downSQL, err := migration.Down.Render(dialect.PostgreSQL{})
	if err != nil {
		t.Fatalf("render down migration: %v", err)
	}

	assertContainsAll(t, upSQL, []string{
		`DROP INDEX "ix_users_legacy"`,
		`ALTER TABLE "users"` + "\n" + `    DROP CONSTRAINT "old_check"`,
		`DROP COLUMN "legacy"`,
		`ADD COLUMN "display_name" VARCHAR(120)`,
		`ALTER COLUMN "email" TYPE VARCHAR(320)`,
		`ALTER COLUMN "email" SET NOT NULL`,
		`ALTER COLUMN "email" SET DEFAULT lower('root@example.com')`,
		`ADD CONSTRAINT "ck_users_email" CHECK ("email" LIKE '%@%')`,
		`CREATE INDEX "ix_users_display" ON "users" ("display_name")`,
	})
	assertContainsAll(t, downSQL, []string{
		`DROP INDEX "ix_users_display"`,
		`ALTER TABLE "users"` + "\n" + `    DROP CONSTRAINT "ck_users_email"`,
		`DROP COLUMN "display_name"`,
		`ALTER COLUMN "email" TYPE TEXT`,
		`ALTER COLUMN "email" DROP NOT NULL`,
		`ALTER COLUMN "email" DROP DEFAULT`,
		`ADD CONSTRAINT "old_check" CHECK (legacy <> '')`,
		`CREATE INDEX "ix_users_legacy" ON "users" (legacy)`,
	})
}

func TestGenerateMigrationCodeFromDiff(t *testing.T) {
	code, err := GenerateMigrationCodeFromDiff(migrationDiffFixture(), MigrationCodeOptions{
		PackageName:  "dbmigrations",
		UpFuncName:   "Apply",
		DownFuncName: "Rollback",
	})
	if err != nil {
		t.Fatalf("generate migration code: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "generated.go", code, parser.AllErrors); err != nil {
		t.Fatalf("generated code does not parse: %v\n%s", err, code)
	}

	source := string(code)
	assertContainsAll(t, source, []string{
		`package dbmigrations`,
		`func Apply() qddl.Statements`,
		`func Rollback() qddl.Statements`,
		`qddl.AlterTable("users").AddColumn(qddl.Column("display_name", qddl.SQLType("VARCHAR(120)"))`,
		`qddl.AlterTable("users").DropConstraint("old_check")`,
		`qddl.CreateIndex("ix_users_display").On("users", qddl.Key(qddl.RawExpr("\"display_name\"")))`,
		`qddl.DropIndex("ix_users_display")`,
	})
}

func TestGenerateMigrationCodeForTableConfigReadsSchema(t *testing.T) {
	introspector := &staticSchemaIntrospector{}
	result, err := GenerateMigrationCodeForTableConfig[configSchemaUser](
		context.Background(),
		&MigrationToolConfig{
			DB:           fakeMigrationDatabase{},
			Introspector: introspector,
			Dialect:      dialect.PostgreSQL{},
			Code: MigrationCodeOptions{
				PackageName: "dbmigrations",
			},
		},
	)
	if err != nil {
		t.Fatalf("generate migration code from table config: %v", err)
	}
	if !introspector.called {
		t.Fatal("introspector was not called")
	}
	if len(result.Diff.AddedTables) != 1 || result.Diff.AddedTables[0].Name != "users" {
		t.Fatalf("added tables = %#v", result.Diff.AddedTables)
	}
	assertContainsAll(t, string(result.Code), []string{
		`package dbmigrations`,
		`qddl.CreateTable("users")`,
		`qddl.CreateIndex("ix_users_email")`,
	})
}

func migrationDiffFixture() SchemaDiff {
	current := Schema{Tables: []Table{{
		Schema: "public",
		Name:   "users",
		Columns: []Column{
			{Name: "id", Type: ddl.BigInt(), NotNull: true},
			{Name: "email", Type: ddl.Text()},
			{Name: "legacy", Type: ddl.Text(), HasDefault: true, DefaultExpr: "'legacy'"},
		},
		Constraints: []Constraint{
			{Name: "users_pkey", Kind: ConstraintPrimaryKey, Columns: []string{"id"}},
			{Name: "old_check", Kind: ConstraintCheck, CheckExpr: "legacy <> ''"},
		},
		Indexes: []Index{
			{Name: "ix_users_legacy", TableName: "users", Keys: []IndexKey{{Expression: "legacy"}}},
			{Name: "ix_users_email", TableName: "users", Keys: []IndexKey{{Expression: "email"}}},
		},
	}}}
	desired := Schema{Tables: []Table{{
		Schema: "public",
		Name:   "users",
		Columns: []Column{
			{Name: "id", Type: ddl.BigInt(), NotNull: true},
			{
				Name:         "email",
				Type:         ddl.VarChar(320),
				NotNull:      true,
				HasDefault:   true,
				DefaultExpr:  "lower('root@example.com')",
				DatabaseType: "VARCHAR(320)",
			},
			{Name: "display_name", Type: ddl.VarChar(120)},
		},
		Constraints: []Constraint{
			{Name: "pk_users", Kind: ConstraintPrimaryKey, Columns: []string{"id"}},
			{Name: "ck_users_email", Kind: ConstraintCheck, CheckExpr: `"email" LIKE '%@%'`},
		},
		Indexes: []Index{
			{
				Name:      "ix_users_email",
				TableName: "users",
				Keys:      []IndexKey{{Expression: `"email"`}},
				Predicate: "deleted_at IS NULL",
			},
			{Name: "ix_users_display", TableName: "users", Keys: []IndexKey{{Expression: `"display_name"`}}},
		},
	}}}
	return DiffSchemas(current, desired)
}

func assertContainsAll(t *testing.T, value string, fragments []string) {
	t.Helper()
	for i := range fragments {
		if !strings.Contains(value, fragments[i]) {
			t.Fatalf("%q not found in:\n%s", fragments[i], value)
		}
	}
}

type staticSchemaIntrospector struct {
	schema Schema
	called bool
}

func (s *staticSchemaIntrospector) ReadSchema(_ context.Context, db Database) (Schema, error) {
	if db == nil {
		return Schema{}, errors.New("db is nil")
	}
	s.called = true
	return s.schema, nil
}

type fakeMigrationDatabase struct{}

func (fakeMigrationDatabase) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	return nil, errUnexpectedMigrationQuery
}

func (fakeMigrationDatabase) QueryRowContext(_ context.Context, _ string, _ ...any) *sql.Row {
	return &sql.Row{}
}
