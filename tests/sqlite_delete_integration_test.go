package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSQLiteDelete(t *testing.T) {
	db := openSQLiteIntegrationDB(t)
	users := bindSQLiteUsers(t)

	sqlText, args := q.
		Delete(users).
		Where(users.ID.Eq(1)).
		Render(dialect.BaseDialect{})

	result, err := db.Exec(sqlText, args...)
	require.NoError(t, err)

	affected, err := result.RowsAffected()
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	var count int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, 1).Scan(&count))
	assert.Equal(t, 0, count)
}
