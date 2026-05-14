package tests

import (
	"testing"

	"github.com/SennovE/qrafter"
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

		assert.Equal(t, "table", u.UserName.Table.Name)
		assert.Equal(t, "user_name", u.UserName.Name)
		assert.Equal(t, "table", u.Age.Table.Name)
		assert.Equal(t, "userAge", u.Age.Name)
	})

	t.Run("MustNewTable binds columns and panics on error", func(t *testing.T) {
		u := qrafter.MustNewTable[User]()

		assert.Equal(t, "table", u.UserName.Table.Name)
		assert.Equal(t, "user_name", u.UserName.Name)
	})
}

func TestTable_MakeAlias(t *testing.T) {
	u, err := qrafter.NewTable[User]()

	alias := "alias"
	aliased, err := qrafter.TableAlias(u, alias)
	require.NoError(t, err)

	t.Run("Table reference is set with alias", func(t *testing.T) {
		assert.Equal(t, u.TableConfig().Name, aliased.UserName.Table.Name)
		assert.Equal(t, alias, aliased.UserName.Table.Alias)
		assert.Equal(t, u.TableConfig().Name, aliased.Age.Table.Name)
		assert.Equal(t, alias, aliased.Age.Table.Alias)
	})
}
