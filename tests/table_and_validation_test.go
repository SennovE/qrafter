package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MissingTableName struct {
	ID q.Column[int]
}

func (MissingTableName) TableConfig() q.TableConfig {
	return q.TableConfig{}
}

type NonStructTable string

func (NonStructTable) TableConfig() q.TableConfig {
	return q.TableConfig{Name: "bad"}
}

func TestTableAliasAndPointerGetTableRef(t *testing.T) {
	users := q.MustNewTable[User]()
	aliased, err := q.TableAlias(users, "u")
	require.NoError(t, err)

	ref := q.GetTableRef(&aliased)
	assert.Equal(t, "table", ref.Name)
	assert.Equal(t, "u", ref.Alias)

	sql, args := q.Select(aliased.UserName, aliased.Age).MustRender(dialect.PostgreSQL{})
	assert.Equal(t, `SELECT "u"."user_name", "u"."userAge"
FROM "table" AS "u"`, sql)
	assert.Empty(t, args)
}

func TestNewTableErrors(t *testing.T) {
	_, err := q.NewTable[MissingTableName]()
	assert.EqualError(t, err, "table name is empty")

	_, err = q.NewTable[NonStructTable]()
	assert.EqualError(t, err, "type T must be a struct, got string")

	assert.PanicsWithError(t, "table name is empty", func() {
		_ = q.MustNewTable[MissingTableName]()
	})
}

func TestInsertValuesRowsFromIgnoresInvalidInputs(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Insert(users).
		Columns(users.UserName).
		ValuesRowsFrom(nil).
		ValuesRowsFrom("not rows").
		ValuesRows([][]any{}).
		Values("Alice").
		MustRender(dialect.PostgreSQL{})

	assert.Equal(t, `INSERT INTO "table" ("user_name")
VALUES ($1)`, sql)
	assert.Equal(t, []any{"Alice"}, args)
}

func TestDDLValidationBranches(t *testing.T) {
	assert.PanicsWithError(t, "varchar size must be positive", func() { _ = ddl.VarChar(0) })
	assert.PanicsWithError(t, "char size must be positive", func() { _ = ddl.Char(0) })
	assert.PanicsWithError(t, "numeric precision must be positive", func() { _ = ddl.Numeric(0, 0) })
	assert.PanicsWithError(t, "numeric scale cannot be negative", func() { _ = ddl.Numeric(1, -1) })
	assert.PanicsWithValue(t, "ddl: unsupported identity kind", func() {
		_ = ddl.Column("id", ddl.Integer()).Identity("sometimes")
	})
	assert.PanicsWithValue(t, "ddl: unsupported generated column kind", func() {
		_ = ddl.Column("slug", ddl.Text()).Generated("remote", "lower(name)")
	})

	_, err := ddl.CreateTable("bad").
		Columns(ddl.Column("id", ddl.Integer()).IdentityAlways().Default(1)).
		Render(dialect.PostgreSQL{})
	assert.EqualError(t, err, "ddl: identity column cannot have DEFAULT")

	_, err = ddl.CreateTable("bad").
		Columns(ddl.Column("slug", ddl.Text()).GeneratedStored("lower(name)").Default("x")).
		Render(dialect.PostgreSQL{})
	assert.EqualError(t, err, "ddl: generated column cannot have DEFAULT")

	_, err = ddl.CreateTable("bad").
		Constraints(ddl.ForeignKey("org_id", "tenant_id").References("orgs", "id")).
		Render(dialect.PostgreSQL{})
	assert.EqualError(t, err, "the number of columns on the left and right must match")

	_, err = ddl.CreateTable("bad").
		Columns(ddl.Column("mystery", ddl.Type{})).
		Render(dialect.PostgreSQL{})
	assert.EqualError(t, err, "ddl type is empty")
}
