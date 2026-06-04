package dialect

const (
	sqliteDeleteUsingFeature = "DELETE USING"
	sqliteDialectName        = "SQLite"
	dropTableBehaviorFeature = "DROP TABLE CASCADE/RESTRICT"
	alterColumnTypeFeature   = "ALTER COLUMN TYPE"
	alterColumnDefault       = "ALTER COLUMN DEFAULT"
	alterTableAddConstraint  = "ALTER TABLE ADD CONSTRAINT"
	alterTableDropConstraint = "ALTER TABLE DROP CONSTRAINT"
)

// SQLite renders qrafter queries using SQLite placeholder and LIMIT/OFFSET
// syntax. Unsupported SQLite features, such as DELETE USING, fail fast with
// UnsupportedFeatureError.
type SQLite struct {
	BaseDialect
}

// DialectName returns the dialect name.
func (SQLite) DialectName() string {
	return sqliteDialectName
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

// CompileNode renders SQLite-specific compiler nodes.
func (SQLite) CompileNode(c Compiler, node any) bool {
	switch n := node.(type) {
	case DeleteTarget:
		if len(n.Using) > 0 {
			c.Unsupported(sqliteDeleteUsingFeature)
			return true
		}
		return false
	case DeleteUsing:
		if len(n.Using) > 0 {
			c.Unsupported(sqliteDeleteUsingFeature)
			return true
		}
		return false
	case DropTableBehavior:
		if n.Behavior != "" {
			c.Unsupported(dropTableBehaviorFeature)
			return true
		}
		return false
	case AlterColumnType:
		c.Unsupported(alterColumnTypeFeature)
		return true
	case AlterColumnNullability:
		c.Unsupported(alterColumnNullable)
		return true
	case AlterColumnDefault:
		c.Unsupported(alterColumnDefault)
		return true
	case AlterTableAddConstraint:
		c.Unsupported(alterTableAddConstraint)
		return true
	case AlterTableDropConstraint:
		c.Unsupported(alterTableDropConstraint)
		return true
	case AlterTableOperationSeparator:
		if n.Index == 0 {
			c.Write(" ")
			return true
		}
		c.Write(";\nALTER TABLE ")
		c.Write(c.Renderer().QuoteIdent(n.Table))
		c.Write(" ")
		return true
	case LimitOffset:
		return compileSQLITELimitOffset(c, n)
	default:
		return false
	}
}

func compileSQLITELimitOffset(c Compiler, n LimitOffset) bool {
	switch {
	case n.Limit > 0 && n.Offset > 0:
		c.Write("\nLIMIT ")
		c.WriteInt(n.Limit)
		c.Write(" OFFSET ")
		c.WriteInt(n.Offset)
	case n.Limit > 0:
		c.Write("\nLIMIT ")
		c.WriteInt(n.Limit)
	case n.Offset > 0:
		c.Write("\nLIMIT -1 OFFSET ")
		c.WriteInt(n.Offset)
	default:
		return false
	}
	return true
}
