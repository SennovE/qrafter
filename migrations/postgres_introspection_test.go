package migrations

import (
	"context"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SennovE/qrafter/ddl"
)

func TestPostgreSQLReadSchema(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(150000))
	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{"schema", "name"}).
			AddRow("public", "users"))
	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{
			"schema", "table", "position", "name", "type", "not_null", "default", "identity", "generated",
		}).
			AddRow("public", "users", 1, "id", "integer", true, nil, "d", "").
			AddRow("public", "users", 2, "email", "character varying(255)", true, "'unknown'::character varying", "", "").
			AddRow("public", "users", 3, "full_name", "text", false, "first_name || ' ' || last_name", "", "s").
			AddRow("public", "missing", 1, "ignored", "text", false, nil, "", ""))
	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{
			"schema", "table", "name", "kind", "columns", "ref_schema", "ref_table", "ref_columns",
			"on_delete", "on_update", "check_expr",
		}).
			AddRow("public", "users", "pk_users_id", "p", "id", "", "", "", "a", "a", "").
			AddRow("public", "users", "uq_users_email", "u", "email", "", "", "", "a", "a", "").
			AddRow("public", "users", "chk_users_email", "c", "", "", "", "", "a", "a", "email <> ''").
			AddRow("public", "users", "fk_users_org", "f", "org_id", "public", "orgs", "id", "c", "r", ""))
	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{
			"schema", "table", "name", "method", "unique", "predicate", "tablespace", "nulls_not_distinct",
		}).
			AddRow("public", "users", "ix_users_email", "btree", true, "email IS NOT NULL", "fastspace", true).
			AddRow("public", "missing", "ix_missing", "btree", false, "", "", false))
	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{"schema", "table", "index", "position", "is_key", "expression"}).
			AddRow("public", "users", "ix_users_email", 1, true, `"email"`).
			AddRow("public", "users", "ix_users_email", 2, false, `"id"`).
			AddRow("public", "users", "ix_missing", 1, true, `"ignored"`))

	got, err := NewPostgreSQL(WithSchemas(" public ")).ReadSchema(context.Background(), db)
	if err != nil {
		t.Fatalf("ReadSchema error = %v", err)
	}

	want := Schema{Tables: []Table{{
		Schema: "public",
		Name:   "users",
		Columns: []Column{
			{
				Schema:       "public",
				TableName:    "users",
				Position:     1,
				Name:         "id",
				Type:         ddl.Integer(),
				DatabaseType: "integer",
				NotNull:      true,
				Identity:     ddl.IdentityByDefault,
			},
			{
				Schema:       "public",
				TableName:    "users",
				Position:     2,
				Name:         "email",
				Type:         ddl.VarChar(255),
				DatabaseType: "character varying(255)",
				NotNull:      true,
				HasDefault:   true,
				DefaultExpr:  "'unknown'::character varying",
			},
			{
				Schema:        "public",
				TableName:     "users",
				Position:      3,
				Name:          "full_name",
				Type:          ddl.Text(),
				DatabaseType:  "text",
				Generated:     ddl.GeneratedStored,
				GeneratedExpr: "first_name || ' ' || last_name",
			},
		},
		Constraints: []Constraint{
			{
				Schema: "public", TableName: "users", Name: "pk_users_id",
				Kind: ConstraintPrimaryKey, Columns: []string{"id"}, OnDelete: ddl.NoAction, OnUpdate: ddl.NoAction,
			},
			{
				Schema: "public", TableName: "users", Name: "uq_users_email",
				Kind: ConstraintUnique, Columns: []string{"email"}, OnDelete: ddl.NoAction, OnUpdate: ddl.NoAction,
			},
			{
				Schema: "public", TableName: "users", Name: "chk_users_email",
				Kind: ConstraintCheck, CheckExpr: "email <> ''", OnDelete: ddl.NoAction, OnUpdate: ddl.NoAction,
			},
			{
				Schema:    "public",
				TableName: "users",
				Name:      "fk_users_org",
				Kind:      ConstraintForeignKey,
				Columns:   []string{"org_id"},
				Reference: Reference{Schema: "public", TableName: "orgs", Columns: []string{"id"}},
				OnDelete:  ddl.Cascade,
				OnUpdate:  ddl.Restrict,
			},
		},
		Indexes: []Index{{
			Schema:           "public",
			TableSchema:      "public",
			TableName:        "users",
			Name:             "ix_users_email",
			Unique:           true,
			Method:           ddl.IndexBTree,
			Keys:             []IndexKey{{Expression: `"email"`}},
			Include:          []string{`"id"`},
			Predicate:        "email IS NOT NULL",
			Tablespace:       "fastspace",
			NullsNotDistinct: true,
		}},
	}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("schema = %#v\nwant %#v", got, want)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPostgreSQLHelpers(t *testing.T) {
	if got := postgresIdentityKind("a"); got != ddl.IdentityAlways {
		t.Fatalf("identity a = %q", got)
	}
	if got := postgresGeneratedKind("v"); got != ddl.GeneratedVirtual {
		t.Fatalf("generated v = %q", got)
	}
	if got := splitPostgresList("a\x1fb"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("splitPostgresList = %#v", got)
	}
	if got := splitPostgresList(""); got != nil {
		t.Fatalf("empty splitPostgresList = %#v, want nil", got)
	}
}
