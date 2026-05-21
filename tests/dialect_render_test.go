package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
)

func TestMySQLDialectRender(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(users.UserName).
		Where(users.Age.Ge("18")).
		Limit(10).
		Offset(20).
		MustRender(dialect.MySQL{})

	assert.Equal(t, "SELECT `table`.`user_name`\nFROM `table`\nWHERE `table`.`userAge` >= ?\nLIMIT 20, 10", sql)
	assert.Equal(t, []any{"18"}, args)
	assert.Equal(t, "`weird``name`", dialect.MySQL{}.QuoteIdent("weird`name"))
}

func TestMySQLDialectRender_OffsetWithoutLimit(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(users.UserName).
		Offset(20).
		MustRender(dialect.MySQL{})

	assert.Equal(
		t,
		"SELECT `table`.`user_name`\nFROM `table`\nLIMIT 18446744073709551615 OFFSET 20",
		sql,
	)
	assert.Empty(t, args)
}

func TestMySQLDialectRender_DefaultValues(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Insert(users).
		DefaultValues().
		MustRender(dialect.MySQL{})

	assert.Equal(t, "INSERT INTO `table` ()\nVALUES ()", sql)
	assert.Empty(t, args)
}

func TestMySQLDialectRender_UpdateFrom(t *testing.T) {
	users := q.MustNewTable[User]()
	managers, err := q.TableAlias(users, "manager")
	assert.NoError(t, err)

	sql, args := q.Update(users).
		Set(users.Age, managers.Age).
		From(managers).
		Where(users.UserName.Eq(managers.UserName)).
		MustRender(dialect.MySQL{})

	assert.Equal(t, "UPDATE `table`, `table` AS `manager`\nSET `userAge` = `manager`.`userAge`\nWHERE `table`.`user_name` = `manager`.`user_name`", sql)
	assert.Empty(t, args)
}

func TestMySQLDialectRender_DeleteUsing(t *testing.T) {
	users := q.MustNewTable[User]()
	managers, err := q.TableAlias(users, "manager")
	assert.NoError(t, err)

	sql, args := q.Delete(users).
		Using(managers).
		Where(users.Age.Eq(managers.Age), managers.UserName.Eq("Bob")).
		MustRender(dialect.MySQL{})

	assert.Equal(t, "DELETE `table`\nFROM `table`, `table` AS `manager`\nWHERE `table`.`userAge` = `manager`.`userAge` AND `manager`.`user_name` = ?", sql)
	assert.Equal(t, []any{"Bob"}, args)
}

func TestMySQLDialectRender_NullsOrdering(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(users.UserName).
		OrderBy(users.Age.Asc().NullsLast()).
		MustRender(dialect.MySQL{})

	assert.Equal(t, "SELECT `table`.`user_name`\nFROM `table`\nORDER BY `table`.`userAge` IS NULL, `table`.`userAge` ASC", sql)
	assert.Empty(t, args)
}

func TestMySQLDialectRender_UnsupportedReturning(t *testing.T) {
	users := q.MustNewTable[User]()

	query := q.Insert(users).
		Columns(users.UserName).
		Values("Alice").
		Returning(users.UserName)

	t.Run("MySQL dialect does not support RETURNING. Panic on MustRender call.", func(t *testing.T) {
		assert.PanicsWithError(t, "MySQL dialect does not support RETURNING", func() {
			query.MustRender(dialect.MySQL{})
		})
	})

	t.Run("MySQL dialect does not support RETURNING. Get error on Render call.", func(t *testing.T) {
		_, _, err := query.Render(dialect.MySQL{})
		var unsupported dialect.UnsupportedFeatureError
		assert.ErrorAs(t, err, &unsupported)
		assert.Equal(t, "MySQL", unsupported.Dialect)
		assert.Equal(t, "RETURNING", unsupported.Feature)
	})
}

func TestMySQLDialectRender_UnsupportedFullJoin(t *testing.T) {
	users := q.MustNewTable[User]()
	managers, err := q.TableAlias(users, "manager")
	assert.NoError(t, err)

	query := q.Select(users.UserName).FullJoin(managers, users.Age.Eq(managers.Age))

	t.Run("MySQL dialect does not support FULL JOIN. Panic on MustRender call.", func(t *testing.T) {
		assert.PanicsWithError(t, "MySQL dialect does not support FULL JOIN", func() {
			query.MustRender(dialect.MySQL{})
		})
	})

	t.Run("MySQL dialect does not support FULL JOIN. Get error on Render call.", func(t *testing.T) {
		_, _, err := query.Render(dialect.MySQL{})
		var unsupported dialect.UnsupportedFeatureError
		assert.ErrorAs(t, err, &unsupported)
		assert.Equal(t, "MySQL", unsupported.Dialect)
		assert.Equal(t, "FULL JOIN", unsupported.Feature)
	})
}

func TestSQLiteDialectRender(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args, err := q.Select(users.UserName).
		Offset(20).
		Render(dialect.SQLite{})

	assert.Equal(t, `SELECT "table"."user_name"
FROM "table"
LIMIT -1 OFFSET 20`, sql)
	assert.Empty(t, args)
	assert.NoError(t, err)
}

func TestSQLiteDialectRender_BoolLiteral(t *testing.T) {
	sql, args, err := q.Select(q.Literal(true), q.Literal(false)).
		Render(dialect.SQLite{})

	assert.Equal(t, "SELECT 1, 0", sql)
	assert.Empty(t, args)
	assert.NoError(t, err)
}

func TestSQLiteDialectRender_UnsupportedDeleteUsing(t *testing.T) {
	users := q.MustNewTable[User]()
	managers, err := q.TableAlias(users, "manager")
	assert.NoError(t, err)

	_, _, err = q.Delete(users).
		Using(managers).
		Where(users.Age.Eq(managers.Age)).
		Render(dialect.SQLite{})

	var unsupported dialect.UnsupportedFeatureError
	assert.ErrorAs(t, err, &unsupported)
}
