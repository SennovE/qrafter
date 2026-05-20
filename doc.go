// Package qrafter builds dialect-aware SQL queries from typed Go table structs.
//
// A table model is a struct with Column fields and either a TableConfig method
// or an embedded Table tagged with table:"table_name". NewTable binds those
// fields to SQL column names, and the query builders render SQL plus driver
// arguments for a selected dialect.
//
// A typical model looks like this:
//
//	type User struct {
//		qrafter.Table `table:"users"`
//
//		ID       qrafter.Column[int] `db:"id"`
//		UserName qrafter.Column[string]
//		Age      qrafter.Column[int]
//	}
//
// SELECT construction is centered on Select. qrafter infers the FROM clause
// from the columns and predicates used in the query:
//
//	users := qrafter.MustNewTable[User]()
//
//	sql, args := qrafter.Select(users.ID, users.UserName).
//		Where(users.Age.Ge(18), users.UserName.Eq("Alice")).
//		OrderBy(users.ID.Asc()).
//		Limit(10).
//		Render(dialect.PostgreSQL{})
//
// The rendered SQL uses dialect-specific identifier quoting and placeholders.
// For PostgreSQL, the example above renders:
//
//	SELECT "users"."id", "users"."user_name" FROM "users" WHERE "users"."age" >= $1 AND "users"."user_name" = $2 ORDER BY "users"."id" ASC LIMIT 10
//
// and returns []any{18, "Alice"} as driver arguments.
//
// SELECT queries support:
//   - WHERE predicates and AND/OR composition.
//   - INNER, LEFT, RIGHT, FULL, and CROSS joins.
//   - GROUP BY and HAVING.
//   - ORDER BY with ASC, DESC, NULLS FIRST, and NULLS LAST.
//   - LIMIT and OFFSET.
//   - Parameterized values through Param and inline SQL literals through Literal.
//   - Aggregate and function expressions, including DISTINCT arguments.
//   - Window functions and frame clauses.
//   - Common table expressions, recursive CTEs, UNION, and UNION ALL.
//
// INSERT queries support:
//   - Explicit target columns through Columns.
//   - One or more VALUES rows through Values or ValuesRows.
//   - Set-style single row construction through Set.
//   - ValuesFrom for inserting values currently stored in Column fields on one
//     table model.
//   - ValuesRowsFrom for inserting values from a slice of table models.
//   - DEFAULT VALUES and per-column DEFAULT expressions.
//   - INSERT ... SELECT through FromSelect.
//   - RETURNING expressions for dialects that support them.
//
// UPDATE queries support:
//   - SET assignments through Set or SetFrom.
//   - WHERE predicates with parameterized values.
//   - FROM tables, including automatic FROM inference from SET values and
//     WHERE predicates.
//   - Common table expressions referenced from FROM, assignments, or
//     predicates.
//   - RETURNING expressions for dialects that support them.
//
// DELETE queries support:
//   - DELETE FROM a typed table model.
//   - WHERE predicates with parameterized values.
//   - USING tables, including automatic USING inference from WHERE predicates.
//   - Common table expressions referenced from USING or predicates.
//   - RETURNING expressions for dialects that support them.
//
// Columns can also scan values from database/sql rows, which lets the same
// struct describe both query construction and result destinations. The scan
// helpers work with database/sql and are friendly to sqlx struct mapping.
package qrafter
