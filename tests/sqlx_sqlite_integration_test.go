package tests

import (
	"database/sql"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSQLiteSQLXStructScan(t *testing.T) {
	rawDB := openSQLiteIntegrationDB(t)
	db := sqlx.NewDb(rawDB, "sqlite")
	users := bindSQLiteUsers(t)

	sqlText, args := q.
		Select(q.Star()).
		Where(users.ID.Eq(1)).
		Render(dialect.BaseDialect{})

	rows, err := db.Queryx(sqlText, args...)
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())

	require.NoError(t, rows.StructScan(&users))

	assert.Equal(t, 1, users.ID.Get())
	assert.Equal(t, "Alice", users.UserName.Get())
	assert.Equal(t, sql.NullString{String: "Al", Valid: true}, users.NickName.Get())
	require.NoError(t, rows.Err())
}
