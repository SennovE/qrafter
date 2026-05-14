package tests

import (
	"database/sql"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type sqliteUser struct {
	ID       q.Column[int]            `db:"id"`
	UserName q.Column[string]         `db:"user_name"`
	NickName q.Column[sql.NullString] `db:"nick_name"`
}

func (sqliteUser) TableConfig() q.TableConfig {
	return q.TableConfig{Name: "users"}
}

func TestIntegrationSQLiteDirectColumnScan(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	sqlText, args := q.
		Select(users.ID, users.UserName, users.NickName).
		Where(users.ID.Eq(1)).
		Render(dialect.BaseDialect{})

	require.NoError(t, db.QueryRow(sqlText, args...).Scan(&users.ID, &users.UserName, &users.NickName))

	assert.Equal(t, 1, users.ID.Get())
	assert.Equal(t, "Alice", users.UserName.Get())
	assert.Equal(t, sql.NullString{String: "Al", Valid: true}, users.NickName.Get())
}

func TestIntegrationSQLiteScanDest(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	sqlText, args := q.
		Select(users.ID, users.UserName, users.NickName).
		Where(users.ID.Eq(2)).
		Render(dialect.BaseDialect{})

	dest, err := q.ScanDest(&users)
	require.NoError(t, err)

	require.NoError(t, db.QueryRow(sqlText, args...).Scan(dest...))

	assert.Equal(t, 2, users.ID.Get())
	assert.Equal(t, "Bob", users.UserName.Get())
	assert.Equal(t, sql.NullString{}, users.NickName.Get())
}

func bindSQLiteUsers(t *testing.T) sqliteUser {
	t.Helper()

	users := sqliteUser{}
	require.NoError(t, q.Bind(&users))

	return users
}

func openSQLiteIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			user_name TEXT NOT NULL,
			nick_name TEXT NULL
		)
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (id, user_name, nick_name)
		VALUES (1, 'Alice', 'Al'), (2, 'Bob', NULL)
	`)
	require.NoError(t, err)

	return db
}
