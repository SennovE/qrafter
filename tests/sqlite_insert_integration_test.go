package tests

import (
	"database/sql"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSQLiteInsertValuesFrom(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	users.ID.Set(3)
	users.UserName.Set("Carol")
	users.NickName.Set(sql.NullString{String: "C", Valid: true})

	sqlText, args := q.
		InsertInto(users).
		ValuesFrom(users).
		Render(dialect.BaseDialect{})

	_, err := db.Exec(sqlText, args...)
	require.NoError(t, err)

	var got sqliteUser
	require.NoError(
		t,
		db.QueryRow(
			`SELECT id, user_name, nick_name FROM users WHERE id = ?`,
			3,
		).Scan(&got.ID, &got.UserName, &got.NickName),
	)

	assert.Equal(t, 3, got.ID.Get())
	assert.Equal(t, "Carol", got.UserName.Get())
	assert.Equal(t, sql.NullString{String: "C", Valid: true}, got.NickName.Get())
}

func TestIntegrationSQLiteInsertValuesRowsFromSlice(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	carol := bindSQLiteUsers(t)
	carol.ID.Set(3)
	carol.UserName.Set("Carol")
	carol.NickName.Set(sql.NullString{String: "C", Valid: true})

	dave := bindSQLiteUsers(t)
	dave.ID.Set(4)
	dave.UserName.Set("Dave")
	dave.NickName.Set(sql.NullString{})

	sqlText, args := q.
		InsertInto(users).
		ValuesRowsFrom([]sqliteUser{carol, dave}).
		Render(dialect.BaseDialect{})

	_, err := db.Exec(sqlText, args...)
	require.NoError(t, err)

	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM users WHERE id IN (3, 4)`).Scan(&count))

	assert.Equal(t, 2, count)
}
