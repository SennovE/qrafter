package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

const (
	// DefaultMigrationVersionTable is used to store the current migration version.
	DefaultMigrationVersionTable = "qrafter_schema_version"
	baseMigrationVersion         = ""
	baseMigrationTarget          = "base"
	headMigrationVersion         = "head"
)

// MigrationDirection selects whether registered migrations are applied or reverted.
type MigrationDirection string

const (
	// DirectionUp applies migrations forward.
	DirectionUp MigrationDirection = "up"
	// DirectionDown reverts migrations backward.
	DirectionDown MigrationDirection = "down"
)

// MigrationApplyConfig configures registered migration execution.
type MigrationApplyConfig struct {
	DB migrationApplyDatabase

	DriverName     string
	DataSourceName string

	Dialect            dialect.Renderer
	Registry           []Migration
	Direction          MigrationDirection
	Target             string
	VersionTable       string
	DisableTransaction bool
}

type migrationApplyDatabase interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// MigrationApplyResult describes migration versions executed by ApplyMigrations.
type MigrationApplyResult struct {
	Direction MigrationDirection
	From      string
	To        string
	Applied   []string
}

type migrationTransactionBeginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// ApplyMigrations applies registered migrations in the requested direction up to target.
func ApplyMigrations(ctx context.Context, config *MigrationApplyConfig) (*MigrationApplyResult, error) {
	if err := validateMigrationApplyConfig(config); err != nil {
		return nil, err
	}

	db, closeDB, err := migrationApplyDatabaseConnection(ctx, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closeDB() }()

	registry, err := normalizeMigrationRegistry(config.Registry)
	if err != nil {
		return nil, err
	}

	versionTable := migrationVersionTable(config.VersionTable)
	if err := ensureMigrationVersionTable(ctx, db, config.Dialect, versionTable); err != nil {
		return nil, err
	}
	if err := ensureMigrationVersionRow(ctx, db, config.Dialect, versionTable); err != nil {
		return nil, err
	}

	current, err := readCurrentMigrationVersion(ctx, db, config.Dialect, versionTable)
	if err != nil {
		return nil, err
	}

	direction := normalizeMigrationDirection(config.Direction)
	target, err := resolveMigrationTarget(config.Target, registry, direction)
	if err != nil {
		return nil, err
	}

	plan, err := migrationPlan(registry, current, target, direction)
	if err != nil {
		return nil, err
	}

	result := &MigrationApplyResult{
		Direction: direction,
		From:      current,
		To:        target,
	}
	for _, planned := range plan {
		err := applyPlannedMigration(
			ctx,
			db,
			config.Dialect,
			versionTable,
			planned,
			direction,
			config.DisableTransaction,
		)
		if err != nil {
			return nil, err
		}
		result.Applied = append(result.Applied, planned.migration.Version)
		result.To = planned.currentAfter
	}
	return result, nil
}

func validateMigrationApplyConfig(config *MigrationApplyConfig) error {
	if config == nil {
		return fmt.Errorf("migrations: apply config is nil")
	}
	if config.Dialect == nil {
		return fmt.Errorf("migrations: dialect is required")
	}
	if config.DB != nil {
		return nil
	}
	if config.DriverName == "" {
		return fmt.Errorf("migrations: driver name is required")
	}
	if config.DataSourceName == "" {
		return fmt.Errorf("migrations: data source name is required")
	}
	return nil
}

func normalizeMigrationDirection(direction MigrationDirection) MigrationDirection {
	direction = MigrationDirection(strings.ToLower(strings.TrimSpace(string(direction))))
	if direction == "" {
		return DirectionUp
	}
	return direction
}

func migrationApplyDatabaseConnection(
	ctx context.Context,
	config *MigrationApplyConfig,
) (migrationApplyDatabase, func() error, error) {
	if config.DB != nil {
		return config.DB, noopClose, nil
	}

	db, err := sql.Open(config.DriverName, config.DataSourceName)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, nil, fmt.Errorf("ping database: %w; close database: %v", err, closeErr)
		}
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}
	return db, db.Close, nil
}

