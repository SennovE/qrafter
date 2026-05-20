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
	UserName qrafter.Column[string]
	Age      qrafter.Column[string] `db:"userAge"`

	Other string
	meta  string
}

func (User) TableConfig() qrafter.TableConfig {
	return qrafter.TableConfig{
		Name: "table",
	}
}

type EmbeddedConfigUser struct {
	qrafter.Table `table:"embedded_users"`

	ID   qrafter.Column[int]
	Name qrafter.Column[string] `db:"full_name"`
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

	t.Run("NewTable accepts embedded Table", func(t *testing.T) {
		u, err := qrafter.NewTable[EmbeddedConfigUser]()
		require.NoError(t, err, "NewTable should not return an error")

		assert.Equal(t, "embedded_users", u.TableConfig().Name)
		checkRenderedColumn(t, u.TableConfig().Name, "id", u.ID)
		checkRenderedColumn(t, u.TableConfig().Name, "full_name", u.Name)
	})
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

func TestTable_MakeAliasWithEmbeddedConfig(t *testing.T) {
	u, err := qrafter.NewTable[EmbeddedConfigUser]()
	require.NoError(t, err)

	alias := "embedded_alias"
	aliased, err := qrafter.TableAlias(u, alias)
	require.NoError(t, err)

	t.Run("Table reference is set with alias", func(t *testing.T) {
		assert.Equal(t, "embedded_users", aliased.TableConfig().Name)
		checkRenderedColumn(t, alias, "id", aliased.ID)
		checkRenderedColumn(t, alias, "full_name", aliased.Name)
	})
}

func checkRenderedColumn[T any](t *testing.T, table, name string, expr qrafter.Column[T]) {
	t.Helper()

	expected := fmt.Sprintf(`"%s"."%s"`, table, name)

	var w strings.Builder

	expr.Render(&w, dialect.PostgreSQL{})
	assert.Equal(t, expected, w.String())
}
