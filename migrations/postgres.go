package migrations

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

//go:embed inspect/postgresql/version.sql
var getPostgeSQLVersion string

//go:embed inspect/postgresql/tables.sql
var getPostgeSQLTables string

//go:embed inspect/postgresql/columns.sql
var getPostgeSQLColumns string

//go:embed inspect/postgresql/constraints.sql
var getPostgeSQLConstraints string

//go:embed inspect/postgresql/index_metadata.sql
var getPostgeSQLIndexMetadata string

//go:embed inspect/postgresql/index_keys.sql
var getPostgeSQLIndexKeys string

// PostgreSQL reads schema metadata from PostgreSQL.
type PostgreSQL struct {
	options postgreSQLOptions
}

type postgreSQLOptions struct {
	schemas    []string
	allSchemas bool
}

const defaultPostgreSQLSchema = "public"

// PostgreSQLOption configures PostgreSQL schema introspection.
type PostgreSQLOption func(*postgreSQLOptions)

// WithSchemas limits introspection to the given schemas. The default is public.
func WithSchemas(schemas ...string) PostgreSQLOption {
	return func(options *postgreSQLOptions) {
		options.schemas = normalizeSchemaNames(schemas)
		options.allSchemas = false
	}
}

// WithAllSchemas introspects all non-system PostgreSQL schemas.
func WithAllSchemas() PostgreSQLOption {
	return func(options *postgreSQLOptions) {
		options.schemas = nil
		options.allSchemas = true
	}
}

// NewPostgreSQL creates a PostgreSQL schema introspector.
func NewPostgreSQL(opts ...PostgreSQLOption) PostgreSQL {
	options := defaultPostgreSQLOptions()
	for _, opt := range opts {
		opt(&options)
	}
	return PostgreSQL{options: options}
}

// ReadSchema reads a PostgreSQL database schema.
func (p PostgreSQL) ReadSchema(ctx context.Context, db Database) (Schema, error) {
	options := p.effectiveOptions()

	version, err := readPostgreSQLVersion(ctx, db)
	if err != nil {
		return Schema{}, err
	}

	tables, err := readPostgreSQLTables(ctx, db, options)
	if err != nil {
		return Schema{}, err
	}
	if err := readPostgreSQLColumns(ctx, db, options, tables, version); err != nil {
		return Schema{}, err
	}
	if err := readPostgreSQLConstraints(ctx, db, options, tables); err != nil {
		return Schema{}, err
	}
	if err := readPostgreSQLIndexes(ctx, db, options, tables, version); err != nil {
		return Schema{}, err
	}

	schema := schemaFromTableMap(tables)
	schema.normalize()
	return schema, nil
}

// ReadDDL reads a PostgreSQL schema and converts it into ddl statements.
func (p PostgreSQL) ReadDDL(ctx context.Context, db Database) (ddl.Statements, error) {
	return ReadDDL(ctx, db, p)
}

func defaultPostgreSQLOptions() postgreSQLOptions {
	return postgreSQLOptions{schemas: []string{defaultPostgreSQLSchema}}
}

func (p PostgreSQL) effectiveOptions() postgreSQLOptions {
	options := p.options
	if !options.allSchemas && len(options.schemas) == 0 {
		options.schemas = []string{defaultPostgreSQLSchema}
	}
	return options
}

func readPostgreSQLVersion(ctx context.Context, db Database) (int, error) {
	var version int
	if err := db.QueryRowContext(ctx, getPostgeSQLVersion).Scan(&version); err != nil {
		return 0, fmt.Errorf("read PostgreSQL server version: %w", err)
	}
	return version, nil
}

func readPostgreSQLTables(ctx context.Context, db Database, options postgreSQLOptions) (map[tableKey]*Table, error) {
	predicate, args := postgreSQLSchemaPredicate("n", options)
	query := strings.Replace(
		getPostgeSQLTables,
		"{{PREDICATE}}",
		predicate,
		1,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("read PostgreSQL tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tables := make(map[tableKey]*Table)
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.Schema, &table.Name); err != nil {
			return nil, fmt.Errorf("scan PostgreSQL table: %w", err)
		}
		key := tableKey{schema: table.Schema, table: table.Name}
		tables[key] = &table
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PostgreSQL tables: %w", err)
	}
	return tables, nil
}

