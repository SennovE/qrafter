package ddl

const (
	precedenceOr = iota + 1
	precedenceAnd
	precedenceComparison
	precedenceAdditive
	precedenceMultiplicative
	precedenceValue
)

// Expression represents a SQL value expression inside a DDL predicate.
type Expression struct {
	node any
	prec int
}

// IsZero reports whether the expression has not been initialized.
func (e Expression) IsZero() bool {
	return e.node == nil
}

type columnExpression struct {
	name string
}

type literalExpression struct {
	value any
}

type rawExpression struct {
	sql string
}

type functionExpression struct {
	name string
	args []Expression
}

type binaryExpression struct {
	left                  Expression
	op                    string
	right                 Expression
	prec                  int
	parenthesizeRightPeer bool
}

// Predicate represents a SQL boolean predicate inside DDL.
type Predicate struct {
	node any
	prec int
}

type binaryPredicate struct {
	left  Expression
	op    string
	right Expression
}

type logicalPredicate struct {
	op         string
	predicates []Predicate
	prec       int
}

type rawPredicate struct {
	sql string
}

func expression(node any, prec int) Expression {
	return Expression{node: node, prec: prec}
}

func predicate(node any, prec int) Predicate {
	return Predicate{node: node, prec: prec}
}

func asExpression(v any) Expression {
	if expr, ok := v.(Expression); ok {
		return expr
	}
	return Literal(v)
}

func (e Expression) binary(op string, v any, prec int, parenthesizeRightPeer bool) Expression {
	return expression(binaryExpression{
		left:                  e,
		op:                    op,
		right:                 asExpression(v),
		prec:                  prec,
		parenthesizeRightPeer: parenthesizeRightPeer,
	}, prec)
}

// Add returns an addition expression.
func (e Expression) Add(v any) Expression { return e.binary("+", v, precedenceAdditive, false) }

// Sub returns a subtraction expression.
func (e Expression) Sub(v any) Expression { return e.binary("-", v, precedenceAdditive, true) }

// Mul returns a multiplication expression.
func (e Expression) Mul(v any) Expression { return e.binary("*", v, precedenceMultiplicative, false) }

// Div returns a division expression.
func (e Expression) Div(v any) Expression { return e.binary("/", v, precedenceMultiplicative, true) }

// Literal returns an expression rendered inline using the dialect's literal rules.
func Literal(v any) Expression {
	return expression(literalExpression{value: v}, precedenceValue)
}

// Func builds a SQL function call expression.
func Func(name string, args ...any) Expression {
	exprs := make([]Expression, len(args))
	for i, arg := range args {
		exprs[i] = asExpression(arg)
	}
	return expression(functionExpression{name: name, args: exprs}, precedenceValue)
}

// Col creates an unqualified column reference for DDL predicates.
func Col(name string) Expression {
	return expression(columnExpression{name: name}, precedenceValue)
}

// And combines predicates with SQL AND.
func And(ps ...Predicate) Predicate {
	return logical("AND", precedenceAnd, ps)
}

// Or combines predicates with SQL OR.
func Or(ps ...Predicate) Predicate {
	return logical("OR", precedenceOr, ps)
}

func logical(op string, prec int, ps []Predicate) Predicate {
	return predicate(logicalPredicate{op: op, predicates: append([]Predicate(nil), ps...), prec: prec}, prec)
}

func (e Expression) compare(op string, v any) Predicate {
	return predicate(binaryPredicate{left: e, op: op, right: asExpression(v)}, precedenceComparison)
}

// Lt returns a less-than predicate.
func (e Expression) Lt(v any) Predicate { return e.compare("<", v) }

// Gt returns a greater-than predicate.
func (e Expression) Gt(v any) Predicate { return e.compare(">", v) }

// Le returns a less-than-or-equal predicate.
func (e Expression) Le(v any) Predicate { return e.compare("<=", v) }

// Ge returns a greater-than-or-equal predicate.
func (e Expression) Ge(v any) Predicate { return e.compare(">=", v) }

// Eq returns an equality predicate.
func (e Expression) Eq(v any) Predicate { return e.compare("=", v) }

// Like returns a LIKE predicate.
func (e Expression) Like(v any) Predicate { return e.compare("LIKE", v) }

// NotLike returns a NOT LIKE predicate.
func (e Expression) NotLike(v any) Predicate { return e.compare("NOT LIKE", v) }

// IsNull returns an IS NULL predicate.
func (e Expression) IsNull() Predicate { return e.compare("IS", Literal(nil)) }

// IsNotNull returns an IS NOT NULL predicate.
func (e Expression) IsNotNull() Predicate { return e.compare("IS NOT", Literal(nil)) }

// RawExpr returns a SQL expression rendered verbatim.
func RawExpr(sql string) Expression {
	return expression(rawExpression{sql: sql}, precedenceValue)
}

// RawPred returns a SQL predicate rendered verbatim.
func RawPred(sql string) Predicate {
	return predicate(rawPredicate{sql: sql}, precedenceValue)
}
