package tests

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnScan(t *testing.T) {
	var id q.Column[int]

	require.NoError(t, id.Scan(int64(42)))

	assert.Equal(t, 42, id.Get())
}

func TestColumnScan_NullableTypes(t *testing.T) {
	var name q.Column[sql.NullString]

	require.NoError(t, name.Scan("Alice"))
	assert.Equal(t, sql.NullString{String: "Alice", Valid: true}, name.Get())

	require.NoError(t, name.Scan(nil))
	assert.Equal(t, sql.NullString{}, name.Get())
}

func TestColumnScan_PointerTypes(t *testing.T) {
	var name q.Column[*string]

	require.NoError(t, name.Scan([]byte("Alice")))
	require.NotNil(t, name.Get())
	assert.Equal(t, "Alice", *name.Get())

	require.NoError(t, name.Scan(nil))
	assert.Nil(t, name.Get())
}

func TestColumnScan_NullIntoValueTypeReturnsError(t *testing.T) {
	var name q.Column[string]

	assert.Error(t, name.Scan(nil))
}

func TestColumnValue(t *testing.T) {
	var id q.Column[int]
	id.Set(42)

	value, err := id.Value()

	require.NoError(t, err)
	assert.Equal(t, int64(42), value)
}

func TestColumnValue_NullableTypes(t *testing.T) {
	var name q.Column[sql.NullString]
	name.Set(sql.NullString{String: "Alice", Valid: true})

	value, err := name.Value()

	require.NoError(t, err)
	assert.Equal(t, driver.Value("Alice"), value)
}

func TestScanDest(t *testing.T) {
	type row struct {
		ID      q.Column[int]
		Name    q.Column[string]
		Ignored string
		hidden  q.Column[int]
	}

	var user row
	dest, err := q.ScanDest(&user)
	require.NoError(t, err)
	require.Len(t, dest, 2)

	require.NoError(t, dest[0].(sql.Scanner).Scan(int64(42)))
	require.NoError(t, dest[1].(sql.Scanner).Scan("Alice"))

	assert.Equal(t, 42, user.ID.Get())
	assert.Equal(t, "Alice", user.Name.Get())
}

func TestScanDest_RequiresPointerToStruct(t *testing.T) {
	dest, err := q.ScanDest(struct{}{})

	assert.Nil(t, dest)
	assert.Error(t, err)
}
