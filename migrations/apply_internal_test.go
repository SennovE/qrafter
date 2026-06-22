package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

func TestValidateMigrationApplyConfigErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *MigrationApplyConfig
		want   string
	}{
		{name: "nil", config: nil, want: "apply config is nil"},
		{name: "dialect", config: &MigrationApplyConfig{}, want: "dialect is required"},
		{name: "driver", config: &MigrationApplyConfig{Dialect: dialect.PostgreSQL{}}, want: "driver name is required"},
		{
			name:   "dsn",
			config: &MigrationApplyConfig{Dialect: dialect.PostgreSQL{}, DriverName: "postgres"},
			want:   "data source name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMigrationApplyConfig(tt.config)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want to contain %q", err, tt.want)
			}
		})
	}
}

func TestNormalizeMigrationRegistry(t *testing.T) {
	up := func() ddl.Statements { return nil }
	down := func() ddl.Statements { return nil }
	registry, err := normalizeMigrationRegistry([]Migration{
		New("002", up, down),
		New("001", up, down),
	})
	if err != nil {
		t.Fatalf("normalize registry: %v", err)
	}
	if registry[0].Version != "001" || registry[1].Version != "002" {
		t.Fatalf("registry order = %#v", registry)
	}
	registry, err = normalizeMigrationRegistry([]Migration{New(" 003 ", up, down)})
	if err != nil {
		t.Fatalf("normalize trimmed registry: %v", err)
	}
	if registry[0].Version != "003" {
		t.Fatalf("trimmed version = %q", registry[0].Version)
	}

	errorCases := []Migration{
		New(" ", up, down),
		{Version: "001", Up: nil, Down: down},
		{Version: "001", Up: up, Down: nil},
	}
	for _, migration := range errorCases {
		if _, err := normalizeMigrationRegistry([]Migration{migration}); err == nil {
			t.Fatalf("normalize registry %#v error = nil", migration)
		}
	}
	if _, err := normalizeMigrationRegistry([]Migration{New("001", up, down), New("001", up, down)}); err == nil {
		t.Fatal("duplicate version error = nil")
	}
}

func TestMigrationTargetsAndPlans(t *testing.T) {
	registry := []Migration{
		New("001", func() ddl.Statements { return nil }, func() ddl.Statements { return nil }),
		New("002", func() ddl.Statements { return nil }, func() ddl.Statements { return nil }),
		New("003", func() ddl.Statements { return nil }, func() ddl.Statements { return nil }),
	}

	target, err := resolveMigrationTarget("", registry, DirectionUp)
	if err != nil || target != "003" {
		t.Fatalf("default up target = %q, %v", target, err)
	}
	target, err = resolveMigrationTarget("base", registry, DirectionDown)
	if err != nil || target != "" {
		t.Fatalf("base target = %q, %v", target, err)
	}
	if _, err := resolveMigrationTarget("999", registry, DirectionUp); err == nil {
		t.Fatal("missing target error = nil")
	}

	upPlan, err := migrationPlan(registry, "", "002", DirectionUp)
	if err != nil {
		t.Fatalf("up plan: %v", err)
	}
	if got := []string{upPlan[0].migration.Version, upPlan[1].migration.Version}; fmt.Sprint(got) != "[001 002]" {
		t.Fatalf("up plan = %#v", upPlan)
	}

	downPlan, err := migrationPlan(registry, "003", "001", DirectionDown)
	if err != nil {
		t.Fatalf("down plan: %v", err)
	}
	if downPlan[0].migration.Version != "003" || downPlan[0].currentAfter != "002" ||
		downPlan[1].migration.Version != "002" || downPlan[1].currentAfter != "001" {
		t.Fatalf("down plan = %#v", downPlan)
	}

	for _, tt := range []struct {
		current   string
		target    string
		direction MigrationDirection
	}{
		{current: "003", target: "001", direction: DirectionUp},
		{current: "001", target: "003", direction: DirectionDown},
		{current: "001", target: "003", direction: "sideways"},
		{current: "999", target: "003", direction: DirectionUp},
	} {
		if _, err := migrationPlan(registry, tt.current, tt.target, tt.direction); err == nil {
			t.Fatalf("migrationPlan(%#v) error = nil", tt)
		}
	}
}

