package migrations

import "testing"

func TestWithoutMigrationVersionTableRemovesConfiguredTable(t *testing.T) {
	schema := Schema{Tables: []Table{
		{Name: "users"},
		{Name: "schema_version"},
		{Name: "orders"},
	}}

	got := withoutMigrationVersionTable(schema, "schema_version")
	if len(got.Tables) != 2 {
		t.Fatalf("tables = %#v, want users and orders", got.Tables)
	}
	if got.Tables[0].Name != "users" || got.Tables[1].Name != "orders" {
		t.Fatalf("tables = %#v, want users and orders", got.Tables)
	}
}

func TestMigrationVersionTableDefaultsWhenBlank(t *testing.T) {
	if got := migrationVersionTable(""); got != DefaultMigrationVersionTable {
		t.Fatalf("default version table = %q, want %q", got, DefaultMigrationVersionTable)
	}
	if got := migrationVersionTable(" custom_versions "); got != "custom_versions" {
		t.Fatalf("custom version table = %q", got)
	}
}
