package ddl

import (
	"strings"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

func TestColumnPredicateRendersColumnAsIdentifier(t *testing.T) {
	var w strings.Builder

	Col("age").Ge(0).Render(&w, dialect.PostgreSQL{})

	if got, want := w.String(), `"age" >= 0`; got != want {
		t.Fatalf("rendered predicate = %q, want %q", got, want)
	}
}

func TestColumnPredicateKeepsStringValuesAsLiterals(t *testing.T) {
	var w strings.Builder

	Col("status").Eq("active").Render(&w, dialect.PostgreSQL{})

	if got, want := w.String(), `"status" = 'active'`; got != want {
		t.Fatalf("rendered predicate = %q, want %q", got, want)
	}
}

func TestCheckAcceptsRootPredicate(_ *testing.T) {
	_ = Check(q.Literal(1).Eq(1))
}
