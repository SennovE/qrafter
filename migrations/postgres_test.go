package migrations

import (
	"reflect"
	"testing"
)

func TestPostgreSQLOptionsDefaultToPublicSchema(t *testing.T) {
	options := NewPostgreSQL().effectiveOptions()

	if options.allSchemas {
		t.Fatal("allSchemas = true, want false")
	}
	if !reflect.DeepEqual(options.schemas, []string{defaultPostgreSQLSchema}) {
		t.Fatalf("schemas = %#v, want public", options.schemas)
	}
}

func TestPostgreSQLOptionsNormalizeConfiguredSchemas(t *testing.T) {
	options := NewPostgreSQL(WithSchemas(" public ", "", "tenant", "tenant")).effectiveOptions()

	if options.allSchemas {
		t.Fatal("allSchemas = true, want false")
	}
	if !reflect.DeepEqual(options.schemas, []string{"public", "tenant"}) {
		t.Fatalf("schemas = %#v, want public and tenant", options.schemas)
	}
}

func TestPostgreSQLOptionsCanReadAllSchemas(t *testing.T) {
	options := NewPostgreSQL(WithAllSchemas()).effectiveOptions()

	if !options.allSchemas {
		t.Fatal("allSchemas = false, want true")
	}
	if len(options.schemas) != 0 {
		t.Fatalf("schemas = %#v, want empty when reading all schemas", options.schemas)
	}
}

func TestPostgreSQLSchemaPredicate(t *testing.T) {
	predicate, args := postgreSQLSchemaPredicate("n", postgreSQLOptions{schemas: []string{"public", "tenant"}})
	if predicate != "n.nspname IN ($1, $2)" {
		t.Fatalf("predicate = %q", predicate)
	}
	if !reflect.DeepEqual(args, []any{"public", "tenant"}) {
		t.Fatalf("args = %#v", args)
	}

	predicate, args = postgreSQLSchemaPredicate("n", postgreSQLOptions{allSchemas: true})
	if predicate != "n.nspname <> 'information_schema' AND pg_catalog.left(n.nspname, 3) <> 'pg_'" {
		t.Fatalf("all schemas predicate = %q", predicate)
	}
	if len(args) != 0 {
		t.Fatalf("all schemas args = %#v, want empty", args)
	}
}
