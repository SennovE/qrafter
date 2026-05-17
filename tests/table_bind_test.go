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

func checkRenderedColumn[T any](t *testing.T, table, name string, expr qrafter.Column[T]) {
	t.Helper()

	expected := fmt.Sprintf(`"%s"."%s"`, table, name)

	var w strings.Builder

	expr.Render(&w, dialect.PostgreSQL{})
	assert.Equal(t, expected, w.String())
}
