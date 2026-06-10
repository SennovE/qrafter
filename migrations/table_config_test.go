package migrations

import (
	"reflect"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

type configSchemaUser struct {
	q.Table `table:"users"`

	ID          q.Column[int64] `q:"pk"`
	OrgID       q.Column[int64]
	Email       q.Column[string] `q:"uq"`
	DisplayName q.Column[*string]
	DeletedAt   q.Column[*string]
}

func (u configSchemaUser) TableConfig() q.TableConfig { //nolint:gocritic // q.NewTable currently binds value table models.
	return q.TableConfig{
		Name:   "users",
		Schema: "public",
		Columns: q.ColumnsConfig{
			u.Email.DDLKey(): {
				Type:    ddl.VarChar(320),
				NotNull: true,
				Unique:  true,
				Default: ddl.RawExpr("lower('root@example.com')"),
			},
			u.DisplayName.DDLKey(): {
				Type: ddl.VarChar(120),
			},
		},
		Constraints: q.ConstraintsConfig{
			ddl.PrimaryKey(u.ID.Name()).Named("pk_users"),
			ddl.Check(ddl.Col(u.Email.Name()).Like("%@%")).Named("ck_users_email"),
			ddl.ForeignKey(u.OrgID.Name()).
				References("orgs", "id").
				OnDelete(ddl.Cascade).
				Named("fk_users_org"),
		},
		Indexes: q.IndexesConfig{
			ddl.IndexCols("ix_users_email", u.Email.Name()).
				Where(ddl.Col(u.DeletedAt.Name()).IsNull()),
			ddl.Index(
				"ix_users_org_email",
				ddl.KeyCol(u.OrgID.Name()).Desc().NullsLast(),
				ddl.Key(ddl.Func("lower", ddl.Col(u.Email.Name()))),
			).Using(ddl.IndexBTree),
		},
	}
}

func TestTableConfigToSchemaTableUsesFullConfig(t *testing.T) {
	schema := TableConfigToSchema[configSchemaUser](dialect.PostgreSQL{})
	table := schema.Tables[0]

	if table.Schema != "public" || table.Name != "users" {
		t.Fatalf("table = %s.%s, want public.users", table.Schema, table.Name)
	}

	email := findColumn(t, &table, "email")
	if email.Type.Name != "VARCHAR(320)" || !email.NotNull || !email.HasDefault {
		t.Fatalf("email column = %#v", email)
	}
	if email.DefaultExpr != "lower('root@example.com')" {
		t.Fatalf("email default = %q", email.DefaultExpr)
	}

	displayName := findColumn(t, &table, "display_name")
	if displayName.Type.Name != "VARCHAR(120)" || displayName.NotNull {
		t.Fatalf("display_name column = %#v", displayName)
	}

	pk := findConstraint(t, &table, ConstraintPrimaryKey, "pk_users")
	if pk.Schema != "public" || pk.TableName != "users" || len(pk.Columns) != 1 || pk.Columns[0] != "id" {
		t.Fatalf("primary key = %#v", pk)
	}

	unique := findConstraint(t, &table, ConstraintUnique, "")
	if unique.Schema != "public" || unique.TableName != "users" || len(unique.Columns) != 1 || unique.Columns[0] != "email" {
		t.Fatalf("unique constraint = %#v", unique)
	}

	check := findConstraint(t, &table, ConstraintCheck, "ck_users_email")
	if normalizeSQL(check.CheckExpr) != "email like '%@%'" {
		t.Fatalf("check expr = %q", check.CheckExpr)
	}

	fk := findConstraint(t, &table, ConstraintForeignKey, "fk_users_org")
	if fk.Reference.TableName != "orgs" || len(fk.Reference.Columns) != 1 || fk.Reference.Columns[0] != "id" {
		t.Fatalf("foreign key reference = %#v", fk.Reference)
	}
	if fk.OnDelete != ddl.Cascade || fk.OnUpdate != ddl.NoAction {
		t.Fatalf("foreign key actions = delete %q update %q", fk.OnDelete, fk.OnUpdate)
	}

	emailIndex := findIndex(t, &table, "ix_users_email")
	if emailIndex.TableName != "users" ||
		len(emailIndex.Keys) != 1 ||
		emailIndex.Keys[0].Expression != `"email"` ||
		normalizeSQL(emailIndex.Predicate) != "deleted_at is null" {
		t.Fatalf("email index = %#v", emailIndex)
	}

	orgEmailIndex := findIndex(t, &table, "ix_users_org_email")
	if orgEmailIndex.Method != ddl.IndexBTree ||
		len(orgEmailIndex.Keys) != 2 ||
		orgEmailIndex.Keys[0].Expression != `"org_id" DESC NULLS LAST` ||
		orgEmailIndex.Keys[1].Expression != `lower("email")` {
		t.Fatalf("org/email index = %#v", orgEmailIndex)
	}
}

func TestDiffSchemas(t *testing.T) {
	current := Schema{Tables: []Table{
		{
			Schema: "public",
			Name:   "legacy",
			Columns: []Column{
				{Name: "id", Type: ddl.BigInt(), NotNull: true},
			},
		},
		{
			Schema: "public",
			Name:   "users",
			Columns: []Column{
				{Name: "id", Type: ddl.BigInt(), NotNull: true},
				{Name: "org_id", Type: ddl.BigInt(), NotNull: true},
				{Name: "email", Type: ddl.Text()},
				{Name: "deleted_at", Type: ddl.Text()},
				{Name: "old", Type: ddl.Text()},
			},
			Constraints: []Constraint{
				{Name: "users_pkey", Kind: ConstraintPrimaryKey, Columns: []string{"id"}},
				{Name: "users_email_key", Kind: ConstraintUnique, Columns: []string{"email"}},
				{Name: "old_check", Kind: ConstraintCheck, CheckExpr: "old <> ''"},
			},
			Indexes: []Index{
				{Name: "ix_old", TableName: "users", Keys: []IndexKey{{Expression: "old"}}},
				{Name: "ix_users_email", TableName: "users", Keys: []IndexKey{{Expression: "email"}}},
			},
		},
	}}
	desired := TableConfigToSchema[configSchemaUser](dialect.PostgreSQL{})
	desired.Tables = append(desired.Tables, Table{
		Schema: "public",
		Name:   "audit",
		Columns: []Column{
			{Name: "id", Type: ddl.BigInt(), NotNull: true},
		},
	})

	diff := DiffSchemas(current, desired)
	if diff.IsEmpty() {
		t.Fatal("diff is empty")
	}
	if len(diff.AddedTables) != 1 || diff.AddedTables[0].Name != "audit" {
		t.Fatalf("added tables = %#v", diff.AddedTables)
	}
	if len(diff.RemovedTables) != 1 || diff.RemovedTables[0].Name != "legacy" {
		t.Fatalf("removed tables = %#v", diff.RemovedTables)
	}
	if len(diff.ChangedTables) != 1 {
		t.Fatalf("changed tables = %#v", diff.ChangedTables)
	}

	users := diff.ChangedTables[0]
	assertNames(t, columnsToNames(users.AddedColumns), []string{"display_name"})
	assertNames(t, columnsToNames(users.RemovedColumns), []string{"old"})
	assertNames(t, columnDiffDesiredNames(users.ChangedColumns), []string{"email"})

	assertNames(t, constraintsToNames(users.AddedConstraints), []string{"ck_users_email", "fk_users_org"})
	assertNames(t, constraintsToNames(users.RemovedConstraints), []string{"old_check"})
	assertNames(t, constraintDiffDesiredNames(users.ChangedConstraints), []string{"pk_users"})

	assertNames(t, indexesToNames(users.AddedIndexes), []string{"ix_users_org_email"})
	assertNames(t, indexesToNames(users.RemovedIndexes), []string{"ix_old"})
	assertNames(t, indexDiffDesiredNames(users.ChangedIndexes), []string{"ix_users_email"})
}

func findColumn(t *testing.T, table *Table, name string) Column {
	t.Helper()
	for i := range table.Columns {
		if table.Columns[i].Name == name {
			return table.Columns[i]
		}
	}
	t.Fatalf("column %q not found in %#v", name, table.Columns)
	return Column{}
}

func findConstraint(t *testing.T, table *Table, kind ConstraintKind, name string) Constraint {
	t.Helper()
	for i := range table.Constraints {
		if table.Constraints[i].Kind == kind && table.Constraints[i].Name == name {
			return table.Constraints[i]
		}
	}
	t.Fatalf("constraint %s/%q not found in %#v", kind, name, table.Constraints)
	return Constraint{}
}

func findIndex(t *testing.T, table *Table, name string) Index {
	t.Helper()
	for i := range table.Indexes {
		if table.Indexes[i].Name == name {
			return table.Indexes[i]
		}
	}
	t.Fatalf("index %q not found in %#v", name, table.Indexes)
	return Index{}
}

func assertNames(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("names = %#v, want %#v", got, want)
	}
}

func columnsToNames(columns []Column) []string {
	out := make([]string, len(columns))
	for i := range columns {
		out[i] = columns[i].Name
	}
	return out
}

func columnDiffDesiredNames(diffs []ColumnDiff) []string {
	out := make([]string, len(diffs))
	for i := range diffs {
		out[i] = diffs[i].Desired.Name
	}
	return out
}

func constraintsToNames(constraints []Constraint) []string {
	out := make([]string, len(constraints))
	for i := range constraints {
		out[i] = constraints[i].Name
	}
	return out
}

func constraintDiffDesiredNames(diffs []ConstraintDiff) []string {
	out := make([]string, len(diffs))
	for i := range diffs {
		out[i] = diffs[i].Desired.Name
	}
	return out
}

func indexesToNames(indexes []Index) []string {
	out := make([]string, len(indexes))
	for i := range indexes {
		out[i] = indexes[i].Name
	}
	return out
}

func indexDiffDesiredNames(diffs []IndexDiff) []string {
	out := make([]string, len(diffs))
	for i := range diffs {
		out[i] = diffs[i].Desired.Name
	}
	return out
}
