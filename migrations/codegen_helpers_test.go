package migrations

import (
	"strings"
	"testing"

	"github.com/SennovE/qrafter/ddl"
)

func TestTypeCodeVariants(t *testing.T) {
	tests := []struct {
		name string
		typ  ddl.Type
		want string
	}{
		{name: "simple", typ: ddl.UUID(), want: `qddl.UUID()`},
		{name: "varchar", typ: ddl.SQLType("VARCHAR(42)"), want: `qddl.VarChar(42)`},
		{name: "char", typ: ddl.SQLType("CHAR(2)"), want: `qddl.Char(2)`},
		{name: "numeric precision", typ: ddl.SQLType("NUMERIC(12)"), want: `qddl.Numeric(12, 0)`},
		{name: "numeric scale", typ: ddl.SQLType("NUMERIC(12, 3)"), want: `qddl.Numeric(12, 3)`},
		{
			name: "custom dialects sorted",
			typ: ddl.SQLType("GEOMETRY").
				ForDialect("PostgreSQL", "geometry").
				ForDialect("MySQL", "GEOMETRY"),
			want: `qddl.SQLType("GEOMETRY").ForDialect("MySQL", "GEOMETRY").ForDialect("PostgreSQL", "geometry")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typeCode(tt.typ); got != tt.want {
				t.Fatalf("typeCode = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeCodeFallsBackForInvalidParameterizedTypes(t *testing.T) {
	for _, typ := range []ddl.Type{
		ddl.SQLType("VARCHAR(0)"),
		ddl.SQLType("CHAR(nope)"),
		ddl.SQLType("NUMERIC(0, 1)"),
		ddl.SQLType("NUMERIC(10, -1)"),
		ddl.SQLType("NUMERIC(10, 2, 1)"),
	} {
		got := typeCode(typ)
		if !strings.HasPrefix(got, `qddl.SQLType(`) {
			t.Fatalf("typeCode(%q) = %q, want SQLType fallback", typ.Name, got)
		}
	}
}

func TestExpressionCodeDetectsSimpleColumns(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{expr: "email", want: `qddl.Col("email")`},
		{expr: `"org_id"`, want: `qddl.Col("org_id")`},
		{expr: `"weird""name"`, want: `qddl.Col("weird\"name")`},
		{expr: "lower(email)", want: `qddl.RawExpr("lower(email)")`},
		{expr: `"unterminated`, want: `qddl.RawExpr("\"unterminated")`},
		{expr: "9lives", want: `qddl.RawExpr("9lives")`},
	}

	for _, tt := range tests {
		if got := rawExpressionCode(tt.expr); got != tt.want {
			t.Fatalf("rawExpressionCode(%q) = %q, want %q", tt.expr, got, tt.want)
		}
	}
}

func TestReferenceActionAndIndexMethodCode(t *testing.T) {
	for action, want := range map[ddl.ReferenceAction]string{
		ddl.Restrict:   "qddl.Restrict",
		ddl.Cascade:    "qddl.Cascade",
		ddl.SetNull:    "qddl.SetNull",
		ddl.SetDefault: "qddl.SetDefault",
		ddl.NoAction:   "qddl.NoAction",
		"custom":       `qddl.ReferenceAction("custom")`,
	} {
		if got := referenceActionCode(action); got != want {
			t.Fatalf("referenceActionCode(%q) = %q, want %q", action, got, want)
		}
	}

	for method, want := range map[ddl.IndexMethod]string{
		ddl.IndexBTree:       "qddl.IndexBTree",
		ddl.IndexHash:        "qddl.IndexHash",
		ddl.IndexGin:         "qddl.IndexGin",
		ddl.IndexGist:        "qddl.IndexGist",
		ddl.IndexSpGist:      "qddl.IndexSpGist",
		ddl.IndexBrin:        "qddl.IndexBrin",
		ddl.IndexFullText:    "qddl.IndexFullText",
		ddl.IndexSpatial:     "qddl.IndexSpatial",
		ddl.IndexColumnstore: "qddl.IndexColumnstore",
		ddl.IndexBitmap:      "qddl.IndexBitmap",
		"custom":             `qddl.IndexMethod("custom")`,
	} {
		if got := indexMethodCode(method); got != want {
			t.Fatalf("indexMethodCode(%q) = %q, want %q", method, got, want)
		}
	}
}
