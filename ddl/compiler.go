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
		c.compileColumn(n, true)
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
			c.writeStatementSeparator(before)
		}
	}
}

func (c *compiler) writeStatementSeparator(statementStart int) {
	statement := strings.TrimSpace(c.w.String()[statementStart:])
	if strings.HasSuffix(statement, ";") {
		c.Write("\n")
		return
	}
	c.Write(";\n")
}

func (c *compiler) compileCreateTable(s CreateTableStmt) {
	columnConstraints := columnConstraintsAsTableConstraints(s.ColumnDefs)
	if len(s.ColumnDefs) == 0 && len(columnConstraints) == 0 && len(s.TableConstraints) == 0 {
		panic(fmt.Errorf("CREATE TABLE %q must include at least one column or constraint", s.TableName))
	}

	c.Write("CREATE TABLE ")
	if s.IfNotExistsFlag {
		c.Write("IF NOT EXISTS ")
	}
	c.Write(c.renderer.QuoteIdent(s.TableName))
	c.Write(" (\n")

	item := 0
	for _, column := range s.ColumnDefs {
		if item > 0 {
			c.Write(",\n")
		}
		c.Write("    ")
		c.compileColumn(column, false)
		item++
	}
	for _, constraint := range columnConstraints {
		if item > 0 {
			c.Write(",\n")
		}
		c.Write("    ")
		c.Compile(tableConstraintNode{table: s.TableName, constraint: constraint})
		item++
	}
	for _, constraint := range s.TableConstraints {
		if item > 0 {
			c.Write(",\n")
		}
		c.Write("    ")
		c.Compile(tableConstraintNode{table: s.TableName, constraint: constraint})
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
	if len(s.Keys) == 0 {
		panic(fmt.Errorf("CREATE INDEX %q must include at least one key", s.Name))
	}

	c.compileCreateIndexPrefix(s)
	c.compileCreateIndexTarget(s)
	c.compileCreateIndexKeys(s)
	c.compileCreateIndexInclude(s)
	c.compileCreateIndexOptions(s)
}

func (c *compiler) compileCreateIndexPrefix(s CreateIndexStmt) {
	c.Write("CREATE ")
	if s.Options != nil && s.Options.Unique {
		c.Write("UNIQUE ")
	}
	if s.Options != nil && s.Options.Clustered != nil {
		if *s.Options.Clustered {
			c.Write("CLUSTERED ")
		} else {
			c.Write("NONCLUSTERED ")
		}
	}
	c.Write("INDEX ")
	if s.Options != nil && s.Options.Concurrently {
		c.Write("CONCURRENTLY ")
	}
	if s.Options != nil && s.Options.IfNotExists {
		c.Write("IF NOT EXISTS ")
	}
}

func (c *compiler) compileCreateIndexTarget(s CreateIndexStmt) {
	c.Write(c.renderer.QuoteIdent(s.Name))
	c.Write(" ON ")
	c.Write(c.renderer.QuoteIdent(s.Table))
	if s.Options != nil && s.Options.Method != IndexDefault {
		c.Write(" USING ")
		c.Write(string(s.Options.Method))
	}
}

func (c *compiler) compileCreateIndexKeys(s CreateIndexStmt) {
	c.Write(" (")
	c.CompileList(indexKeysAsAny(s.Keys), ", ")
	c.Write(")")
}

func (c *compiler) compileCreateIndexInclude(s CreateIndexStmt) {
	if s.Options != nil && len(s.Options.Include) > 0 {
		c.Write(" INCLUDE (")
		c.CompileList(expressionsAsAny(s.Options.Include), ", ")
		c.Write(")")
	}
}

func (c *compiler) compileCreateIndexOptions(s CreateIndexStmt) {
	if s.Options != nil && s.Options.NullsNotDistinct {
		c.Write(" NULLS NOT DISTINCT")
	}
	c.compileCreateIndexStorage(s)
	if s.Options != nil && s.Options.Tablespace != "" {
		c.Write(" TABLESPACE ")
		c.Write(c.renderer.QuoteIdent(s.Options.Tablespace))
	}
	if s.Options != nil && s.Options.Predicate != nil {
		c.Compile(dialect.PartialIndexPredicate{Predicate: *s.Options.Predicate})
	}
	if s.Options != nil && s.Options.Invisible {
		c.Write(" INVISIBLE")
	}
}

func (c *compiler) compileCreateIndexStorage(s CreateIndexStmt) {
	if s.Options == nil || len(s.Options.With) == 0 {
		return
	}

	c.Write(" WITH (")
	for i, opt := range s.Options.With {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(opt.Name)
		c.Write(" = ")
		c.Write(renderIndexOptionValue(c.renderer, opt.Value))
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

func (c *compiler) compileColumn(column ColumnDef, includeConstraints bool) {
	c.validateColumnGeneratedOptions(column)

	c.Write(c.renderer.QuoteIdent(column.Name))
	c.Write(" ")
	c.Write(column.Type.render(c.renderer))
	c.compileColumnGeneratedOptions(column)

	if column.IsNotNull {
		c.Write(" NOT NULL")
	}
	if column.Options != nil && column.Options.Default != nil {
		c.Write(" DEFAULT ")
		if column.Options.Default.IsExpr {
			c.Write(column.Options.Default.Expr)
		} else {
			c.Write(c.renderer.Literal(column.Options.Default.Value))
		}
	}
	if !includeConstraints {
		return
	}
	if column.IsPrimaryKey {
		c.Write(" PRIMARY KEY")
	}
	if column.IsUnique {
		c.Write(" UNIQUE")
	}
	if column.Options != nil {
		for i := range column.Options.Checks {
			c.Write(" CHECK (")
			c.Write(column.Options.Checks[i].Expr)
			c.Write(")")
		}
	}
	if column.Options == nil || column.Options.References == nil {
		return
	}

	c.Write(" REFERENCES ")
	c.Write(c.renderer.QuoteIdent(column.Options.References.Table))
	if len(column.Options.References.Columns) > 0 {
		c.Write(" (")
		c.compileColumnList(column.Options.References.Columns)
		c.Write(")")
	}
}

func columnConstraintsAsTableConstraints(columns []ColumnDef) []TableConstraint {
	var constraints []TableConstraint
	for _, column := range columns {
		if column.IsPrimaryKey {
			constraints = append(constraints, PrimaryKey(column.Name))
		}
		if column.IsUnique {
			constraints = append(constraints, Unique(column.Name))
		}
		if column.Options == nil {
			continue
		}
		for _, check := range column.Options.Checks {
			constraints = append(constraints, Check(RawPred(check.Expr)))
		}
		if column.Options.References != nil {
			constraints = append(constraints, ForeignKey(column.Name).References(
				column.Options.References.Table,
				column.Options.References.Columns...,
			))
		}
	}
	return constraints
}

func (c *compiler) validateColumnGeneratedOptions(column ColumnDef) {
	if column.Options == nil {
		return
	}
	if column.Options.Identity != nil && column.Options.Generated != nil {
		panic("ddl: column cannot be both identity and generated")
	}
	if column.Options.Identity != nil && column.Options.Default != nil {
		panic("ddl: identity column cannot have DEFAULT")
	}
	if column.Options.Generated != nil && column.Options.Default != nil {
		panic("ddl: generated column cannot have DEFAULT")
	}
}

func (c *compiler) compileColumnGeneratedOptions(column ColumnDef) {
	if column.Options == nil {
		return
	}
	if column.Options.Identity != nil {
		switch column.Options.Identity.Kind {
		case IdentityAlways:
			c.Write(" GENERATED ALWAYS AS IDENTITY")
		case IdentityByDefault:
			c.Write(" GENERATED BY DEFAULT AS IDENTITY")
		default:
			panic("ddl: unsupported identity kind")
		}
	}
	if column.Options.Generated != nil {
		c.Write(" GENERATED ALWAYS AS (")
		c.Write(column.Options.Generated.Expr)
		c.Write(")")
		switch column.Options.Generated.Kind {
		case GeneratedStored:
			c.Write(" STORED")
		case GeneratedVirtual:
			c.Write(" VIRTUAL")
		default:
			panic("ddl: unsupported generated column kind")
		}
	}
}

func (c *compiler) compileIndexKey(k IndexKey) {
	c.Compile(k.Expr)

	if k.Options == nil {
		return
	}
	if k.Options.Length > 0 {
		c.Write("(")
		c.WriteInt(k.Options.Length)
		c.Write(")")
	}
	if k.Options.Collation != "" {
		c.Write(" COLLATE ")
		c.Write(c.renderer.QuoteIdent(k.Options.Collation))
	}
	if k.Options.OpClass != "" {
		c.Write(" ")
		c.Write(k.Options.OpClass)
	}
	if k.Options.Sort != "" {
		c.Write(" ")
		c.Write(string(k.Options.Sort))
	}
	if k.Options.Nulls != "" {
		c.Write(" NULLS ")
		c.Write(string(k.Options.Nulls))
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
	case Constraint[PrimaryKeyKind]:
		c.compilePrimaryKey(table, n)
	case Constraint[UniqueKind]:
		c.compileUnique(table, n)
	case Constraint[CheckKind]:
		c.compileCheck(table, n)
	case ForeignKeyConstraint:
		c.compileForeignKey(table, n.Constraint)
	case Constraint[ForeignKeyKind]:
		c.compileForeignKey(table, n)
	default:
		panic(fmt.Errorf("unsupported table constraint %T", constraint))
	}
}

func (c *compiler) compilePrimaryKey(table string, constraint Constraint[PrimaryKeyKind]) {
	parts := append([]string{table}, constraint.Kind.Columns...)
	name := constraintName(constraint.Name, "pk", parts...)
	c.compileNamedColumnConstraint(name, "PRIMARY KEY", constraint.Kind.Columns)
}

func (c *compiler) compileUnique(table string, constraint Constraint[UniqueKind]) {
	parts := append([]string{table}, constraint.Kind.Columns...)
	name := constraintName(constraint.Name, "uq", parts...)
	c.compileNamedColumnConstraint(name, "UNIQUE", constraint.Kind.Columns)
}

func (c *compiler) compileCheck(table string, constraint Constraint[CheckKind]) {
	var tmp strings.Builder
	compileInto(&tmp, c.renderer, constraint.Kind.Expr)
	sql := tmp.String()

	name := constraint.Name
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

func (c *compiler) compileForeignKey(table string, constraint Constraint[ForeignKeyKind]) {
	fk := constraint.Kind
	if fk.Reference == nil {
		panic("foreign key reference is required")
	}
	if len(fk.Reference.Columns) > 0 && len(fk.SourceColumns) != len(fk.Reference.Columns) {
		panic("the number of columns on the left and right must match")
	}

	parts := append([]string{table}, fk.SourceColumns...)
	parts = append(parts, fk.Reference.Table)
	parts = append(parts, fk.Reference.Columns...)
	name := constraintName(constraint.Name, "fk", parts...)

	c.compileNamedColumnConstraint(name, "FOREIGN KEY", fk.SourceColumns)
	c.Write(" REFERENCES ")
	c.Write(c.renderer.QuoteIdent(fk.Reference.Table))
	if len(fk.Reference.Columns) > 0 {
		c.Write(" (")
		c.compileColumnList(fk.Reference.Columns)
		c.Write(")")
	}
	if fk.Options != nil && fk.Options.OnDelete != nil {
		c.Write(" ON DELETE ")
		c.Write(string(*fk.Options.OnDelete))
	}
	if fk.Options != nil && fk.Options.OnUpdate != nil {
		c.Write(" ON UPDATE ")
		c.Write(string(*fk.Options.OnUpdate))
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
