package ddl

import (
	"testing"

	"github.com/SennovE/qrafter/dialect"
)

func TestColumnPredicateRendersColumnAsIdentifier(t *testing.T) {
	sql, err := Render(dialect.PostgreSQL{}, Col("age").Ge(0))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := sql, `"age" >= 0`; got != want {
		t.Fatalf("rendered predicate = %q, want %q", got, want)
	}
}

func TestColumnPredicateKeepsStringValuesAsLiterals(t *testing.T) {
	sql, err := Render(dialect.PostgreSQL{}, Col("status").Eq("active"))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := sql, `"status" = 'active'`; got != want {
		t.Fatalf("rendered predicate = %q, want %q", got, want)
	}
}

func TestCheckAcceptsRootPredicate(_ *testing.T) {
	_ = Check(Literal(1).Eq(1))
}
