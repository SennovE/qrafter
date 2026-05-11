package tests

import (
	"testing"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/internal/core"
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

func TestTable_Bind(t *testing.T) {
	u := User{}
	err := qrafter.Bind(&u)
	require.NoError(t, err, "Bind should not return an error")

	t.Run("Table reference is set", func(t *testing.T) {
		expectedTable := core.TableRef{
			Name:  u.TableConfig().Name,
			Alias: "",
		}
		assert.Equal(t, expectedTable, u.UserName.Table)
		assert.Equal(t, expectedTable, u.Age.Table)
	})

	t.Run("Column name is snake_case of struct field", func(t *testing.T) {
		assert.Equal(t, "user_name", u.UserName.Name)
	})

	t.Run("Column name is taken from db tag when present", func(t *testing.T) {
		assert.Equal(t, "userAge", u.Age.Name)
	})
}

func TestTable_MakeAlias(t *testing.T) {
	u := User{}
	err := qrafter.Bind(&u)
	require.NoError(t, err)

	alias := "alias"
	aliased, err := qrafter.TableAlias(u, alias)
	require.NoError(t, err)

	t.Run("Table reference is set with alias", func(t *testing.T) {
		expectedTable := core.TableRef{
			Name:  u.TableConfig().Name,
			Alias: alias,
		}
		assert.Equal(t, expectedTable, aliased.UserName.Table)
		assert.Equal(t, expectedTable, aliased.Age.Table)
	})
}
