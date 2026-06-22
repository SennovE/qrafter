package migrations

import (
	"strings"
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

func TestDiffSchemasAndMigrationStepsCoverTableChanges(t *testing.T) {
	current := Schema{Tables: []Table{
		{
			Name: "users",
			Columns: []Column{
				{Name: "id", Position: 1, Type: ddl.Integer(), NotNull: true},
				{Name: "email", Position: 2, Type: ddl.Text(), NotNull: true, HasDefault: true, DefaultExpr: "'old'"},
				{Name: "legacy", Position: 3, Type: ddl.Text()},
				{Name: "computed", Position: 4, Type: ddl.Text()},
			},
			Constraints: []Constraint{
				{Name: "pk_users_id", Kind: ConstraintPrimaryKey, Columns: []string{"id"}},
				{Name: "uq_users_email_old", Kind: ConstraintUnique, Columns: []string{"email"}},
				{Name: "chk_users_email_old", Kind: ConstraintCheck, CheckExpr: "email <> ''"},
				{
					Name:      "fk_users_org_old",
					Kind:      ConstraintForeignKey,
					Columns:   []string{"org_id"},
					Reference: Reference{TableName: "orgs", Columns: []string{"id"}},
				},
			},
			Indexes: []Index{
				{Name: "ix_users_email", TableName: "users", Keys: []IndexKey{{Expression: `"email"`}}},
				{Name: "ix_users_lower_email", TableName: "users", Keys: []IndexKey{{Expression: "lower(email)"}}},
			},
		},
		{
			Name:    "old_events",
			Columns: []Column{{Name: "id", Position: 1, Type: ddl.Integer()}},
			Indexes: []Index{{Name: "ix_old_events_id", TableName: "old_events", Keys: []IndexKey{{Expression: `"id"`}}}},
		},
	}}
	desired := Schema{Tables: []Table{
		{
			Name: "audit_log",
			Columns: []Column{
				{Name: "id", Position: 1, Type: ddl.BigSerial(), NotNull: true},
				{Name: "payload", Position: 2, Type: ddl.JSONB(), HasDefault: true, DefaultExpr: "'{}'::jsonb"},
			},
			Constraints: []Constraint{{Kind: ConstraintPrimaryKey, Columns: []string{"id"}}},
			Indexes: []Index{{
				Name:             "ix_audit_payload",
				TableName:        "audit_log",
				Method:           ddl.IndexGin,
				Keys:             []IndexKey{{Expression: `"payload"`}},
				NullsNotDistinct: true,
			}},
		},
		{
			Name: "users",
			Columns: []Column{
				{Name: "id", Position: 1, Type: ddl.BigInt(), NotNull: true},
				{Name: "email", Position: 2, Type: ddl.VarChar(255), NotNull: false, HasDefault: true, DefaultExpr: "''"},
				{Name: "nickname", Position: 3, Type: ddl.Text(), HasDefault: true, DefaultExpr: "''"},
				{Name: "computed", Position: 4, Type: ddl.Text(), Generated: ddl.GeneratedStored, GeneratedExpr: "lower(email)"},
			},
			Constraints: []Constraint{
				{Name: "pk_users_id", Kind: ConstraintPrimaryKey, Columns: []string{"id"}},
				{Name: "uq_users_email_new", Kind: ConstraintUnique, Columns: []string{"email"}},
				{Name: "chk_users_email_new", Kind: ConstraintCheck, CheckExpr: "length(email) > 3"},
				{
					Name:      "fk_users_org_new",
					Kind:      ConstraintForeignKey,
					Columns:   []string{"org_id"},
					Reference: Reference{TableName: "orgs", Columns: []string{"id"}},
					OnDelete:  ddl.Cascade,
					OnUpdate:  ddl.SetNull,
				},
			},
			Indexes: []Index{
				{
					Name:       "ix_users_email",
					TableName:  "users",
					Unique:     true,
					Method:     ddl.IndexBTree,
					Keys:       []IndexKey{{Expression: `"email"`}},
					Include:    []string{`"id"`},
					Predicate:  "email IS NOT NULL",
					Tablespace: "fastspace",
				},
				{Name: "ix_users_nickname", TableName: "users", Keys: []IndexKey{{Expression: `"nickname"`}}},
			},
		},
	}}

	diff := diffSchemas(current, desired)
	if diff.isEmpty() {
		t.Fatal("diff is empty")
	}
	if got, want := len(diff.AddedTables), 1; got != want {
		t.Fatalf("added tables = %d, want %d", got, want)
	}
	if got, want := len(diff.RemovedTables), 1; got != want {
		t.Fatalf("removed tables = %d, want %d", got, want)
	}
	if got, want := len(diff.ChangedTables), 1; got != want {
		t.Fatalf("changed tables = %d, want %d", got, want)
	}

	steps := migrationSteps(diff)
	if len(steps) < 10 {
		t.Fatalf("steps = %d, want many table-change steps", len(steps))
	}

	var sql strings.Builder
	for i := range steps {
		if steps[i].up != nil {
			rendered, err := ddl.Render(dialect.PostgreSQL{}, steps[i].up)
			if err != nil {
				t.Fatalf("render up step %d: %v", i, err)
			}
			sql.WriteString(rendered)
			sql.WriteByte('\n')
		}
		if steps[i].down != nil {
			if _, err := ddl.Render(dialect.PostgreSQL{}, steps[i].down); err != nil {
				t.Fatalf("render down step %d: %v", i, err)
			}
		}
	}

	gotSQL := sql.String()
	for _, want := range []string{
		`CREATE TABLE "audit_log"`,
		`DROP TABLE "old_events"`,
		`ALTER TABLE "users"`,
		`DROP INDEX "ix_users_lower_email"`,
		`CREATE UNIQUE INDEX "ix_users_email" ON "users" USING btree ("email") INCLUDE ("id") TABLESPACE "fastspace" WHERE email IS NOT NULL`,
	} {
		if !strings.Contains(gotSQL, want) {
			t.Fatalf("rendered SQL does not contain %q:\n%s", want, gotSQL)
		}
	}

	upCodes := strings.Join(upStepCodes(steps), "\n")
	downCodes := strings.Join(downStepCodes(steps), "\n")
	for _, want := range []string{
		`qddl.CreateTable("audit_log")`,
		`qddl.AlterTable("users").AlterColumnType("email", qddl.VarChar(255))`,
		`qddl.AlterTable("users").`,
		`qddl.KeyCol("payload")`,
		`qddl.ForeignKey("org_id").References("orgs", "id").OnDelete(qddl.Cascade).OnUpdate(qddl.SetNull)`,
	} {
		if !strings.Contains(upCodes, want) {
			t.Fatalf("up code does not contain %q:\n%s", want, upCodes)
		}
	}
	if !strings.Contains(downCodes, `qddl.CreateIndex("ix_old_events_id")`) {
		t.Fatalf("down code does not restore removed index:\n%s", downCodes)
	}
}

func TestColumnChangeSteps(t *testing.T) {
	tests := []struct {
		name string
		diff columnDiff
		want string
	}{
		{
			name: "type",
			diff: columnDiff{
				Current: Column{Name: "age", Type: ddl.Integer()},
				Desired: Column{Name: "age", Type: ddl.BigInt()},
			},
			want: `ALTER COLUMN "age" TYPE BIGINT`,
		},
		{
			name: "not null",
			diff: columnDiff{
				Current: Column{Name: "name", Type: ddl.Text()},
				Desired: Column{Name: "name", Type: ddl.Text(), NotNull: true},
			},
			want: `ALTER COLUMN "name" SET NOT NULL`,
		},
		{
			name: "drop default",
			diff: columnDiff{
				Current: Column{Name: "flag", Type: ddl.Boolean(), HasDefault: true, DefaultExpr: "true"},
				Desired: Column{Name: "flag", Type: ddl.Boolean()},
			},
			want: `ALTER COLUMN "flag" DROP DEFAULT`,
		},
		{
			name: "generated replacement",
			diff: columnDiff{
				Current: Column{Name: "slug", Type: ddl.Text()},
				Desired: Column{Name: "slug", Type: ddl.Text(), Generated: ddl.GeneratedStored, GeneratedExpr: "lower(name)"},
			},
			want: `ADD COLUMN "slug" TEXT GENERATED ALWAYS AS (lower(name)) STORED`,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			steps := appendColumnChangeSteps(nil, "users", &tt.diff)
			if len(steps) == 0 {
				t.Fatal("no steps generated")
			}
			rendered, err := ddl.Render(dialect.PostgreSQL{}, steps[0].up)
			if err != nil {
				t.Fatalf("render step: %v", err)
			}
			if !strings.Contains(rendered, tt.want) {
				t.Fatalf("rendered step = %q, want to contain %q", rendered, tt.want)
			}
		})
	}
}
