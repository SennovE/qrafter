package tests

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pointerOnlyValuer string

func (v *pointerOnlyValuer) Value() (driver.Value, error) {
	return "ptr:" + string(*v), nil
}

type valueValuer string

func (v valueValuer) Value() (driver.Value, error) {
	return "value:" + string(v), nil
}

func TestColumnScanPrimitiveConversions(t *testing.T) {
	var flag q.Column[bool]
	require.NoError(t, flag.Scan([]byte("true")))
	assert.True(t, flag.Get())

	var signed q.Column[int8]
	require.NoError(t, signed.Scan("12"))
	assert.Equal(t, int8(12), signed.Get())

	var unsigned q.Column[uint16]
	require.NoError(t, unsigned.Scan(int64(42)))
	assert.Equal(t, uint16(42), unsigned.Get())

	var float q.Column[float32]
	require.NoError(t, float.Scan([]byte("3.5")))
	assert.Equal(t, float32(3.5), float.Get())

	var text q.Column[string]
	require.NoError(t, text.Scan([]byte("hello")))
	assert.Equal(t, "hello", text.Get())
}

func TestColumnScanBytesAreCloned(t *testing.T) {
	src := []byte("abc")
	var col q.Column[[]byte]

	require.NoError(t, col.Scan(src))
	src[0] = 'z'

	assert.Equal(t, []byte("abc"), col.Get())
}

func TestColumnScanUnsupportedConversions(t *testing.T) {
	var ints q.Column[[]int]
	assert.Error(t, ints.Scan([]byte("123")))

	var flag q.Column[bool]
	assert.Error(t, flag.Scan("not-bool"))

	users := q.MustNewTable[Person]()
	err := users.Age.Scan("not-an-int")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `scan column "age"`)
}

func TestColumnScanNullAndValueEdges(t *testing.T) {
	var anyValue q.Column[any]
	require.NoError(t, anyValue.Scan(nil))
	assert.Nil(t, anyValue.Get())

	var values q.Column[map[string]string]
	require.NoError(t, values.Scan(nil))
	assert.Nil(t, values.Get())

	var flag q.Column[bool]
	err := flag.Scan(struct{ X int }{X: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "converting driver.Value")

	var valuer q.Column[pointerOnlyValuer]
	valuer.Set("ok")
	got, err := valuer.Value()
	require.NoError(t, err)
	assert.Equal(t, driver.Value("ptr:ok"), got)

	var direct q.Column[valueValuer]
	direct.Set("ok")
	got, err = direct.Value()
	require.NoError(t, err)
	assert.Equal(t, driver.Value("value:ok"), got)
}

func TestColumnScanConversionErrorBranches(t *testing.T) {
	var text q.Column[string]
	assert.Error(t, text.Scan(123))

	var unsigned q.Column[uint]
	assert.Error(t, unsigned.Scan("-1"))

	var float q.Column[float64]
	assert.Error(t, float.Scan("not-float"))

	var bytes q.Column[[]byte]
	assert.Error(t, bytes.Scan(123))
}

func TestScanDestIgnoresNonScannerFieldsAndNilInput(t *testing.T) {
	type row struct {
		ID      q.Column[int]
		Name    sql.NullString
		Ignored string
	}

	var r row
	dest, err := q.ScanDest(&r)
	require.NoError(t, err)
	assert.Len(t, dest, 2)

	_, err = q.ScanDest((*row)(nil))
	assert.Error(t, err)
}

func TestPredicateOptions(t *testing.T) {
	users := q.MustNewTable[User]()
	name := "Alice"

	predicates := q.Predicates(
		q.When(true, users.UserName.Eq("Bob")),
		q.When(false, users.UserName.Eq("Ignored")),
		q.WhenFunc(true, func() q.Predicater { return users.Age.Ge("18") }),
		q.WhenFunc(false, func() q.Predicater { return users.Age.Ge("99") }),
		q.WhenPtr(&name, func(v string) q.Predicater { return users.UserName.NotLike(v + "%") }),
		q.WhenPtr[string](nil, func(v string) q.Predicater { return users.UserName.Eq(v) }),
		q.WhenNotEmpty("admin", func(v string) q.Predicater { return users.UserName.Like(v + "%") }),
		q.WhenNotEmpty("", func(v string) q.Predicater { return users.UserName.Eq(v) }),
	)

	sql, args := q.Select(users.UserName).
		Where(predicates...).
		MustRender(dialect.PostgreSQL{})

	assert.Equal(t, `SELECT "table"."user_name"
FROM "table"
WHERE "table"."user_name" = $1 AND "table"."userAge" >= $2 AND "table"."user_name" NOT LIKE $3 AND "table"."user_name" LIKE $4`, sql)
	assert.Equal(t, []any{"Bob", "18", "Alice%", "admin%"}, args)
}
