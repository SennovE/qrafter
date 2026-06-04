package dialect

import "fmt"

const (
	oracleDialectName  = "Oracle"
	dropTableRestrict  = "DROP TABLE RESTRICT"
	oraclePartialIndex = "PARTIAL INDEX"
)

// Oracle renders qrafter queries using Oracle SQL syntax.
type Oracle struct {
	BaseDialect
}

// DialectName returns the dialect name.
func (Oracle) DialectName() string {
	return oracleDialectName
}

// Literal renders Oracle-friendly inline SQL literals.
func (Oracle) Literal(value any) string {
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

// Placeholder renders an Oracle numbered bind placeholder.
func (Oracle) Placeholder(position int) string {
	return fmt.Sprintf(":%d", position)
}

// CompileNode renders Oracle-specific compiler nodes.
func (Oracle) CompileNode(c Compiler, node any) bool {
	switch n := node.(type) {
	case Returning:
		c.Unsupported("RETURNING")
		return true
	case PartialIndexPredicate:
		c.Unsupported(oraclePartialIndex)
		return true
	case DropTableBehavior:
		return compileOracleDropTableBehavior(c, n)
	case AlterColumnType:
		c.Write("MODIFY ")
		c.Write(c.Renderer().QuoteIdent(n.Column))
		c.Write(" ")
		c.Write(n.Type)
		return true
	case AlterColumnNullability:
		c.Write("MODIFY ")
		c.Write(c.Renderer().QuoteIdent(n.Column))
		if n.Set {
			c.Write(" NOT NULL")
		} else {
			c.Write(" NULL")
		}
		return true
	case AlterColumnDefault:
		c.Write("MODIFY ")
		c.Write(c.Renderer().QuoteIdent(n.Column))
		if n.Drop {
			c.Write(" DEFAULT NULL")
			return true
		}
		c.Write(" DEFAULT ")
		if n.IsExpr {
			c.Write(n.Expr)
		} else {
			c.Write(c.Renderer().Literal(n.Value))
		}
		return true
	case LimitOffset:
		return compileOracleLimitOffset(c, n)
	default:
		return false
	}
}

func compileOracleDropTableBehavior(c Compiler, n DropTableBehavior) bool {
	switch n.Behavior {
	case "":
		return false
	case "CASCADE":
		c.Write(" CASCADE CONSTRAINTS")
		return true
	case "RESTRICT":
		c.Unsupported(dropTableRestrict)
		return true
	default:
		return false
	}
}

func compileOracleLimitOffset(c Compiler, n LimitOffset) bool {
	switch {
	case n.Limit > 0 && n.Offset > 0:
		c.Write("\nOFFSET ")
		c.WriteInt(n.Offset)
		c.Write(" ROWS FETCH NEXT ")
		c.WriteInt(n.Limit)
		c.Write(" ROWS ONLY")
	case n.Limit > 0:
		c.Write("\nFETCH FIRST ")
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
