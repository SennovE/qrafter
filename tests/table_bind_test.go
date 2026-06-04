package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	qrafter.Table `table:"table"`

	UserName qrafter.Column[string]
	Age      qrafter.Column[string] `db:"userAge"`

	Other string
	meta  string
}

type ExplicitConfigUser struct {
	ID   qrafter.Column[int]
	Name qrafter.Column[string] `db:"full_name"`
}

type CustomMappedUser struct {
	qrafter.Table `table:"mapped_users"`

	UserName qrafter.Column[string]
	Age      qrafter.Column[int] `db:"age_years"`
}

func (ExplicitConfigUser) TableConfig() qrafter.TableConfig {
	return qrafter.TableConfig{
		Name: "explicit_users",
	}
}

func TestTable_NewTable(t *testing.T) {
	t.Run("NewTable binds columns automatically", func(t *testing.T) {
		u, err := qrafter.NewTable[User]()
		require.NoError(t, err, "NewTable should not return an error")

		checkRenderedColumn(t, u.TableConfig().Name, "user_name", u.UserName)
		checkRenderedColumn(t, u.TableConfig().Name, "userAge", u.Age)
	})

	t.Run("MustNewTable binds columns and panics on error", func(t *testing.T) {
		u := qrafter.MustNewTable[User]()

		checkRenderedColumn(t, u.TableConfig().Name, "user_name", u.UserName)
		checkRenderedColumn(t, u.TableConfig().Name, "userAge", u.Age)
	})

	t.Run("NewTable accepts explicit TableConfig method", func(t *testing.T) {
		u, err := qrafter.NewTable[ExplicitConfigUser]()
		require.NoError(t, err, "NewTable should not return an error")

		assert.Equal(t, "explicit_users", u.TableConfig().Name)
		checkRenderedColumn(t, u.TableConfig().Name, "id", u.ID)
		checkRenderedColumn(t, u.TableConfig().Name, "full_name", u.Name)
	})
}

func TestTable_NameMapper(t *testing.T) {
	original := qrafter.NameMapper
	qrafter.NameMapper = func(field string) string {
		return "mapped_" + strings.ToLower(field)
	}
	t.Cleanup(func() {
		qrafter.NameMapper = original
	})

	u, err := qrafter.NewTable[CustomMappedUser]()
	require.NoError(t, err)

	checkRenderedColumn(t, u.TableConfig().Name, "mapped_username", u.UserName)
	checkRenderedColumn(t, u.TableConfig().Name, "age_years", u.Age)
}

func TestTable_MakeAlias(t *testing.T) {
	u, err := qrafter.NewTable[User]()
	require.NoError(t, err)

	alias := "alias"
	aliased, err := qrafter.TableAlias(u, alias)
	require.NoError(t, err)

	t.Run("Table reference is set with alias", func(t *testing.T) {
		checkRenderedColumn(t, alias, "user_name", aliased.UserName)
		checkRenderedColumn(t, alias, "userAge", aliased.Age)
	})
}

func TestTable_MakeAliasWithExplicitConfig(t *testing.T) {
	u, err := qrafter.NewTable[ExplicitConfigUser]()
	require.NoError(t, err)

	alias := "explicit_alias"
	aliased, err := qrafter.TableAlias(u, alias)
	require.NoError(t, err)

	t.Run("Table reference is set with alias", func(t *testing.T) {
		assert.Equal(t, "explicit_users", aliased.TableConfig().Name)
		checkRenderedColumn(t, alias, "id", aliased.ID)
		checkRenderedColumn(t, alias, "full_name", aliased.Name)
	})
}

func checkRenderedColumn[T any](t *testing.T, table, name string, expr qrafter.Column[T]) {
	t.Helper()

	expected := fmt.Sprintf("%q.%q", table, name)

	var w strings.Builder

	expr.Render(&w, dialect.PostgreSQL{})
	assert.Equal(t, expected, w.String())
}