func TestExecuteRegisteredMigration(t *testing.T) {
	db := &fakeMigrationDB{}
	migration := New(
		"001",
		func() ddl.Statements {
			return ddl.Statements{ddl.RawSQL("  "), ddl.RawSQL("CREATE TABLE users (id integer)")}
		},
		func() ddl.Statements {
			return ddl.Statements{ddl.RawSQL("DROP TABLE users")}
		},
	)

	if err := executeRegisteredMigration(context.Background(), db, dialect.PostgreSQL{}, migration, DirectionUp); err != nil {
		t.Fatalf("execute up: %v", err)
	}
	if len(db.execs) != 1 || db.execs[0] != "CREATE TABLE users (id integer)" {
		t.Fatalf("execs after up = %#v", db.execs)
	}
	if err := executeRegisteredMigration(context.Background(), db, dialect.PostgreSQL{}, migration, DirectionDown); err != nil {
		t.Fatalf("execute down: %v", err)
	}
	if err := executeRegisteredMigration(context.Background(), db, dialect.PostgreSQL{}, migration, "sideways"); err == nil {
		t.Fatal("unsupported direction error = nil")
	}

	badRender := New("002", func() ddl.Statements { return ddl.Statements{ddl.CreateTable("empty")} }, func() ddl.Statements { return nil })
	if err := executeRegisteredMigration(context.Background(), db, dialect.PostgreSQL{}, badRender, DirectionUp); err == nil {
		t.Fatal("render error = nil")
	}

	db.failExec = fmt.Errorf("boom")
	if err := executeRegisteredMigration(context.Background(), db, dialect.PostgreSQL{}, migration, DirectionDown); err == nil {
		t.Fatal("exec error = nil")
	}
}

func TestWriteCurrentMigrationVersionInsertsWhenUpdateAffectsNoRows(t *testing.T) {
	db := &fakeMigrationDB{rowsAffected: 0}
	if err := writeCurrentMigrationVersion(context.Background(), db, dialect.PostgreSQL{}, "versions", "001"); err != nil {
		t.Fatalf("write version: %v", err)
	}
	if len(db.execs) != 2 {
		t.Fatalf("execs = %#v, want update and insert", db.execs)
	}
}

func TestApplyMigrationsWithMockDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT "id"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec(`INSERT INTO "versions"`).
		WithArgs(1, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`SELECT "version"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(""))
	mock.ExpectExec("CREATE TABLE users").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE "versions"`).
		WithArgs("001", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	result, err := ApplyMigrations(context.Background(), &MigrationApplyConfig{
		DB:                 db,
		Dialect:            dialect.PostgreSQL{},
		VersionTable:       "versions",
		DisableTransaction: true,
		Registry: []Migration{New(
			"001",
			func() ddl.Statements { return ddl.Statements{ddl.RawSQL("CREATE TABLE users (id integer)")} },
			func() ddl.Statements { return nil },
		)},
	})
	if err != nil {
		t.Fatalf("ApplyMigrations error = %v", err)
	}
	if result.From != "" || result.To != "001" || fmt.Sprint(result.Applied) != "[001]" {
		t.Fatalf("result = %#v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyMigrationsRollsBackTransactionOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	mock.ExpectExec("CREATE TABLE").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`SELECT "id"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(`SELECT "version"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(""))
	mock.ExpectBegin()
	mock.ExpectExec("BROKEN SQL").
		WillReturnError(fmt.Errorf("boom"))
	mock.ExpectRollback()

	_, err = ApplyMigrations(context.Background(), &MigrationApplyConfig{
		DB:           db,
		Dialect:      dialect.PostgreSQL{},
		VersionTable: "versions",
		Registry: []Migration{New(
			"001",
			func() ddl.Statements { return ddl.Statements{ddl.RawSQL("BROKEN SQL")} },
			func() ddl.Statements { return nil },
		)},
	})
	if err == nil || !strings.Contains(err.Error(), "execute up migration 001") {
		t.Fatalf("ApplyMigrations error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestMigrationApplyDatabaseConnectionOpenError(t *testing.T) {
	_, _, err := migrationApplyDatabaseConnection(context.Background(), &MigrationApplyConfig{
		DriverName:     "missing-driver",
		DataSourceName: "ignored",
	})
	if err == nil || !strings.Contains(err.Error(), "open database") {
		t.Fatalf("open error = %v", err)
	}

	if got := headTarget(nil); got != "" {
		t.Fatalf("empty head target = %q", got)
	}
	if got := displayVersion(""); got != "base" {
		t.Fatalf("display base = %q", got)
	}
}

type fakeMigrationDB struct {
	execs        []string
	failExec     error
	rowsAffected int64
}

func (f *fakeMigrationDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, fmt.Errorf("unexpected query")
}

func (f *fakeMigrationDB) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

func (f *fakeMigrationDB) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	if f.failExec != nil {
		return nil, f.failExec
	}
	f.execs = append(f.execs, query)
	return fakeSQLResult{rowsAffected: f.rowsAffected}, nil
}

type fakeSQLResult struct {
	rowsAffected int64
}

func (f fakeSQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (f fakeSQLResult) RowsAffected() (int64, error) {
	return f.rowsAffected, nil
}
