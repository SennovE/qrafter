package tests

import (
	"database/sql"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSQLiteUpdate(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	sqlText, args := q.
		Update(users).
		Set(users.UserName, "Alicia").
		Set(users.NickName, sql.NullString{String: "Ace", Valid: true}).
		Where(users.ID.Eq(1)).
		MustRender(dialect.SQLite{})

	result, err := db.Exec(sqlText, args...)
	require.NoError(t, err)

	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	var got sqliteUser
	require.NoError(
		t,
		db.QueryRow(
			`SELECT id, user_name, nick_name FROM users WHERE id = ?`,
			1,
		).Scan(&got.ID, &got.UserName, &got.NickName),
	)

	assert.Equal(t, 1, got.ID.Get())
	assert.Equal(t, "Alicia", got.UserName.Get())
	assert.Equal(t, sql.NullString{String: "Ace", Valid: true}, got.NickName.Get())
}