func readPostgreSQLColumns(
	ctx context.Context,
	db Database,
	options postgreSQLOptions,
	tables map[tableKey]*Table,
	version int,
) error {
	generatedColumn := "''"
	if version >= 120000 {
		generatedColumn = "a.attgenerated::text"
	}
	predicate, args := postgreSQLSchemaPredicate("n", options)
	query := strings.Replace(
		getPostgeSQLColumns,
		"{{GENERATED}}",
		generatedColumn,
		1,
	)
	query = strings.Replace(
		query,
		"{{PREDICATE}}",
		predicate,
		1,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("read PostgreSQL columns: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var column Column
		var defaultExpr sql.NullString
		var typeName string
		var identity string
		var generated string
		if err := rows.Scan(
			&column.Schema,
			&column.TableName,
			&column.Position,
			&column.Name,
			&typeName,
			&column.NotNull,
			&defaultExpr,
			&identity,
			&generated,
		); err != nil {
			return fmt.Errorf("scan PostgreSQL column: %w", err)
		}

		column.DatabaseType = typeName
		column.Type = postgreSQLType(typeName)
		column.Identity = postgresIdentityKind(identity)
		column.Generated = postgresGeneratedKind(generated)
		if defaultExpr.Valid {
			if column.Generated != GeneratedNone {
				column.GeneratedExpr = defaultExpr.String
			} else {
				column.HasDefault = true
				column.DefaultExpr = defaultExpr.String
			}
		}

		key := tableKey{schema: column.Schema, table: column.TableName}
		if table, ok := tables[key]; ok {
			table.Columns = append(table.Columns, column)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate PostgreSQL columns: %w", err)
	}
	return nil
}

func readPostgreSQLConstraints(
	ctx context.Context,
	db Database,
	options postgreSQLOptions,
	tables map[tableKey]*Table,
) error {
	predicate, args := postgreSQLSchemaPredicate("n", options)
	query := strings.Replace(
		getPostgeSQLConstraints,
		"{{PREDICATE}}",
		predicate,
		1,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("read PostgreSQL constraints: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var constraint Constraint
		var kind string
		var columns string
		var referenceColumns string
		var onDelete string
		var onUpdate string
		if err := rows.Scan(
			&constraint.Schema,
			&constraint.TableName,
			&constraint.Name,
			&kind,
			&columns,
			&constraint.Reference.Schema,
			&constraint.Reference.TableName,
			&referenceColumns,
			&onDelete,
			&onUpdate,
			&constraint.CheckExpr,
		); err != nil {
			return fmt.Errorf("scan PostgreSQL constraint: %w", err)
		}

		constraint.Kind = postgresConstraintKind(kind)
		constraint.Columns = splitPostgresList(columns)
		constraint.Reference.Columns = splitPostgresList(referenceColumns)
		constraint.OnDelete = postgresReferenceAction(onDelete)
		constraint.OnUpdate = postgresReferenceAction(onUpdate)

		key := tableKey{schema: constraint.Schema, table: constraint.TableName}
		if table, ok := tables[key]; ok {
			table.Constraints = append(table.Constraints, constraint)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate PostgreSQL constraints: %w", err)
	}
	return nil
}

func readPostgreSQLIndexes(
	ctx context.Context,
	db Database,
	options postgreSQLOptions,
	tables map[tableKey]*Table,
	version int,
) error {
	indexes, err := readPostgreSQLIndexMetadata(ctx, db, options, tables, version)
	if err != nil {
		return err
	}
	if err := readPostgreSQLIndexKeys(ctx, db, options, indexes, version); err != nil {
		return err
	}
	appendPostgreSQLIndexes(tables, indexes)
	return nil
}

func readPostgreSQLIndexMetadata(
	ctx context.Context,
	db Database,
	options postgreSQLOptions,
	tables map[tableKey]*Table,
	version int,
) (map[indexKey]*Index, error) {
	nullsNotDistinct := "false"
	if version >= 150000 {
		nullsNotDistinct = "i.indnullsnotdistinct"
	}
	predicate, args := postgreSQLSchemaPredicate("n", options)

	query := strings.Replace(
		getPostgeSQLIndexMetadata,
		"NULLS_NOT_DISTINCT",
		nullsNotDistinct,
		1,
	)
	query = strings.Replace(
		query,
		"{{PREDICATE}}",
		predicate,
		1,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("read PostgreSQL indexes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	indexes := make(map[indexKey]*Index)
	for rows.Next() {
		var index Index
		var method string
		if err := rows.Scan(
			&index.Schema,
			&index.TableName,
			&index.Name,
			&method,
			&index.Unique,
			&index.Predicate,
			&index.Tablespace,
			&index.NullsNotDistinct,
		); err != nil {
			return nil, fmt.Errorf("scan PostgreSQL index: %w", err)
		}
		index.TableSchema = index.Schema
		index.Method = ddl.IndexMethod(method)

		tkey := tableKey{schema: index.TableSchema, table: index.TableName}
		if _, ok := tables[tkey]; !ok {
			continue
		}

		key := indexKey{schema: index.Schema, table: index.TableName, index: index.Name}
		indexes[key] = &index
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate PostgreSQL indexes: %w", err)
	}
	return indexes, nil
}

func readPostgreSQLIndexKeys(
	ctx context.Context,
	db Database,
	options postgreSQLOptions,
	indexes map[indexKey]*Index,
	version int,
) error {
	keyLimit := "i.indnkeyatts"
	if version < 110000 {
		keyLimit = "i.indnatts"
	}
	predicate, args := postgreSQLSchemaPredicate("n", options)
	query := strings.Replace(
		getPostgeSQLIndexKeys,
		"KEY_LIMIT",
		keyLimit,
		1,
	)
	query = strings.Replace(
		query,
		"{{PREDICATE}}",
		predicate,
		1,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("read PostgreSQL index keys: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var schema string
		var table string
		var indexName string
		var position int
		var isKey bool
		var expression string
		if err := rows.Scan(&schema, &table, &indexName, &position, &isKey, &expression); err != nil {
			return fmt.Errorf("scan PostgreSQL index key: %w", err)
		}
		index, ok := indexes[indexKey{schema: schema, table: table, index: indexName}]
		if !ok {
			continue
		}
		if isKey {
			index.Keys = append(index.Keys, IndexKey{Expression: expression})
		} else {
			index.Include = append(index.Include, expression)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate PostgreSQL index keys: %w", err)
	}
	return nil
}

func appendPostgreSQLIndexes(tables map[tableKey]*Table, indexes map[indexKey]*Index) {
	keys := make([]indexKey, 0, len(indexes))
	for key := range indexes {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].schema == keys[j].schema {
			if keys[i].table == keys[j].table {
				return keys[i].index < keys[j].index
			}
			return keys[i].table < keys[j].table
		}
		return keys[i].schema < keys[j].schema
	})
	for _, key := range keys {
		table := tables[tableKey{schema: key.schema, table: key.table}]
		if table == nil {
			continue
		}
		table.Indexes = append(table.Indexes, *indexes[key])
	}
}

type tableKey struct {
	schema string
	table  string
}

type indexKey struct {
	schema string
	table  string
	index  string
}

func schemaFromTableMap(tables map[tableKey]*Table) Schema {
	keys := make([]tableKey, 0, len(tables))
	for key := range tables {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].schema == keys[j].schema {
			return keys[i].table < keys[j].table
		}
		return keys[i].schema < keys[j].schema
	})

	schema := Schema{Tables: make([]Table, 0, len(keys))}
	for _, key := range keys {
		schema.Tables = append(schema.Tables, *tables[key])
	}
	return schema
}

func postgreSQLSchemaPredicate(alias string, options postgreSQLOptions) (predicate string, args []any) {
	identifier := alias + ".nspname"
	if options.allSchemas {
		return identifier + " <> 'information_schema' AND pg_catalog.left(" + identifier + ", 3) <> 'pg_'", nil
	}

	schemas := options.schemas
	if len(schemas) == 0 {
		schemas = []string{defaultPostgreSQLSchema}
	}
	placeholders := make([]string, 0, len(schemas))
	args = make([]any, 0, len(schemas))
	for i, schema := range schemas {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, schema)
	}
	return identifier + " IN (" + strings.Join(placeholders, ", ") + ")", args
}

func normalizeSchemaNames(schemas []string) []string {
	seen := make(map[string]struct{}, len(schemas))
	normalized := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		schema = strings.TrimSpace(schema)
		if schema == "" {
			continue
		}
		if _, ok := seen[schema]; ok {
			continue
		}
		seen[schema] = struct{}{}
		normalized = append(normalized, schema)
	}
	return normalized
}

func postgresIdentityKind(code string) IdentityKind {
	switch code {
	case "a":
		return IdentityAlways
	case "d":
		return IdentityByDefault
	default:
		return IdentityNone
	}
}

func postgresGeneratedKind(code string) GeneratedKind {
	switch code {
	case "s":
		return GeneratedStored
	case "v":
		return GeneratedVirtual
	default:
		return GeneratedNone
	}
}

func postgresConstraintKind(code string) ConstraintKind {
	switch code {
	case "p":
		return ConstraintPrimaryKey
	case "u":
		return ConstraintUnique
	case "c":
		return ConstraintCheck
	case "f":
		return ConstraintForeignKey
	default:
		return ConstraintKind(code)
	}
}

func postgresReferenceAction(code string) ddl.ReferenceAction {
	switch code {
	case "r":
		return ddl.Restrict
	case "c":
		return ddl.Cascade
	case "n":
		return ddl.SetNull
	case "d":
		return ddl.SetDefault
	default:
		return ddl.NoAction
	}
}

func splitPostgresList(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, "\x1f")
}