func normalizeMigrationRegistry(registry []Migration) ([]Migration, error) {
	normalized := append([]Migration(nil), registry...)
	sort.SliceStable(normalized, func(i, j int) bool {
		return normalized[i].Version < normalized[j].Version
	})

	seen := make(map[string]struct{}, len(normalized))
	for i := range normalized {
		version := strings.TrimSpace(normalized[i].Version)
		if version == "" {
			return nil, fmt.Errorf("migrations: registry contains migration without version")
		}
		if normalized[i].Up == nil {
			return nil, fmt.Errorf("migrations: migration %s has nil Up function", version)
		}
		if normalized[i].Down == nil {
			return nil, fmt.Errorf("migrations: migration %s has nil Down function", version)
		}
		if _, ok := seen[version]; ok {
			return nil, fmt.Errorf("migrations: duplicate migration version %s", version)
		}
		seen[version] = struct{}{}
		normalized[i].Version = version
	}
	return normalized, nil
}

func migrationVersionTable(table string) string {
	table = strings.TrimSpace(table)
	if table == "" {
		return DefaultMigrationVersionTable
	}
	return table
}

func ensureMigrationVersionTable(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	table string,
) error {
	stmt := ddl.CreateTable(table).
		IfNotExists().
		Columns(
			ddl.Column("id", ddl.Integer()).NotNull(),
			ddl.Column("version", ddl.VarChar(255)).NotNull().Default(""),
		).
		Constraints(ddl.PrimaryKey("id"))

	sqlText, err := ddl.Render(d, stmt)
	if err != nil {
		return fmt.Errorf("render migration version table: %w", err)
	}
	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("create migration version table: %w", err)
	}
	return nil
}

func ensureMigrationVersionRow(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	table string,
) error {
	query := "SELECT " + d.QuoteIdent("id") +
		" FROM " + d.QuoteIdent(table) +
		" WHERE " + d.QuoteIdent("id") + " = " + d.Placeholder(1)

	var id int
	if err := db.QueryRowContext(ctx, query, 1).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return insertCurrentMigrationVersion(ctx, db, d, table, baseMigrationVersion)
		}
		return fmt.Errorf("read migration version row: %w", err)
	}
	return nil
}

func readCurrentMigrationVersion(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	table string,
) (string, error) {
	query := "SELECT " + d.QuoteIdent("version") +
		" FROM " + d.QuoteIdent(table) +
		" WHERE " + d.QuoteIdent("id") + " = " + d.Placeholder(1)

	var version string
	if err := db.QueryRowContext(ctx, query, 1).Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			return baseMigrationVersion, nil
		}
		return "", fmt.Errorf("read current migration version: %w", err)
	}
	return version, nil
}

func writeCurrentMigrationVersion(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	table string,
	version string,
) error {
	update := "UPDATE " + d.QuoteIdent(table) +
		" SET " + d.QuoteIdent("version") + " = " + d.Placeholder(1) +
		" WHERE " + d.QuoteIdent("id") + " = " + d.Placeholder(2)

	result, err := db.ExecContext(ctx, update, version, 1)
	if err != nil {
		return fmt.Errorf("update current migration version: %w", err)
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return insertCurrentMigrationVersion(ctx, db, d, table, version)
	}
	return nil
}

func insertCurrentMigrationVersion(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	table string,
	version string,
) error {
	insert := "INSERT INTO " + d.QuoteIdent(table) +
		" (" + d.QuoteIdent("id") + ", " + d.QuoteIdent("version") + ")" +
		" VALUES (" + d.Placeholder(1) + ", " + d.Placeholder(2) + ")"
	if _, err := db.ExecContext(ctx, insert, 1, version); err != nil {
		return fmt.Errorf("insert current migration version: %w", err)
	}
	return nil
}

func resolveMigrationTarget(
	target string,
	registry []Migration,
	direction MigrationDirection,
) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		if direction == DirectionDown {
			return baseMigrationVersion, nil
		}
		return headTarget(registry), nil
	}
	if strings.EqualFold(target, headMigrationVersion) {
		return headTarget(registry), nil
	}
	if isBaseMigrationTarget(target) {
		return baseMigrationVersion, nil
	}
	if _, ok := migrationIndex(registry, target); !ok {
		return "", fmt.Errorf("migrations: target version %s is not registered", target)
	}
	return target, nil
}

func headTarget(registry []Migration) string {
	if len(registry) == 0 {
		return baseMigrationVersion
	}
	return registry[len(registry)-1].Version
}

func isBaseMigrationTarget(target string) bool {
	return target == "0" || strings.EqualFold(target, baseMigrationTarget)
}

type plannedMigration struct {
	migration    Migration
	currentAfter string
}

