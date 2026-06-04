package ddl

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// compiler renders DDL AST/state nodes into SQL for a dialect.
type compiler struct {
	w        *strings.Builder
	renderer dialect.Renderer
}

type tableConstraintNode struct {
	table      string
	constraint TableConstraint
}

func newCompiler(d dialect.Renderer) *compiler {
	return &compiler{
		w:        &strings.Builder{},
		renderer: d,
	}
}

func compileInto(w *strings.Builder, d dialect.Renderer, node any) {
	c := &compiler{w: w, renderer: d}
	c.Compile(node)
}

func (c *compiler) SQL() string {
	return c.w.String()
}

func (c *compiler) Renderer() dialect.Renderer {
	return c.renderer
}

func (c *compiler) Write(s string) {
	c.w.WriteString(s)
}

func (c *compiler) WriteInt(n int) {
	c.w.WriteString(strconv.Itoa(n))
}

func (c *compiler) Unsupported(feature string) {
	panic(dialect.UnsupportedFeatureError{
		Dialect: dialect.UnwrapRenderer(c.renderer).DialectName(),
		Feature: feature,
	})
}

func (c *compiler) CompileList(nodes []any, delimiter string) {
	for i, node := range nodes {
		if i > 0 {
			c.Write(delimiter)
		}
		c.Compile(node)
	}
}

func (c *compiler) Compile(node any) {
	if node == nil {
		return
	}
	if dialect.UnwrapRenderer(c.renderer).CompileNode(c, node) {
		return
	}
	if c.compileStatementNode(node) ||
		c.compileLeafNode(node) ||
		c.compileDialectNode(node) ||
		c.compileAlterTableNode(node) ||
		c.compileFallbackNode(node) {
		return
	}

	panic(fmt.Errorf("unsupported DDL compiler node %T", node))
}

