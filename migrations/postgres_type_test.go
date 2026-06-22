package migrations

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SennovE/qrafter/ddl"
)

func TestPostgreSQLTypeParsesKnownTypes(t *testing.T) {
	tests := []struct {
		name string
		want ddl.Type
	}{
		{name: " smallint ", want: ddl.SmallInt()},
		{name: "INT4", want: ddl.Integer()},
		{name: "bigint", want: ddl.BigInt()},
		{name: "time without time zone", want: ddl.Time()},
		{name: "timestamp   with   time   zone", want: ddl.TimestampTZ()},
		{name: "varchar(64)", want: ddl.VarChar(64)},
		{name: "character varying(128)", want: ddl.VarChar(128)},
		{name: "char(2)", want: ddl.Char(2)},
		{name: "character(3)", want: ddl.Char(3)},
		{name: "numeric(12)", want: ddl.Numeric(12, 0)},
		{name: "numeric(12, 4)", want: ddl.Numeric(12, 4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := postgreSQLType(tt.name)
			if !ok {
				t.Fatalf("postgreSQLType(%q) ok = false", tt.name)
			}
			if !typesEqual(got, tt.want) {
				t.Fatalf("postgreSQLType(%q) = %#v, want %#v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPostgreSQLTypeFallsBackForUnknownOrInvalidTypes(t *testing.T) {
	for _, name := range []string{
		"custom_type",
		"varchar(0)",
		"char(nope)",
		"numeric(0)",
		"numeric(10, -1)",
		"numeric(10, 2, 1)",
	} {
		got, ok := postgreSQLType(name)
		if ok {
			t.Fatalf("postgreSQLType(%q) ok = true", name)
		}
		if got.Name != name {
			t.Fatalf("postgreSQLType(%q) fallback = %#v", name, got)
		}
	}
}

func TestPostgreSQLLegacyVersionBranches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	options := postgreSQLOptions{schemas: []string{"public"}}
	tables := map[tableKey]*Table{{schema: "public", table: "users"}: {Schema: "public", Name: "users"}}
	indexes := map[indexKey]*Index{{schema: "public", table: "users", index: "ix_users_id"}: {
		Schema: "public", TableName: "users", Name: "ix_users_id",
	}}
	ctx := context.Background()

	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{
			"schema", "table", "position", "name", "type", "not_null", "default", "identity", "generated",
		}).
			AddRow("public", "users", 1, "id", "int4", true, nil, "", ""))
	if err := readPostgreSQLColumns(ctx, db, options, tables, 110000); err != nil {
		t.Fatalf("read legacy columns: %v", err)
	}

	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{
			"schema", "table", "name", "method", "unique", "predicate", "tablespace", "nulls_not_distinct",
		}).
			AddRow("public", "users", "ix_users_id", "btree", false, "", "", false))
	if _, err := readPostgreSQLIndexMetadata(ctx, db, options, tables, 140000); err != nil {
		t.Fatalf("read legacy index metadata: %v", err)
	}

	mock.ExpectQuery(".*").
		WithArgs("public").
		WillReturnRows(sqlmock.NewRows([]string{"schema", "table", "index", "position", "is_key", "expression"}).
			AddRow("public", "users", "ix_users_id", 1, true, `"id"`))
	if err := readPostgreSQLIndexKeys(ctx, db, options, indexes, 100000); err != nil {
		t.Fatalf("read legacy index keys: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
