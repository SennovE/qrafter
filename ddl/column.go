package ddl

// ColumnDef describes a column inside CREATE TABLE or ALTER TABLE ADD COLUMN.
type ColumnDef struct {
	name string
	typ  Type

	primaryKey bool
	notNull    bool
	unique     bool

	options *columnOptions
}

type columnOptions struct {
	defaultValue *columnDefault
	checks       []columnCheck
	references   *columnReferences
}

type columnDefault struct {
	isExpr bool
	value  any
	expr   string
}

type columnCheck struct {
	expr string
}

type columnReferences struct {
	table   string
	columns []string
}

// Column creates a column definition. Name can be a string or qrafter.Column.
func Column(name string, typ Type) ColumnDef {
	return ColumnDef{
		name: name,
		typ:  typ,
	}
}

// PrimaryKey marks the column as a primary key.
func (c ColumnDef) PrimaryKey() ColumnDef {
	c.primaryKey = true
	return c
}

// NotNull adds a NOT NULL constraint.
func (c ColumnDef) NotNull() ColumnDef {
	c.notNull = true
	return c
}

// Null adds an explicit NULL marker.
func (c ColumnDef) Null() ColumnDef {
	c.notNull = false
	return c
}

// Unique adds a UNIQUE constraint.
func (c ColumnDef) Unique() ColumnDef {
	c.unique = true
	return c
}

// Default adds a literal DEFAULT value rendered through the dialect.
func (c ColumnDef) Default(value any) ColumnDef {
	options := c.cloneOptions()
	options.defaultValue = &columnDefault{value: value}
	c.options = options
	return c
}

// DefaultExpr adds a raw SQL DEFAULT expression.
func (c ColumnDef) DefaultExpr(expr string) ColumnDef {
	options := c.cloneOptions()
	options.defaultValue = &columnDefault{isExpr: true, expr: expr}
	c.options = options
	return c
}

// Check adds a column-level CHECK expression.
func (c ColumnDef) Check(expr string) ColumnDef {
	options := c.cloneOptions()
	options.checks = append(options.checks, columnCheck{expr: expr})
	c.options = options
	return c
}

// References adds a column-level foreign key reference.
func (c ColumnDef) References(table string, columns ...string) ColumnDef {
	options := c.cloneOptions()
	options.references = &columnReferences{
		table:   table,
		columns: columns,
	}
	c.options = options
	return c
}

func (c ColumnDef) cloneOptions() *columnOptions {
	if c.options == nil {
		return &columnOptions{}
	}
	options := *c.options
	options.checks = append([]columnCheck(nil), c.options.checks...)
	return &options
}