func (c *compiler) compileStatementNode(node any) bool {
	switch n := node.(type) {
	case Statements:
		c.compileStatements(n)
	case CreateTableStmt:
		c.compileCreateTable(n)
	case DropTableStmt:
		c.compileDropTable(n)
	case AlterTableStmt:
		c.compileAlterTable(n)
	case CreateIndexStmt:
		c.compileCreateIndex(n)
	case DropIndexStmt:
		c.compileDropIndex(n)
	case AlterIndexStmt:
		c.compileAlterIndex(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileLeafNode(node any) bool {
	switch n := node.(type) {
	case ColumnDef:
		c.compileColumn(n)
	case IndexKey:
		c.compileIndexKey(n)
	case Expression:
		c.compileDDLExpression(n)
	case Predicate:
		c.compileDDLPredicate(n)
	case tableConstraintNode:
		c.compileTableConstraint(n.table, n.constraint)
	default:
		return false
	}
	return true
}

func (c *compiler) compileDialectNode(node any) bool {
	switch n := node.(type) {
	case dialect.PartialIndexPredicate:
		c.compilePartialIndexPredicate(n)
	case dialect.DropTableBehavior:
		c.compileDropTableBehavior(n)
	case dialect.AlterIndexRename:
		c.compileAlterIndexRename(n)
	case dialect.AlterColumnType:
		c.compileAlterColumnType(n)
	case dialect.AlterColumnNullability:
		c.compileAlterColumnNullability(n)
	case dialect.AlterColumnDefault:
		c.compileAlterColumnDefault(n)
	case dialect.AlterTableAddConstraint:
		c.compileAlterTableAddConstraint(n)
	case dialect.AlterTableDropConstraint:
		c.compileAlterTableDropConstraint(n)
	case dialect.AlterTableOperationSeparator:
		c.compileAlterTableOperationSeparator(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileAlterTableNode(node any) bool {
	switch n := node.(type) {
	case renameColumnStmt:
		c.compileRenameColumn(n)
	case addColumnStmt:
		c.compileAddColumn(n)
	case dropColumnStmt:
		c.compileDropColumn(n)
	case alterColumnTypeStmt:
		c.Compile(dialect.AlterColumnType{Column: n.column, Type: n.typ.render(c.renderer)})
	case changeNotNullStmt:
		c.Compile(dialect.AlterColumnNullability{Column: n.column, Set: n.op == setNotNull})
	case setDefaultStmt:
		c.Compile(dialect.AlterColumnDefault{
			Column: n.column,
			IsExpr: n.isExpr,
			Expr:   n.expr,
			Value:  n.value,
		})
	case dropDefaultStmt:
		c.Compile(dialect.AlterColumnDefault{Column: n.column, Drop: true})
	case addConstraintStmt:
		c.Compile(dialect.AlterTableAddConstraint{
			Constraint: tableConstraintNode{table: n.table, constraint: n.c},
		})
	case dropConstraintStmt:
		c.Compile(dialect.AlterTableDropConstraint{Name: n.name, IfExists: n.ifExists})
	case renameConstraintStmt:
		c.compileRenameConstraint(n)
	case renameIndexStmt:
		c.Compile(dialect.AlterIndexRename{OldName: n.oldName, NewName: n.newName})
	default:
		return false
	}
	return true
}

func (c *compiler) compileFallbackNode(node any) bool {
	switch n := node.(type) {
	case Renderer:
		c.Write(n.MustRender(c.renderer))
	default:
		return false
	}
	return true
}

func (c *compiler) compileStatements(statements Statements) {
	for _, stmt := range statements {
		before := c.w.Len()
		c.Compile(stmt)
		if c.w.Len() > before {
			c.Write(";\n")
		}
	}
}

func (c *compiler) compileCreateTable(s CreateTableStmt) {
	if len(s.columns) == 0 && len(s.constraints) == 0 {
		panic(fmt.Errorf("CREATE TABLE %q must include at least one column or constraint", s.name))
	}

	c.Write("CREATE TABLE ")
	if s.ifNotExists {
		c.Write("IF NOT EXISTS ")
	}
	c.Write(c.renderer.QuoteIdent(s.name))
	c.Write(" (\n")

	item := 0
	for _, column := range s.columns {
		if item > 0 {
			c.Write(",\n")
		}
		c.Write("    ")
		c.Compile(column)
		item++
	}
	for _, constraint := range s.constraints {
		if item > 0 {
			c.Write(",\n")
		}
		c.Write("    ")
		c.Compile(tableConstraintNode{table: s.name, constraint: constraint})
		item++
	}

	c.Write("\n)")
}

func (c *compiler) compileDropTable(s DropTableStmt) {
	c.Write("DROP TABLE ")
	if s.ifExists {
		c.Write("IF EXISTS ")
	}
	for i, table := range s.tables {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.renderer.QuoteIdent(table))
	}
	c.Compile(dialect.DropTableBehavior{Behavior: dropBehaviorSQL(s.behavior)})
}

func (c *compiler) compileAlterTable(s AlterTableStmt) {
	if len(s.operations) == 0 {
		panic(fmt.Errorf("ALTER TABLE %q must include at least one operation", s.table))
	}

	c.Write("ALTER TABLE ")
	c.Write(c.renderer.QuoteIdent(s.table))
	for i, op := range s.operations {
		c.Compile(dialect.AlterTableOperationSeparator{Table: s.table, Index: i})
		c.Compile(op)
	}
}

func (c *compiler) compileCreateIndex(s CreateIndexStmt) {
	if len(s.keys) == 0 {
		panic(fmt.Errorf("CREATE INDEX %q must include at least one key", s.name))
	}

	c.compileCreateIndexPrefix(s)
	c.compileCreateIndexTarget(s)
	c.compileCreateIndexKeys(s)
	c.compileCreateIndexInclude(s)
	c.compileCreateIndexOptions(s)
}

func (c *compiler) compileCreateIndexPrefix(s CreateIndexStmt) {
	c.Write("CREATE ")
	if s.options != nil && s.options.unique {
		c.Write("UNIQUE ")
	}
	if s.options != nil && s.options.clustered != nil {
		if *s.options.clustered {
			c.Write("CLUSTERED ")
		} else {
			c.Write("NONCLUSTERED ")
		}
	}
	c.Write("INDEX ")
	if s.options != nil && s.options.concurrently {
		c.Write("CONCURRENTLY ")
	}
	if s.options != nil && s.options.ifNotExists {
		c.Write("IF NOT EXISTS ")
	}
}

func (c *compiler) compileCreateIndexTarget(s CreateIndexStmt) {
	c.Write(c.renderer.QuoteIdent(s.name))
	c.Write(" ON ")
	c.Write(c.renderer.QuoteIdent(s.table))
	if s.options != nil && s.options.method != IndexDefault {
		c.Write(" USING ")
		c.Write(string(s.options.method))
	}
}

func (c *compiler) compileCreateIndexKeys(s CreateIndexStmt) {
	c.Write(" (")
	c.CompileList(indexKeysAsAny(s.keys), ", ")
	c.Write(")")
}

func (c *compiler) compileCreateIndexInclude(s CreateIndexStmt) {
	if s.options != nil && len(s.options.include) > 0 {
		c.Write(" INCLUDE (")
		c.CompileList(expressionsAsAny(s.options.include), ", ")
		c.Write(")")
	}
}

func (c *compiler) compileCreateIndexOptions(s CreateIndexStmt) {
	if s.options != nil && s.options.nullsNotDistinct {
		c.Write(" NULLS NOT DISTINCT")
	}
	c.compileCreateIndexStorage(s)
	if s.options != nil && s.options.tablespace != "" {
		c.Write(" TABLESPACE ")
		c.Write(c.renderer.QuoteIdent(s.options.tablespace))
	}
	if s.options != nil && s.options.pred != nil {
		c.Compile(dialect.PartialIndexPredicate{Predicate: *s.options.pred})
	}
	if s.options != nil && s.options.invisible {
		c.Write(" INVISIBLE")
	}
}

func (c *compiler) compileCreateIndexStorage(s CreateIndexStmt) {
	if s.options == nil || len(s.options.with) == 0 {
		return
	}

	c.Write(" WITH (")
	for i, opt := range s.options.with {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(opt.name)
		c.Write(" = ")
		c.Write(renderIndexOptionValue(c.renderer, opt.value))
	}
	c.Write(")")
}

func (c *compiler) compileDropIndex(s DropIndexStmt) {
	c.Write("DROP INDEX ")
	if s.concurrently {
		c.Write("CONCURRENTLY ")
	}
	if s.ifExists {
		c.Write("IF EXISTS ")
	}
	c.Write(c.renderer.QuoteIdent(s.name))
	if s.table != nil {
		c.Write(" ON ")
		c.Write(c.renderer.QuoteIdent(*s.table))
	}
	if s.online {
		c.Write(" ONLINE")
	}
	c.Compile(dialect.DropTableBehavior{Behavior: dropBehaviorSQL(s.behavior)})
}

func (c *compiler) compileAlterIndex(s AlterIndexStmt) {
	if s.operation == nil {
		panic(fmt.Errorf("ALTER INDEX %q must include an operation", s.name))
	}

	switch op := s.operation.(type) {
	case renameIndexStmt:
		table := ""
		if s.table != nil {
			table = *s.table
		}
		c.Compile(dialect.AlterIndexRename{
			OldName:  op.oldName,
			NewName:  op.newName,
			Table:    table,
			HasTable: s.table != nil,
		})
	default:
		c.Compile(op)
	}
}

func (c *compiler) compileColumn(column ColumnDef) {
	c.Write(c.renderer.QuoteIdent(column.name))
	c.Write(" ")
	c.Write(column.typ.render(c.renderer))

	if column.notNull {
		c.Write(" NOT NULL")
	}
	if column.options != nil && column.options.defaultValue != nil {
		c.Write(" DEFAULT ")
		if column.options.defaultValue.isExpr {
			c.Write(column.options.defaultValue.expr)
		} else {
			c.Write(c.renderer.Literal(column.options.defaultValue.value))
		}
	}
	if column.primaryKey {
		c.Write(" PRIMARY KEY")
	}
	if column.unique {
		c.Write(" UNIQUE")
	}
	if column.options != nil {
		for i := range column.options.checks {
			c.Write(" CHECK (")
			c.Write(column.options.checks[i].expr)
			c.Write(")")
		}
	}
	if column.options == nil || column.options.references == nil {
		return
	}

	c.Write(" REFERENCES ")
	c.Write(c.renderer.QuoteIdent(column.options.references.table))
	if len(column.options.references.columns) > 0 {
		c.Write(" (")
		c.compileColumnList(column.options.references.columns)
		c.Write(")")
	}
}

func (c *compiler) compileIndexKey(k IndexKey) {
	c.Compile(k.expr)

	if k.options == nil {
		return
	}
	if k.options.length > 0 {
		c.Write("(")
		c.WriteInt(k.options.length)
		c.Write(")")
	}
	if k.options.collation != "" {
		c.Write(" COLLATE ")
		c.Write(c.renderer.QuoteIdent(k.options.collation))
	}
	if k.options.opclass != "" {
		c.Write(" ")
		c.Write(k.options.opclass)
	}
	if k.options.sort != "" {
		c.Write(" ")
		c.Write(string(k.options.sort))
	}
	if k.options.nulls != "" {
		c.Write(" NULLS ")
		c.Write(string(k.options.nulls))
	}
}

func (c *compiler) compileDDLExpression(expr Expression) {
	switch n := expr.node.(type) {
	case columnExpression:
		c.Write(c.renderer.QuoteIdent(n.name))
	case literalExpression:
		c.Write(c.renderer.Literal(n.value))
	case rawExpression:
		c.Write(n.sql)
	case functionExpression:
		c.Write(n.name)
		c.Write("(")
		c.CompileList(expressionsAsAny(n.args), ", ")
		c.Write(")")
	case binaryExpression:
		c.compileDDLExpressionChild(n.left, n.prec, false)
		c.Write(" ")
		c.Write(n.op)
		c.Write(" ")
		c.compileDDLExpressionChild(n.right, n.prec, n.parenthesizeRightPeer)
	default:
		panic(fmt.Errorf("unsupported DDL expression node %T", expr.node))
	}
}

func (c *compiler) compileDDLPredicate(pred Predicate) {
	switch n := pred.node.(type) {
	case binaryPredicate:
		c.compileDDLExpressionChild(n.left, precedenceComparison, false)
		c.Write(" ")
		c.Write(n.op)
		c.Write(" ")
		c.compileDDLExpressionChild(n.right, precedenceComparison, false)
	case logicalPredicate:
		for i, item := range n.predicates {
			if i > 0 {
				c.Write(" ")
				c.Write(n.op)
				c.Write(" ")
			}
			c.compileDDLPredicateChild(item, n.prec, false)
		}
	case rawPredicate:
		c.Write(n.sql)
	default:
		panic(fmt.Errorf("unsupported DDL predicate node %T", pred.node))
	}
}

func (c *compiler) compileDDLExpressionChild(expr Expression, parentPrecedence int, parenthesizeOnEqual bool) {
	if needsParentheses(expr.prec, parentPrecedence, parenthesizeOnEqual) {
		c.Write("(")
		c.compileDDLExpression(expr)
		c.Write(")")
		return
	}
	c.compileDDLExpression(expr)
}

func (c *compiler) compileDDLPredicateChild(pred Predicate, parentPrecedence int, parenthesizeOnEqual bool) {
	if needsParentheses(pred.prec, parentPrecedence, parenthesizeOnEqual) {
		c.Write("(")
		c.compileDDLPredicate(pred)
		c.Write(")")
		return
	}
	c.compileDDLPredicate(pred)
}

func needsParentheses(childPrecedence, parentPrecedence int, parenthesizeOnEqual bool) bool {
	return childPrecedence < parentPrecedence || childPrecedence == parentPrecedence && parenthesizeOnEqual
}

func (c *compiler) compileTableConstraint(table string, constraint TableConstraint) {
	switch n := constraint.(type) {
	case Constraint[primaryKey]:
		c.compilePrimaryKey(table, n)
	case Constraint[unique]:
		c.compileUnique(table, n)
	case Constraint[check]:
		c.compileCheck(table, n)
	case ForeignKeyConstraint:
		c.compileForeignKey(table, n.Constraint)
	case Constraint[foreignKey]:
		c.compileForeignKey(table, n)
	default:
		panic(fmt.Errorf("unsupported table constraint %T", constraint))
	}
}

func (c *compiler) compilePrimaryKey(table string, constraint Constraint[primaryKey]) {
	parts := append([]string{table}, constraint.c.columns...)
	name := constraintName(constraint.name, "pk", parts...)
	c.compileNamedColumnConstraint(name, "PRIMARY KEY", constraint.c.columns)
}

func (c *compiler) compileUnique(table string, constraint Constraint[unique]) {
	parts := append([]string{table}, constraint.c.columns...)
	name := constraintName(constraint.name, "uq", parts...)
	c.compileNamedColumnConstraint(name, "UNIQUE", constraint.c.columns)
}

func (c *compiler) compileCheck(table string, constraint Constraint[check]) {
	var tmp strings.Builder
	compileInto(&tmp, c.renderer, constraint.c.expr)
	sql := tmp.String()

	name := constraint.name
	if name == nil {
		fields := strings.Fields(strings.ToLower(sql))
		normalized := strings.Join(fields, " ")
		sum := sha256.Sum256([]byte(normalized))
		hash := hex.EncodeToString(sum[:])[:8]
		generated := fmt.Sprintf("chk_%s_%s", table, hash)
		name = &generated
	}

	c.Write("CONSTRAINT ")
	c.Write(c.renderer.QuoteIdent(*name))
	c.Write(" CHECK (")
	c.Write(sql)
	c.Write(")")
}

func (c *compiler) compileForeignKey(table string, constraint Constraint[foreignKey]) {
	fk := constraint.c
	if fk.reference == nil {
		panic("foreign key reference is required")
	}
	if len(fk.srcCols) != len(fk.reference.columns) {
		panic("the number of columns on the left and right must match")
	}

	parts := append([]string{table}, fk.srcCols...)
	parts = append(parts, fk.reference.table)
	parts = append(parts, fk.reference.columns...)
	name := constraintName(constraint.name, "fk", parts...)

	c.compileNamedColumnConstraint(name, "FOREIGN KEY", fk.srcCols)
	c.Write(" REFERENCES ")
	c.Write(c.renderer.QuoteIdent(fk.reference.table))
	c.Write(" (")
	c.compileColumnList(fk.reference.columns)
	c.Write(")")
	if fk.options != nil && fk.options.onDelete != nil {
		c.Write(" ON DELETE ")
		c.Write(string(*fk.options.onDelete))
	}
	if fk.options != nil && fk.options.onUpdate != nil {
		c.Write(" ON UPDATE ")
		c.Write(string(*fk.options.onUpdate))
	}
}

func (c *compiler) compileNamedColumnConstraint(name, opType string, columns []string) {
	c.Write("CONSTRAINT ")
	c.Write(c.renderer.QuoteIdent(name))
	c.Write(" ")
	c.Write(opType)
	c.Write(" (")
	c.compileColumnList(columns)
	c.Write(")")
}

func (c *compiler) compileColumnList(columns []string) {
	for i, column := range columns {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.renderer.QuoteIdent(column))
	}
}

func (c *compiler) compilePartialIndexPredicate(n dialect.PartialIndexPredicate) {
	c.Write(" WHERE ")
	c.Compile(n.Predicate)
}

func (c *compiler) compileDropTableBehavior(n dialect.DropTableBehavior) {
	if n.Behavior == "" {
		return
	}
	c.Write(" ")
	c.Write(n.Behavior)
}

func (c *compiler) compileAlterIndexRename(n dialect.AlterIndexRename) {
	c.Write("ALTER INDEX ")
	c.Write(c.renderer.QuoteIdent(n.OldName))
	c.Write(" RENAME TO ")
	c.Write(c.renderer.QuoteIdent(n.NewName))
}

func (c *compiler) compileAlterColumnType(n dialect.AlterColumnType) {
	c.Write("ALTER COLUMN ")
	c.Write(c.renderer.QuoteIdent(n.Column))
	c.Write(" TYPE ")
	c.Write(n.Type)
}

func (c *compiler) compileAlterColumnNullability(n dialect.AlterColumnNullability) {
	c.Write("ALTER COLUMN ")
	c.Write(c.renderer.QuoteIdent(n.Column))
	if n.Set {
		c.Write(" SET NOT NULL")
	} else {
		c.Write(" DROP NOT NULL")
	}
}

func (c *compiler) compileAlterColumnDefault(n dialect.AlterColumnDefault) {
	c.Write("ALTER COLUMN ")
	c.Write(c.renderer.QuoteIdent(n.Column))
	if n.Drop {
		c.Write(" DROP DEFAULT")
		return
	}
	c.Write(" SET DEFAULT ")
	if n.IsExpr {
		c.Write(n.Expr)
		return
	}
	c.Write(c.renderer.Literal(n.Value))
}

func (c *compiler) compileAlterTableAddConstraint(n dialect.AlterTableAddConstraint) {
	c.Write("ADD ")
	c.Compile(n.Constraint)
}

func (c *compiler) compileAlterTableDropConstraint(n dialect.AlterTableDropConstraint) {
	c.Write("DROP CONSTRAINT ")
	if n.IfExists {
		c.Write("IF EXISTS ")
	}
	c.Write(c.renderer.QuoteIdent(n.Name))
}

func (c *compiler) compileAlterTableOperationSeparator(n dialect.AlterTableOperationSeparator) {
	if n.Index == 0 {
		c.Write("\n    ")
		return
	}
	c.Write(",\n    ")
}

func (c *compiler) compileRenameColumn(s renameColumnStmt) {
	c.Write("RENAME COLUMN ")
	c.Write(c.renderer.QuoteIdent(s.column))
	c.Write(" TO ")
	c.Write(c.renderer.QuoteIdent(s.name))
}

func (c *compiler) compileAddColumn(s addColumnStmt) {
	c.Write("ADD COLUMN ")
	if s.ifNotExists {
		c.Write("IF NOT EXISTS ")
	}
	c.Compile(s.column)
}

func (c *compiler) compileDropColumn(s dropColumnStmt) {
	c.Write("DROP COLUMN ")
	if s.ifExists {
		c.Write("IF EXISTS ")
	}
	c.Write(c.renderer.QuoteIdent(s.column))
}

func (c *compiler) compileRenameConstraint(s renameConstraintStmt) {
	c.Write("RENAME CONSTRAINT ")
	c.Write(c.renderer.QuoteIdent(s.column))
	c.Write(" TO ")
	c.Write(c.renderer.QuoteIdent(s.name))
}

func dropBehaviorSQL(behavior dropBehavior) string {
	switch behavior {
	case dropCascade:
		return "CASCADE"
	case dropRestrict:
		return "RESTRICT"
	default:
		return ""
	}
}

func constraintName(existing *string, prefix string, parts ...string) string {
	if existing != nil {
		return *existing
	}
	return fmt.Sprintf("%s_%s", prefix, strings.Join(parts, "_"))
}

func expressionsAsAny(items []Expression) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}

func indexKeysAsAny(items []IndexKey) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}
