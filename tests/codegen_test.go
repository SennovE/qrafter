package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SennovE/qrafter/ddl"
	qmig "github.com/SennovE/qrafter/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type codegenDB struct{}

func (codegenDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	panic("codegenDB.QueryContext must not be called")
}

func (codegenDB) QueryRowContext(context.Context, string, ...any) *sql.Row {
	panic("codegenDB.QueryRowContext must not be called")
}

type codegenIntrospector struct {
	schema qmig.Schema
}

func (i codegenIntrospector) ReadSchema(context.Context, qmig.Database) (qmig.Schema, error) {
	return i.schema, nil
}

func TestMakeMigrationWritesCompleteGeneratedFile(t *testing.T) {
	outDir := t.TempDir()
	path, err := qmig.MakeMigration(
		context.Background(),
		"test generation",
		outDir,
		&qmig.MigrationToolConfig{
			DB:           codegenDB{},
			Introspector: codegenIntrospector{},
			Desired: qmig.Schema{Tables: []qmig.Table{{
				Name: "users",
				Columns: []qmig.Column{
					{Name: "id", Type: ddl.UUID(), NotNull: true, HasDefault: true, DefaultExpr: "uuid_generate_v4()"},
					{Name: "email", Type: ddl.Text(), NotNull: true},
				},
				Constraints: []qmig.Constraint{
					{Name: "users_pkey", Kind: qmig.ConstraintPrimaryKey, Columns: []string{"id"}},
				},
				Indexes: []qmig.Index{{
					Name:      "idx_users_email",
					TableName: "users",
					Keys:      []qmig.IndexKey{{Expression: `"email"`}},
				}},
			}}},
		},
	)
	require.NoError(t, err)
	requirePathInsideDir(t, path, outDir)

	// #nosec G304 -- MakeMigration returned this path inside t.TempDir(), checked above.
	gotBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	revision := migrationRevision(t, path)
	want := fmt.Sprintf(`package migrations

import qddl "github.com/SennovE/qrafter/ddl"

func Up%[1]s() qddl.Statements {
	return qddl.Statements{
		qddl.CreateTable("users").
			Columns(
				qddl.Column("id", qddl.UUID()).NotNull().DefaultExpr("uuid_generate_v4()"),
				qddl.Column("email", qddl.Text()).NotNull(),
			).
			Constraints(
				qddl.PrimaryKey("id").Named("users_pkey"),
			),
		qddl.CreateIndex("idx_users_email").On(
			"users",
			qddl.KeyCol("email"),
		),
	}
}

func Down%[1]s() qddl.Statements {
	return qddl.Statements{
		qddl.DropIndex("idx_users_email"),
		qddl.DropTable("users"),
	}
}
`, revision)

	assert.Equal(t, want, string(gotBytes))
}

func migrationRevision(t *testing.T, path string) string {
	t.Helper()

	revision, _, ok := strings.Cut(filepath.Base(path), "_")
	require.True(t, ok, "migration filename should contain revision and comment: %s", path)
	return revision
}

func requirePathInsideDir(t *testing.T, path, dir string) {
	t.Helper()

	absPath, err := filepath.Abs(path)
	require.NoError(t, err)
	absDir, err := filepath.Abs(dir)
	require.NoError(t, err)

	rel, err := filepath.Rel(absDir, absPath)
	require.NoError(t, err)
	require.NotEqual(t, ".", rel)
	require.False(t, strings.HasPrefix(rel, ".."+string(filepath.Separator)))
	require.False(t, filepath.IsAbs(rel))
}
