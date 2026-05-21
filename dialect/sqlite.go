package dialect

import (
	"fmt"
	"strings"
)

const (
	sqliteDeleteUsingFeature = "DELETE USING"
	sqliteDialectName        = "SQLite"
)

// SQLite renders qrafter queries using SQLite placeholder and LIMIT/OFFSET
// syntax. Unsupported SQLite features, such as DELETE USING, fail fast with
// UnsupportedFeatureError.
type SQLite struct {
	BaseDialect
}

// Literal renders SQLite-friendly inline SQL literals.
func (SQLite) Literal(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		return BaseDialect{}.Literal(v)
	}
}

// RenderDeleteTarget renders SQLite DELETE syntax.
func (SQLite) RenderDeleteTarget(
	w *strings.Builder,
	renderTarget func(),
	_ func(),
	hasUsing bool,
	_ func(),
) {
	if hasUsing {
		panic(UnsupportedFeatureError{Dialect: sqliteDialectName, Feature: sqliteDeleteUsingFeature})
	}

	w.WriteString("DELETE FROM ")
	renderTarget()
}

// RenderDeleteUsing rejects DELETE USING because SQLite has no native USING
// clause for DELETE statements.
func (SQLite) RenderDeleteUsing(_ *strings.Builder, _ func()) {
	panic(UnsupportedFeatureError{Dialect: sqliteDialectName, Feature: sqliteDeleteUsingFeature})
}

// LimitOffset renders SQLite LIMIT/OFFSET clauses.
func (SQLite) LimitOffset(limit, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf("LIMIT -1 OFFSET %d", offset)
	default:
		return ""
	}
}
