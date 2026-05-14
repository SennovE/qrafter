package tests

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sqlxUser struct {
	ID       q.Column[int]            `db:"id"`
	UserName q.Column[string]         `db:"user_name"`
	NickName q.Column[sql.NullString] `db:"nick_name"`
}

func (sqlxUser) TableConfig() q.TableConfig {
	return q.TableConfig{Name: "users"}
}

func TestSQLXStructScanScansIntoColumns(t *testing.T) {
	db := openStaticSQLXDB(t, []string{"id", "user_name", "nick_name"}, [][]driver.Value{
		{int64(42), []byte("Alice"), nil},
	})

	rows, err := db.Queryx("SELECT id, user_name, nick_name FROM users")
	require.NoError(t, err)
	defer rows.Close()

	require.True(t, rows.Next())

	got := q.MustNewTable[sqlxUser]()
	require.NoError(t, rows.StructScan(&got))

	assert.Equal(t, 42, got.ID.Get())
	assert.Equal(t, "Alice", got.UserName.Get())
	assert.Equal(t, sql.NullString{}, got.NickName.Get())
	require.NoError(t, rows.Err())
}

func TestSQLXSelectScansIntoColumnSlice(t *testing.T) {
	db := openStaticSQLXDB(t, []string{"id", "user_name", "nick_name"}, [][]driver.Value{
		{int64(1), "Alice", "Al"},
		{int64(2), "Bob", nil},
	})

	var got []sqlxUser
	require.NoError(t, db.Select(&got, "SELECT id, user_name, nick_name FROM users"))
	require.Len(t, got, 2)

	assert.Equal(t, 1, got[0].ID.Get())
	assert.Equal(t, "Alice", got[0].UserName.Get())
	assert.Equal(t, sql.NullString{String: "Al", Valid: true}, got[0].NickName.Get())

	assert.Equal(t, 2, got[1].ID.Get())
	assert.Equal(t, "Bob", got[1].UserName.Get())
	assert.Equal(t, sql.NullString{}, got[1].NickName.Get())
}

func openStaticSQLXDB(t *testing.T, columns []string, rows [][]driver.Value) *sqlx.DB {
	t.Helper()

	db := sql.OpenDB(staticConnector{
		columns: columns,
		rows:    rows,
	})
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return sqlx.NewDb(db, "qrafter-static")
}

type staticConnector struct {
	columns []string
	rows    [][]driver.Value
}

func (c staticConnector) Connect(context.Context) (driver.Conn, error) {
	return &staticConn{columns: c.columns, rows: c.rows}, nil
}

func (c staticConnector) Driver() driver.Driver {
	return staticDriver{}
}

type staticDriver struct{}

func (staticDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("use sql.OpenDB with staticConnector")
}

type staticConn struct {
	columns []string
	rows    [][]driver.Value
}

func (c *staticConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("static test driver does not support prepared statements")
}

func (c *staticConn) Close() error {
	return nil
}

func (c *staticConn) Begin() (driver.Tx, error) {
	return nil, errors.New("static test driver does not support transactions")
}

func (c *staticConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return &staticRows{columns: c.columns, rows: c.rows}, nil
}

type staticRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *staticRows) Columns() []string {
	return r.columns
}

func (r *staticRows) Close() error {
	return nil
}

func (r *staticRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}

	copy(dest, r.rows[r.index])
	r.index++
	return nil
}
