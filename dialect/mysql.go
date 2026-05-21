package dialect

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/internal/utils"
)

const (
	mysqlOffsetOnlyLimit = "18446744073709551615"
	mysqlDialectName     = "MySQL"
)

// MySQL renders qrafter queries using MySQL syntax.
//
// It renders MySQL-specific forms for empty INSERT rows, multi-table UPDATE,
// multi-table DELETE, and NULL ordering. Unsupported MySQL features, such as
// RETURNING and FULL JOIN, fail fast with UnsupportedFeatureError.
type MySQL struct {
	BaseDialect
}

// QuoteIdent renders a MySQL backtick-quoted identifier.
func (MySQL) QuoteIdent(ident string) string {
	return utils.QuoteWith(ident, "`")
}

// LimitOffset renders MySQL LIMIT/OFFSET clauses.
func (MySQL) LimitOffset(limit, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d, %d", offset, limit)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf("LIMIT %s OFFSET %d", mysqlOffsetOnlyLimit, offset)
	default:
		return ""
	}
}

// RenderDefaultValues renders MySQL's empty-row INSERT syntax.
func (MySQL) RenderDefaultValues(w *strings.Builder) {
	w.WriteString(" ()\nVALUES ()")
}

// RenderReturning rejects RETURNING because MySQL does not support qrafter's
// PostgreSQL-style RETURNING clause.
func (MySQL) RenderReturning(_ *strings.Builder, _ func()) {
	panic(UnsupportedFeatureError{Dialect: mysqlDialectName, Feature: "RETURNING"})
}

// RenderOrder renders NULLS FIRST/LAST through MySQL-compatible expressions.
func (MySQL) RenderOrder(w *strings.Builder, renderExpr func(), direction, nulls string) {
	if nulls == "" {
		renderOrderDefault(w, renderExpr, direction, nulls)
		return
	}

	renderExpr()
	if strings.EqualFold(nulls, "FIRST") {
		w.WriteString(" IS NOT NULL")
	} else {
		w.WriteString(" IS NULL")
	}
	w.WriteString(", ")
	renderExpr()
	if direction != "" {
		w.WriteString(" ")
		w.WriteString(direction)
	}
}

// RenderJoin rejects FULL JOIN because MySQL has no native FULL JOIN syntax.
func (MySQL) RenderJoin(w *strings.Builder, joinType string, renderTable, renderPredicates func()) {
	if strings.EqualFold(joinType, "FULL JOIN") {
		panic(UnsupportedFeatureError{Dialect: mysqlDialectName, Feature: "FULL JOIN"})
	}
	renderJoinDefault(w, joinType, renderTable, renderPredicates)
}

// RenderUpdateTarget renders MySQL multi-table UPDATE syntax.
func (MySQL) RenderUpdateTarget(w *strings.Builder, renderTarget func(), hasFrom bool, renderFrom func()) {
	w.WriteString("UPDATE ")
	renderTarget()
	if hasFrom {
		w.WriteString(", ")
		renderFrom()
	}
}

// RenderUpdateFrom is a no-op because MySQL renders UPDATE source tables in the
// UPDATE table reference list.
func (MySQL) RenderUpdateFrom(_ *strings.Builder, _ func()) {}

// RenderDeleteTarget renders MySQL multi-table DELETE syntax.
func (MySQL) RenderDeleteTarget(
	w *strings.Builder,
	renderTarget func(),
	renderTargetName func(),
	hasUsing bool,
	renderUsing func(),
) {
	if !hasUsing {
		w.WriteString("DELETE FROM ")
		renderTarget()
		return
	}

	w.WriteString("DELETE ")
	renderTargetName()
	w.WriteString("\nFROM ")
	renderTarget()
	w.WriteString(", ")
	renderUsing()
}

// RenderDeleteUsing is a no-op because MySQL renders DELETE source tables in
// the DELETE table reference list.
func (MySQL) RenderDeleteUsing(_ *strings.Builder, _ func()) {}