func migrationPlan(
	registry []Migration,
	current string,
	target string,
	direction MigrationDirection,
) ([]plannedMigration, error) {
	currentIndex, err := migrationVersionIndex(registry, current)
	if err != nil {
		return nil, err
	}
	targetIndex, err := migrationVersionIndex(registry, target)
	if err != nil {
		return nil, err
	}

	switch direction {
	case DirectionUp:
		if targetIndex < currentIndex {
			return nil, fmt.Errorf("migrations: target %s is before current version %s", displayVersion(target), displayVersion(current))
		}
		return upMigrationPlan(registry, currentIndex, targetIndex), nil
	case DirectionDown:
		if targetIndex > currentIndex {
			return nil, fmt.Errorf("migrations: target %s is after current version %s", displayVersion(target), displayVersion(current))
		}
		return downMigrationPlan(registry, currentIndex, targetIndex), nil
	default:
		return nil, fmt.Errorf("migrations: unsupported direction %q", direction)
	}
}

func migrationVersionIndex(registry []Migration, version string) (int, error) {
	if version == baseMigrationVersion {
		return -1, nil
	}
	index, ok := migrationIndex(registry, version)
	if !ok {
		return 0, fmt.Errorf("migrations: current version %s is not registered", version)
	}
	return index, nil
}

func migrationIndex(registry []Migration, version string) (int, bool) {
	for i := range registry {
		if registry[i].Version == version {
			return i, true
		}
	}
	return 0, false
}

func upMigrationPlan(registry []Migration, currentIndex, targetIndex int) []plannedMigration {
	plan := make([]plannedMigration, 0, targetIndex-currentIndex)
	for i := currentIndex + 1; i <= targetIndex; i++ {
		plan = append(plan, plannedMigration{
			migration:    registry[i],
			currentAfter: registry[i].Version,
		})
	}
	return plan
}

func downMigrationPlan(registry []Migration, currentIndex, targetIndex int) []plannedMigration {
	plan := make([]plannedMigration, 0, currentIndex-targetIndex)
	for i := currentIndex; i > targetIndex; i-- {
		currentAfter := baseMigrationVersion
		if i > 0 {
			currentAfter = registry[i-1].Version
		}
		plan = append(plan, plannedMigration{
			migration:    registry[i],
			currentAfter: currentAfter,
		})
	}
	return plan
}

func applyPlannedMigration(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	versionTable string,
	planned plannedMigration,
	direction MigrationDirection,
	disableTransaction bool,
) error {
	if !disableTransaction {
		if beginner, ok := db.(migrationTransactionBeginner); ok {
			return applyPlannedMigrationInTx(ctx, beginner, d, versionTable, planned, direction)
		}
	}
	return applyPlannedMigrationOnDB(ctx, db, d, versionTable, planned, direction)
}

func applyPlannedMigrationInTx(
	ctx context.Context,
	beginner migrationTransactionBeginner,
	d dialect.Renderer,
	versionTable string,
	planned plannedMigration,
	direction MigrationDirection,
) error {
	tx, err := beginner.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s transaction: %w", planned.migration.Version, err)
	}

	if err := applyPlannedMigrationOnDB(ctx, tx, d, versionTable, planned, direction); err != nil {
		return rollbackMigrationTx(tx, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s transaction: %w", planned.migration.Version, err)
	}
	return nil
}

func applyPlannedMigrationOnDB(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	versionTable string,
	planned plannedMigration,
	direction MigrationDirection,
) error {
	if err := executeRegisteredMigration(ctx, db, d, planned.migration, direction); err != nil {
		return err
	}
	if err := writeCurrentMigrationVersion(ctx, db, d, versionTable, planned.currentAfter); err != nil {
		return err
	}
	return nil
}

func rollbackMigrationTx(tx *sql.Tx, cause error) error {
	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return fmt.Errorf("%w; rollback migration transaction: %v", cause, err)
	}
	return cause
}

func executeRegisteredMigration(
	ctx context.Context,
	db migrationApplyDatabase,
	d dialect.Renderer,
	migration Migration,
	direction MigrationDirection,
) error {
	var statements ddl.Statements
	switch direction {
	case DirectionUp:
		statements = migration.Up()
	case DirectionDown:
		statements = migration.Down()
	default:
		return fmt.Errorf("migrations: unsupported direction %q", direction)
	}

	for i := range statements {
		sqlText, err := ddl.Render(d, statements[i])
		if err != nil {
			return fmt.Errorf("render %s migration %s: %w", direction, migration.Version, err)
		}
		if strings.TrimSpace(sqlText) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("execute %s migration %s: %w", direction, migration.Version, err)
		}
	}
	return nil
}

func displayVersion(version string) string {
	if version == baseMigrationVersion {
		return baseMigrationTarget
	}
	return version
}
