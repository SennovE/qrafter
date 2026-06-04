package dialect

import (
	"fmt"
	"strings"
)

const sqlServerDialectName = "SQL Server"

// SQLServer renders qrafter queries using Microsoft SQL Server syntax.
type SQLServer struct {
	BaseDialect
}

// DialectName returns the dialect name.
func (SQLServer) DialectName() string {
	return sqlServerDialectName
}

// QuoteIdent renders a SQL Server bracket-quoted identifier.
func (SQLServer) QuoteIdent(ident string) string {
	return "[" + strings.ReplaceAll(ident, "]", "]]") + "]"
}

// Literal renders SQL Server-friendly inline SQL literals.
func (SQLServer) Literal(value any) string {
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

// Placeholder renders a SQL Server named placeholder.
func (SQLServer) Placeholder(position int) string {
	return fmt.Sprintf("@p%d", position)
}

// CompileNode renders SQL Server-specific compiler nodes.
func (SQLServer) CompileNode(c Compiler, node any) bool {
	switch n := node.(type) {
	case Returning:
		c.Unsupported("RETURNING")
		return true
	case OrderItem:
		return compileSQLServerOrder(c, n)
	case DropTableBehavior:
		if n.Behavior != "" {
			c.Unsupported(dropTableBehaviorFeature)
			return true
		}
		return false
	case AlterIndexRename:
		target := n.OldName
		if n.HasTable {
			target = n.Table + "." + n.OldName
		}
		c.Write("EXEC sp_rename ")
		c.Write(c.Renderer().Literal(target))
		c.Write(", ")
		c.Write(c.Renderer().Literal(n.NewName))
		c.Write(", 'INDEX'")
		return true
	case AlterColumnNullability:
		c.Unsupported(alterColumnNullable)
		return true
	case LimitOffset:
		return compileSQLServerLimitOffset(c, n)
	default:
		return false
	}
}

func compileSQLServerOrder(c Compiler, n OrderItem) bool {
	if n.Nulls == "" {
		return false
	}

	c.Write("CASE WHEN ")
	c.Compile(n.Expr)
	if strings.EqualFold(n.Nulls, "FIRST") {
		c.Write(" IS NULL THEN 0 ELSE 1 END, ")
	} else {
		c.Write(" IS NULL THEN 1 ELSE 0 END, ")
	}
	c.Compile(n.Expr)
	if n.Direction != "" {
		c.Write(" ")
		c.Write(n.Direction)
	}
	return true
}

func compileSQLServerLimitOffset(c Compiler, n LimitOffset) bool {
	switch {
	case n.Limit > 0 && n.Offset > 0:
		c.Write("\nOFFSET ")
		c.WriteInt(n.Offset)
		c.Write(" ROWS FETCH NEXT ")
		c.WriteInt(n.Limit)
		c.Write(" ROWS ONLY")
	case n.Limit > 0:
		c.Write("\nOFFSET 0 ROWS FETCH NEXT ")
		c.WriteInt(n.Limit)
		c.Write(" ROWS ONLY")
	case n.Offset > 0:
		c.Write("\nOFFSET ")
		c.WriteInt(n.Offset)
		c.Write(" ROWS")
	default:
		return false
	}
	return true
}
