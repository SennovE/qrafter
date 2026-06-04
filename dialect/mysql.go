package dialect

import (
	"strings"

	"github.com/SennovE/qrafter/internal/utils"
)

const (
	mysqlOffsetOnlyLimit = "18446744073709551615"
	mysqlDialectName     = "MySQL"
	mysqlPartialIndex    = "PARTIAL INDEX"
	alterColumnNullable  = "ALTER COLUMN NULLABILITY"
)

// MySQL renders qrafter queries using MySQL syntax.
//
// It renders MySQL-specific forms for empty INSERT rows, multi-table UPDATE,
// multi-table DELETE, and NULL ordering. Unsupported MySQL features, such as
// RETURNING and FULL JOIN, fail fast with UnsupportedFeatureError.
type MySQL struct {
	BaseDialect
}

// DialectName returns the dialect name.
func (MySQL) DialectName() string {
	return mysqlDialectName
}

// QuoteIdent renders a MySQL backtick-quoted identifier.
func (MySQL) QuoteIdent(ident string) string {
	return utils.QuoteWith(ident, "`")
}

// CompileNode renders MySQL-specific compiler nodes.
func (MySQL) CompileNode(c Compiler, node any) bool {
	return compileMySQLDML(c, node) || compileMySQLDDL(c, node)
}

func compileMySQLDML(c Compiler, node any) bool {
	switch n := node.(type) {
	case DefaultValues:
		c.Write(" ()\nVALUES ()")
		return true
	case Returning:
		c.Unsupported("RETURNING")
		return true
	case OrderItem:
		return compileMySQLOrder(c, n)
	case Join:
		if strings.EqualFold(n.Type, "FULL JOIN") {
			c.Unsupported("FULL JOIN")
			return true
		}
		return false
	case UpdateTarget:
		c.Write("UPDATE ")
		c.Compile(n.Target)
		if len(n.From) > 0 {
			c.Write(", ")
			c.CompileList(n.From, ", ")
		}
		return true
	case UpdateFrom:
		return len(n.From) > 0
	case DeleteTarget:
		if len(n.Using) == 0 {
			return false
		}
		c.Write("DELETE ")
		c.Write(c.Renderer().QuoteIdent(n.TargetName))
		c.Write("\nFROM ")
		c.Compile(n.Target)
		c.Write(", ")
		c.CompileList(n.Using, ", ")
		return true
	case DeleteUsing:
		return len(n.Using) > 0
	default:
		return false
	}
}

func compileMySQLDDL(c Compiler, node any) bool {
	switch n := node.(type) {
	case PartialIndexPredicate:
		c.Unsupported(mysqlPartialIndex)
		return true
	case AlterIndexRename:
		if !n.HasTable {
			panic("MySQL requires table name")
		}
		c.Write("ALTER TABLE ")
		c.Write(c.Renderer().QuoteIdent(n.Table))
		c.Write(" RENAME INDEX ")
		c.Write(c.Renderer().QuoteIdent(n.OldName))
		c.Write(" TO ")
		c.Write(c.Renderer().QuoteIdent(n.NewName))
		return true
	case AlterColumnType:
		c.Write("MODIFY COLUMN ")
		c.Write(c.Renderer().QuoteIdent(n.Column))
		c.Write(" ")
		c.Write(n.Type)
		return true
	case AlterColumnNullability:
		c.Unsupported(alterColumnNullable)
		return true
	case AlterTableDropConstraint:
		c.Write("DROP ")
		c.Write(c.Renderer().QuoteIdent(n.Name))
		return true
	case LimitOffset:
		return compileMySQLLimitOffset(c, n)
	default:
		return false
	}
}

func compileMySQLOrder(c Compiler, n OrderItem) bool {
	if n.Nulls == "" {
		return false
	}

	c.Compile(n.Expr)
	if strings.EqualFold(n.Nulls, "FIRST") {
		c.Write(" IS NOT NULL")
	} else {
		c.Write(" IS NULL")
	}
	c.Write(", ")
	c.Compile(n.Expr)
	if n.Direction != "" {
		c.Write(" ")
		c.Write(n.Direction)
	}
	return true
}

func compileMySQLLimitOffset(c Compiler, n LimitOffset) bool {
	switch {
	case n.Limit > 0 && n.Offset > 0:
		c.Write("\nLIMIT ")
		c.WriteInt(n.Offset)
		c.Write(", ")
		c.WriteInt(n.Limit)
	case n.Limit > 0:
		c.Write("\nLIMIT ")
		c.WriteInt(n.Limit)
	case n.Offset > 0:
		c.Write("\nLIMIT ")
		c.Write(mysqlOffsetOnlyLimit)
		c.Write(" OFFSET ")
		c.WriteInt(n.Offset)
	default:
		return false
	}
	return true
}
