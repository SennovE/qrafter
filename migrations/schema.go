package migrations

import (
	"sort"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

// Schema is a normalized database schema snapshot.
type Schema struct {
	Tables []Table
}

// RegisterTable appends the schema inferred from table config T to schema.
func RegisterTable[T qrafter.TableConfigProvider](schema *Schema, d dialect.Renderer) {
	newSchema := schemaFromTableConfig[T](d)
	schema.Tables = append(schema.Tables, newSchema.Tables...)
}

func (s *Schema) normalize() {
	sort.SliceStable(s.Tables, func(i, j int) bool {
		if s.Tables[i].Schema == s.Tables[j].Schema {
			return s.Tables[i].Name < s.Tables[j].Name
		}
		return s.Tables[i].Schema < s.Tables[j].Schema
	})
	for i := range s.Tables {
		s.Tables[i].normalize()
	}
}

// Table describes a database table.
type Table struct {
	Schema      string
	Name        string
	Columns     []Column
	Constraints []Constraint
	Indexes     []Index
}

func (t *Table) createTable() ddl.CreateTableStmt {
	stmt := ddl.CreateTable(t.Name)
	for i := range t.Columns {
		stmt = stmt.Columns(t.Columns[i].columnDef())
	}
	for i := range t.Constraints {
		stmt = stmt.Constraints(t.Constraints[i].tableConstraint())
	}
	return stmt
}

func (t *Table) normalize() {
	sort.SliceStable(t.Columns, func(i, j int) bool {
		return t.Columns[i].Position < t.Columns[j].Position
	})
	sort.SliceStable(t.Constraints, func(i, j int) bool {
		left := constraintKindOrder(t.Constraints[i].Kind)
		right := constraintKindOrder(t.Constraints[j].Kind)
		if left == right {
			return t.Constraints[i].Name < t.Constraints[j].Name
		}
		return left < right
	})
	sort.SliceStable(t.Indexes, func(i, j int) bool {
		return t.Indexes[i].Name < t.Indexes[j].Name
	})
}

// Column describes a table column.
type Column struct {
	Schema    string
	TableName string
	Position  int
	Name      string
	Type      ddl.Type

	// DatabaseType is the type name returned by database introspection.
	DatabaseType string

	NotNull bool

	HasDefault  bool
	DefaultExpr string

	Identity ddl.IdentityKind

	Generated     ddl.GeneratedKind
	GeneratedExpr string
}

func (c *Column) columnDef() ddl.ColumnDef {
	column := ddl.Column(c.Name, c.ddlType())
	switch c.Identity {
	case ddl.IdentityAlways:
		column = column.IdentityAlways()
	case ddl.IdentityByDefault:
		column = column.IdentityByDefault()
	}
	switch c.Generated {
	case ddl.GeneratedStored:
		column = column.GeneratedStored(c.GeneratedExpr)
	case ddl.GeneratedVirtual:
		column = column.GeneratedVirtual(c.GeneratedExpr)
	}
	if c.NotNull {
		column = column.NotNull()
	}
	if c.HasDefault && c.Identity == ddl.IdentityNone && c.Generated == ddl.GeneratedNone {
		column = column.DefaultExpr(c.DefaultExpr)
	}
	return column
}

func (c *Column) ddlType() ddl.Type {
	if !c.Type.IsZero() {
		return c.Type
	}
	return ddl.SQLType(c.DatabaseType)
}

// ConstraintKind describes a table-level constraint type.
type ConstraintKind string

const (
	// ConstraintPrimaryKey is a PRIMARY KEY constraint.
	ConstraintPrimaryKey ConstraintKind = "primary_key"
	// ConstraintUnique is a UNIQUE constraint.
	ConstraintUnique ConstraintKind = "unique"
	// ConstraintCheck is a CHECK constraint.
	ConstraintCheck ConstraintKind = "check"
	// ConstraintForeignKey is a FOREIGN KEY constraint.
	ConstraintForeignKey ConstraintKind = "foreign_key"
)

// Reference describes a foreign-key target.
type Reference struct {
	Schema    string
	TableName string
	Columns   []string
}

// Constraint describes a table-level constraint.
type Constraint struct {
	Schema    string
	TableName string
	Name      string
	Kind      ConstraintKind
	Columns   []string

	CheckExpr string

	Reference Reference
	OnDelete  ddl.ReferenceAction
	OnUpdate  ddl.ReferenceAction
}

func (c *Constraint) tableConstraint() ddl.TableConstraint {
	switch c.Kind {
	case ConstraintPrimaryKey:
		constraint := ddl.PrimaryKey(c.Columns...)
		if c.Name != "" {
			return constraint.Named(c.Name)
		}
		return constraint
	case ConstraintUnique:
		constraint := ddl.Unique(c.Columns...)
		if c.Name != "" {
			return constraint.Named(c.Name)
		}
		return constraint
	case ConstraintCheck:
		constraint := ddl.Check(ddl.RawPred(c.CheckExpr))
		if c.Name != "" {
			return constraint.Named(c.Name)
		}
		return constraint
	case ConstraintForeignKey:
		constraint := ddl.ForeignKey(c.Columns...).References(c.Reference.TableName, c.Reference.Columns...)
		if c.OnDelete != "" && c.OnDelete != ddl.NoAction {
			constraint = constraint.OnDelete(c.OnDelete)
		}
		if c.OnUpdate != "" && c.OnUpdate != ddl.NoAction {
			constraint = constraint.OnUpdate(c.OnUpdate)
		}
		if c.Name != "" {
			return constraint.Named(c.Name)
		}
		return constraint
	default:
		panic("migrations: unsupported constraint kind " + string(c.Kind))
	}
}

func constraintKindOrder(kind ConstraintKind) int {
	switch kind {
	case ConstraintPrimaryKey:
		return 1
	case ConstraintUnique:
		return 2
	case ConstraintCheck:
		return 3
	case ConstraintForeignKey:
		return 4
	default:
		return 100
	}
}

// IndexKey describes one PostgreSQL index key expression.
type IndexKey struct {
	Expression string
}

// Index describes a non-constraint index.
type Index struct {
	Schema      string
	TableSchema string
	TableName   string
	Name        string

	Unique bool
	Method ddl.IndexMethod

	Keys    []IndexKey
	Include []string

	Predicate string

	Tablespace       string
	NullsNotDistinct bool
}

func (i *Index) createIndex() ddl.CreateIndexStmt {
	keys := make([]ddl.IndexKey, 0, len(i.Keys))
	for j := range i.Keys {
		keys = append(keys, ddl.Key(ddl.RawExpr(i.Keys[j].Expression)))
	}

	stmt := ddl.CreateIndex(i.Name).On(i.TableName, keys...)
	if i.Unique {
		stmt = stmt.Unique()
	}
	if i.Method != "" {
		stmt = stmt.Using(i.Method)
	}
	if len(i.Include) > 0 {
		include := make([]ddl.Expression, 0, len(i.Include))
		for j := range i.Include {
			include = append(include, ddl.RawExpr(i.Include[j]))
		}
		stmt = stmt.Include(include...)
	}
	if i.NullsNotDistinct {
		stmt = stmt.NullsNotDistinct()
	}
	if i.Tablespace != "" {
		stmt = stmt.Tablespace(i.Tablespace)
	}
	if i.Predicate != "" {
		stmt = stmt.Where(ddl.RawPred(i.Predicate))
	}
	return stmt
}
