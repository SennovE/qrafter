package migrations

import (
	"fmt"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

func tableConstraintFromDDL(d dialect.Renderer, table *Table, src ddl.TableConstraint) Constraint {
	switch c := src.(type) {
	case ddl.Constraint[ddl.PrimaryKeyKind]:
		return Constraint{
			Schema:    table.Schema,
			TableName: table.Name,
			Name:      constraintDDLName(c.Name),
			Kind:      ConstraintPrimaryKey,
			Columns:   append([]string(nil), c.Kind.Columns...),
		}
	case ddl.Constraint[ddl.UniqueKind]:
		return Constraint{
			Schema:    table.Schema,
			TableName: table.Name,
			Name:      constraintDDLName(c.Name),
			Kind:      ConstraintUnique,
			Columns:   append([]string(nil), c.Kind.Columns...),
		}
	case ddl.Constraint[ddl.CheckKind]:
		return Constraint{
			Schema:    table.Schema,
			TableName: table.Name,
			Name:      constraintDDLName(c.Name),
			Kind:      ConstraintCheck,
			CheckExpr: renderDDL(d, c.Kind.Expr),
		}
	case ddl.ForeignKeyConstraint:
		return foreignKeyFromDDL(table, c.Constraint)
	case ddl.Constraint[ddl.ForeignKeyKind]:
		return foreignKeyFromDDL(table, c)
	default:
		panic(fmt.Sprintf("migrations: unsupported table constraint %T", src))
	}
}

func foreignKeyFromDDL(table *Table, c ddl.Constraint[ddl.ForeignKeyKind]) Constraint {
	out := Constraint{
		Schema:    table.Schema,
		TableName: table.Name,
		Name:      constraintDDLName(c.Name),
		Kind:      ConstraintForeignKey,
		Columns:   append([]string(nil), c.Kind.SourceColumns...),
		OnDelete:  ddl.NoAction,
		OnUpdate:  ddl.NoAction,
	}
	if c.Kind.Reference != nil {
		out.Reference = Reference{
			Schema:    table.Schema,
			TableName: c.Kind.Reference.Table,
			Columns:   append([]string(nil), c.Kind.Reference.Columns...),
		}
	}
	if c.Kind.Options != nil {
		if c.Kind.Options.OnDelete != nil {
			out.OnDelete = *c.Kind.Options.OnDelete
		}
		if c.Kind.Options.OnUpdate != nil {
			out.OnUpdate = *c.Kind.Options.OnUpdate
		}
	}
	return out
}

func constraintDDLName(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}

func indexFromDDL(d dialect.Renderer, table *Table, src ddl.CreateIndexStmt) Index {
	tableName := src.Table
	if tableName == "" {
		tableName = table.Name
	}

	out := Index{
		Schema:      table.Schema,
		TableSchema: table.Schema,
		TableName:   tableName,
		Name:        src.Name,
		Keys:        make([]IndexKey, 0, len(src.Keys)),
	}
	for _, key := range src.Keys {
		out.Keys = append(out.Keys, IndexKey{Expression: renderDDL(d, key)})
	}

	if src.Options == nil {
		return out
	}
	out.Unique = src.Options.Unique
	out.Method = src.Options.Method
	out.Tablespace = src.Options.Tablespace
	out.NullsNotDistinct = src.Options.NullsNotDistinct
	if src.Options.Predicate != nil {
		out.Predicate = renderDDL(d, *src.Options.Predicate)
	}
	for _, include := range src.Options.Include {
		out.Include = append(out.Include, renderDDL(d, include))
	}
	return out
}

func renderDDL(d dialect.Renderer, node any) string {
	sql, err := ddl.Render(d, node)
	if err != nil {
		panic(err)
	}
	return sql
}
