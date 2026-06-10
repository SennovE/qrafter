package qrafter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/internal/clauses"
	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/expr"
	"github.com/SennovE/qrafter/internal/pred"
)

type compiler struct {
	w        *strings.Builder
	renderer dialect.Renderer
}

func newCompiler(d dialect.Renderer) *compiler {
	return &compiler{
		w:        &strings.Builder{},
		renderer: core.NewArgsRenderer(d),
	}
}

func newSubCompiler(d dialect.Renderer) *compiler {
	return &compiler{
		w:        &strings.Builder{},
		renderer: d,
	}
}

func (c *compiler) SQL() string {
	return c.w.String()
}

func (c *compiler) Args() []any {
	if renderer, ok := c.renderer.(interface{ Args() []any }); ok {
		return renderer.Args()
	}
	return nil
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
	if c.compileQueryNode(node) ||
		c.compileClauseNode(node) ||
		c.compileDialectNode(node) ||
		c.compileWrapperNode(node) ||
		c.compileExpressionNode(node) ||
		c.compilePredicateNode(node) ||
		c.compileWindowNode(node) {
		return
	}

	panic(fmt.Errorf("unsupported compiler node %T", node))
}

func (c *compiler) compileQueryNode(node any) bool {
	switch n := node.(type) {
	case SelectQuery:
		c.compileSelect(n)
	case InsertQuery:
		c.compileInsert(n)
	case UpdateQuery:
		c.compileUpdate(n)
	case DeleteQuery:
		c.compileDelete(n)
	case CompoundQuery:
		c.compileCompound(n)
	case CommonTableExpression:
		c.Compile(n.ref)
	case *core.CTERef:
		c.compileCTE(n)
	case core.TableRef:
		c.compileTable(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileClauseNode(node any) bool {
	switch n := node.(type) {
	case clauses.WithClause:
		c.compileWith(n)
	case clauses.SelectClause:
		c.compileSelectClause(n)
	case clauses.FromClause:
		c.compileFrom(n)
	case clauses.JoinClause:
		c.compileJoinClause(&n)
	case clauses.WhereClause:
		c.compileWhere(n)
	case clauses.GroupByClause:
		c.compileGroupBy(n)
	case clauses.HavingClause:
		c.compileHaving(n)
	case clauses.OrderByClause:
		c.compileOrderBy(n)
	case clauses.LimitOffsetClause:
		c.Compile(dialect.LimitOffset{Limit: n.Limit, Offset: n.Offset})
	default:
		return false
	}
	return true
}

func (c *compiler) compileDialectNode(node any) bool {
	switch n := node.(type) {
	case dialect.DefaultValues:
		c.compileDefaultValues()
	case dialect.Returning:
		c.compileReturningNode(n)
	case dialect.OrderItem:
		c.compileOrderItem(n)
	case dialect.Join:
		c.compileJoinNode(n)
	case dialect.LimitOffset:
		c.compileLimitOffset(n)
	case dialect.UpdateTarget:
		c.compileUpdateTargetNode(n)
	case dialect.UpdateFrom:
		c.compileUpdateFromNode(n)
	case dialect.DeleteTarget:
		c.compileDeleteTargetNode(n)
	case dialect.DeleteUsing:
		c.compileDeleteUsingNode(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileWrapperNode(node any) bool {
	switch n := node.(type) {
	case Expression:
		c.Compile(n.selecter)
	case Aggregate:
		c.Compile(n.selecter)
	case Predicate:
		c.Compile(n.predicater)
	case Order:
		c.Compile(dialect.OrderItem{Expr: n.expr, Direction: n.direction, Nulls: n.nulls})
	case ColumnRef:
		c.compileColumn(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileExpressionNode(node any) bool {
	switch n := node.(type) {
	case expr.ArgExpression:
		c.compileArg(n)
	case expr.ConstExpression:
		c.Write(c.renderer.Literal(n.Value()))
	case expr.DefaultExpression:
		c.Write("DEFAULT")
	case expr.StarExpression:
		c.Write("*")
	case expr.AliasedExpression:
		c.Compile(n.Expr())
		c.Write(" AS ")
		c.Write(c.renderer.QuoteIdent(n.Alias()))
	case expr.BinaryExpression:
		c.compileBinaryExpr(n)
	case expr.FunctionExpression:
		c.Write(n.Name())
		c.Write("(")
		c.CompileList(selectersAsAny(n.Args()), ", ")
		c.Write(")")
	case expr.DistinctExpression:
		c.Write("DISTINCT ")
		c.Compile(n.Expr())
	default:
		return false
	}
	return true
}

func (c *compiler) compilePredicateNode(node any) bool {
	switch n := node.(type) {
	case pred.BinaryPredicate:
		c.compileBinaryPred(n)
	case pred.LogicalPredicate:
		c.compileLogicalPred(n)
	default:
		return false
	}
	return true
}

func (c *compiler) compileWindowNode(node any) bool {
	switch n := node.(type) {
	case WindowSpec:
		c.compileWindowSpec(n)
	case WindowFrame:
		c.compileWindowFrame(n)
	case WindowFrameBound:
		c.Write(n.value)
	case windowExpression:
		c.Compile(n.expr)
		c.Write(" OVER ")
		c.Compile(n.spec)
	default:
		return false
	}
	return true
}

func (c *compiler) compileSelect(q SelectQuery) {
	state := q.currentState()
	c.Compile(state.selectCl)
	c.Compile(state.fromCl)
	c.Compile(state.whereCl)
	c.Compile(state.groupByCl)
	c.Compile(state.havingCl)
	c.Compile(state.orderByCl)
	c.Compile(state.limitOffsetCl)
}

func (c *compiler) compileInsert(q InsertQuery) {
	state := q.currentState()

	c.Write("INSERT INTO ")
	c.Compile(state.table)
	c.compileInsertColumns(state.columns)

	switch {
	case state.defaultValues || state.source == nil && len(state.rows) == 0:
		c.Compile(dialect.DefaultValues{})
	case state.source != nil:
		c.Write("\n")
		c.compileQueryExpression(state.source)
	default:
		c.Write("\nVALUES ")
		c.compileInsertRows(state.rows)
	}

	c.compileReturning(state.returning)
}

func (c *compiler) compileUpdate(q UpdateQuery) {
	state := q.currentState()
	c.Compile(dialect.UpdateTarget{
		Target: state.table,
		From:   tableRefsAsAny(state.from),
	})
	c.compileUpdateAssignments(state.assignments)
	c.Compile(dialect.UpdateFrom{From: tableRefsAsAny(state.from)})
	c.Compile(state.whereCl)
	c.compileReturning(state.returning)
}

func (c *compiler) compileDelete(q DeleteQuery) {
	state := q.currentState()
	c.Compile(dialect.DeleteTarget{
		Target:     state.table,
		TargetName: state.table.SQLName(),
		Using:      tableRefsAsAny(state.using),
	})
	c.Compile(dialect.DeleteUsing{Using: tableRefsAsAny(state.using)})
	c.Compile(state.whereCl)
	c.compileReturning(state.returning)
}

func (c *compiler) compileCompound(q CompoundQuery) {
	state := q.currentState()
	c.compileSetOperand(state.left)
	c.Write("\n")
	c.Write(state.operator.String())
	c.Write("\n")
	c.compileSetOperand(state.right)
	c.Compile(state.orderByCl)
	c.Compile(state.limitOffsetCl)
}

func (c *compiler) compileQueryExpression(q core.QueryExpression) {
	switch q := q.(type) {
	case SelectQuery:
		c.compileSelect(q)
	case CompoundQuery:
		c.compileCompound(q)
	case CommonTableExpression:
		c.compileQueryExpression(q.ref.Query)
	default:
		panic(fmt.Errorf("unsupported query expression %T", q))
	}
}

func (c *compiler) compileSetOperand(q core.QueryExpression) {
	switch q := q.(type) {
	case SelectQuery:
		state := q.currentState()
		if len(state.orderByCl.Items) > 0 || state.limitOffsetCl.Limit != 0 || state.limitOffsetCl.Offset != 0 {
			c.Write("(")
			c.compileSelect(q)
			c.Write(")")
			return
		}
		c.compileSelect(q)
	case CompoundQuery:
		c.Write("(")
		c.compileCompound(q)
		c.Write(")")
	case CommonTableExpression:
		c.compileSetOperand(q.ref.Query)
	default:
		panic(fmt.Errorf("unsupported set operand %T", q))
	}
}

func (c *compiler) compileWith(cl clauses.WithClause) {
	if len(cl.CTEs) == 0 {
		return
	}

	c.Write("WITH ")
	if cl.Recursive {
		c.Write("RECURSIVE ")
	}
	c.CompileList(cteRefsAsAny(cl.CTEs), ",\n")
	c.Write("\n")
}

func (c *compiler) compileCTE(cte *core.CTERef) {
	if cte == nil {
		return
	}

	c.Write(c.renderer.QuoteIdent(cte.Name))
	if len(cte.Columns) > 0 {
		c.Write(" (")
		for i, column := range cte.Columns {
			if i > 0 {
				c.Write(", ")
			}
			c.Write(c.renderer.QuoteIdent(column))
		}
		c.Write(")")
	}

	body := newSubCompiler(c.renderer)
	body.compileQueryExpression(cte.Query)

	c.Write(" AS (\n")
	c.writeIndentedLines(body.SQL(), "    ")
	c.Write("\n)")
}

func (c *compiler) compileSelectClause(cl clauses.SelectClause) {
	c.Write("SELECT ")
	c.CompileList(selectersAsAny(cl.Columns), ", ")
}

func (c *compiler) compileFrom(cl clauses.FromClause) {
	if len(cl.Tables) == 0 && len(cl.Joins) == 0 {
		return
	}

	c.Write("\nFROM ")
	tables := core.GetSortedTables(cl.Tables)
	joins := cl.Joins
	if len(tables) == 0 {
		c.Compile(joins[0].Table)
		joins = joins[1:]
	} else {
		c.CompileList(tableRefsAsAny(tables), ", ")
	}
	for i := range joins {
		c.Compile(joins[i])
	}
}

func (c *compiler) compileJoinClause(cl *clauses.JoinClause) {
	c.Compile(dialect.Join{
		Type:       cl.Type,
		Table:      cl.Table,
		Predicates: predicatersAsAny(cl.Predicates),
	})
}

func (c *compiler) compileWhere(cl clauses.WhereClause) {
	if len(cl.Predicates) == 0 {
		return
	}

	c.Write("\nWHERE ")
	c.compilePredicates(cl.Predicates)
}

func (c *compiler) compileGroupBy(cl clauses.GroupByClause) {
	if len(cl.Columns) == 0 {
		return
	}

	c.Write("\nGROUP BY ")
	c.CompileList(selectersAsAny(cl.Columns), ", ")
}

func (c *compiler) compileHaving(cl clauses.HavingClause) {
	if len(cl.Predicates) == 0 {
		return
	}

	c.Write("\nHAVING ")
	c.compilePredicates(cl.Predicates)
}

func (c *compiler) compileOrderBy(cl clauses.OrderByClause) {
	if len(cl.Items) == 0 {
		return
	}

	c.Write("\nORDER BY ")
	c.CompileList(selectersAsAny(cl.Items), ", ")
}

func (c *compiler) compileDefaultValues() {
	c.Write("\nDEFAULT VALUES")
}

func (c *compiler) compileReturning(returning []core.Selecter) {
	if len(returning) == 0 {
		return
	}
	c.Compile(dialect.Returning{Items: selectersAsAny(returning)})
}

func (c *compiler) compileReturningNode(n dialect.Returning) {
	c.Write("\nRETURNING ")
	c.CompileList(n.Items, ", ")
}

func (c *compiler) compileOrderItem(n dialect.OrderItem) {
	c.Compile(n.Expr)
	if n.Direction != "" {
		c.Write(" ")
		c.Write(n.Direction)
	}
	if n.Nulls != "" {
		c.Write(" NULLS ")
		c.Write(n.Nulls)
	}
}

func (c *compiler) compileJoinNode(n dialect.Join) {
	c.Write("\n")
	c.Write(n.Type)
	c.Write(" ")
	c.Compile(n.Table)
	if len(n.Predicates) == 0 {
		return
	}
	c.Write(" ON ")
	c.compileAnyPredicates(n.Predicates)
}

func (c *compiler) compileLimitOffset(n dialect.LimitOffset) {
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
		c.Write("\nOFFSET ")
		c.WriteInt(n.Offset)
	}
}

func (c *compiler) compileUpdateTargetNode(n dialect.UpdateTarget) {
	c.Write("UPDATE ")
	c.Compile(n.Target)
}

func (c *compiler) compileUpdateFromNode(n dialect.UpdateFrom) {
	if len(n.From) == 0 {
		return
	}
	c.Write("\nFROM ")
	c.CompileList(n.From, ", ")
}

func (c *compiler) compileDeleteTargetNode(n dialect.DeleteTarget) {
	c.Write("DELETE FROM ")
	c.Compile(n.Target)
}

func (c *compiler) compileDeleteUsingNode(n dialect.DeleteUsing) {
	if len(n.Using) == 0 {
		return
	}
	c.Write("\nUSING ")
	c.CompileList(n.Using, ", ")
}

func (c *compiler) compileTable(t core.TableRef) {
	c.Write(c.renderer.QuoteIdent(t.Name))
	if t.Alias != "" {
		c.Write(" AS ")
		c.Write(c.renderer.QuoteIdent(t.Alias))
	}
}

func (c *compiler) compileColumn(col ColumnRef) {
	c.Write(c.renderer.QuoteIdent(col.TableRef().SQLName()))
	c.Write(".")
	c.Write(c.renderer.QuoteIdent(col.Name()))
}

func (c *compiler) compileArg(arg expr.ArgExpression) {
	if renderer, ok := c.renderer.(core.ArgRenderer); ok {
		c.Write(renderer.AddArg(arg.Value()))
		return
	}
	c.Write(c.renderer.Literal(arg.Value()))
}

func (c *compiler) compileBinaryExpr(e expr.BinaryExpression) {
	c.compileChild(e.Left(), e.Precedence(), false)
	c.Write(" ")
	c.Write(e.Op())
	c.Write(" ")
	c.compileChild(e.Right(), e.Precedence(), parenthesizeRightPeer(e.Op()))
}

func (c *compiler) compileBinaryPred(p pred.BinaryPredicate) {
	c.compileChild(p.Left(), p.Precedence(), false)
	c.Write(" ")
	c.Write(p.Op())
	c.Write(" ")
	c.compileChild(p.Right(), p.Precedence(), false)
}

func (c *compiler) compileLogicalPred(p pred.LogicalPredicate) {
	predicates := p.Predicates()
	for i, item := range predicates {
		if i > 0 {
			c.Write(" ")
			c.Write(p.Op())
			c.Write(" ")
		}
		c.compileChild(item, p.Precedence(), false)
	}
}

func (c *compiler) compilePredicates(predicates []core.Predicater) {
	if len(predicates) == 1 {
		c.Compile(predicates[0])
		return
	}
	c.compileLogicalPred(pred.Logical(pred.OpAnd, predicates...))
}

func (c *compiler) compileAnyPredicates(predicates []any) {
	if len(predicates) == 1 {
		c.Compile(predicates[0])
		return
	}
	for i, item := range predicates {
		if i > 0 {
			c.Write(" AND ")
		}
		c.Compile(item)
	}
}

func (c *compiler) compileChild(node any, parentPrecedence int, parenthesizeOnEqual bool) {
	child, ok := node.(core.Precedencer)
	if !ok {
		c.Compile(node)
		return
	}

	childPrecedence := child.Precedence()
	if childPrecedence < parentPrecedence || childPrecedence == parentPrecedence && parenthesizeOnEqual {
		c.Write("(")
		c.Compile(node)
		c.Write(")")
		return
	}
	c.Compile(node)
}

func (c *compiler) compileWindowSpec(spec WindowSpec) {
	c.Write("(")

	rendered := false
	if len(spec.partitionBy) > 0 {
		c.Write("PARTITION BY ")
		c.CompileList(selectersAsAny(spec.partitionBy), ", ")
		rendered = true
	}
	if len(spec.orderBy) > 0 {
		if rendered {
			c.Write(" ")
		}
		c.Write("ORDER BY ")
		c.CompileList(selectersAsAny(spec.orderBy), ", ")
		rendered = true
	}
	if spec.frame != nil {
		if rendered {
			c.Write(" ")
		}
		c.Compile(*spec.frame)
	}

	c.Write(")")
}

func (c *compiler) compileWindowFrame(frame WindowFrame) {
	c.Write(frame.mode)
	if frame.end != nil {
		c.Write(" BETWEEN ")
		c.Compile(frame.start)
		c.Write(" AND ")
		c.Compile(*frame.end)
		return
	}
	c.Write(" ")
	c.Compile(frame.start)
}

func (c *compiler) compileInsertColumns(columns []ColumnRef) {
	if len(columns) == 0 {
		return
	}

	c.Write(" (")
	for i, column := range columns {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.renderer.QuoteIdent(column.Name()))
	}
	c.Write(")")
}

func (c *compiler) compileInsertRows(rows [][]core.Selecter) {
	for i, row := range rows {
		if i > 0 {
			c.Write(", ")
		}
		c.Write("(")
		c.CompileList(selectersAsAny(row), ", ")
		c.Write(")")
	}
}

func (c *compiler) compileUpdateAssignments(assignments []updateAssignment) {
	c.Write("\nSET ")
	for i, assignment := range assignments {
		if i > 0 {
			c.Write(", ")
		}
		c.Write(c.renderer.QuoteIdent(assignment.column.Name()))
		c.Write(" = ")
		c.Compile(assignment.value)
	}
}

func (c *compiler) writeIndentedLines(s, indent string) {
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			c.Write("\n")
		}
		if line == "" {
			continue
		}
		c.Write(indent)
		c.Write(line)
	}
}

func parenthesizeRightPeer(op string) bool {
	switch op {
	case "-", "/", "%":
		return true
	default:
		return false
	}
}

func selectersAsAny(items []core.Selecter) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}

func predicatersAsAny(items []core.Predicater) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}

func tableRefsAsAny(items []core.TableRef) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}

func cteRefsAsAny(items []*core.CTERef) []any {
	nodes := make([]any, len(items))
	for i, item := range items {
		nodes[i] = item
	}
	return nodes
}
