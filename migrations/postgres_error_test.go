package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgreSQLReadQueryErrors(t *testing.T) {
	ctx := context.Background()
	options := postgreSQLOptions{schemas: []string{"public"}}
	tables := map[tableKey]*Table{{schema: "public", table: "users"}: {Schema: "public", Name: "users"}}
	indexes := map[indexKey]*Index{{schema: "public", table: "users", index: "ix_users_id"}: {
		Schema: "public", TableName: "users", Name: "ix_users_id",
	}}

	tests := []struct {
		name string
		run  func(*sql.DB, sqlmock.Sqlmock) error
		want string
	}{
		{
			name: "version",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				_, err := readPostgreSQLVersion(ctx, db)
				return err
			},
			want: "read PostgreSQL server version",
		},
		{
			name: "tables",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				_, err := readPostgreSQLTables(ctx, db, options)
				return err
			},
			want: "read PostgreSQL tables",
		},
		{
			name: "columns",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				return readPostgreSQLColumns(ctx, db, options, tables, 150000)
			},
			want: "read PostgreSQL columns",
		},
		{
			name: "constraints",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				return readPostgreSQLConstraints(ctx, db, options, tables)
			},
			want: "read PostgreSQL constraints",
		},
		{
			name: "index metadata",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				_, err := readPostgreSQLIndexMetadata(ctx, db, options, tables, 150000)
				return err
			},
			want: "read PostgreSQL indexes",
		},
		{
			name: "index keys",
			run: func(db *sql.DB, mock sqlmock.Sqlmock) error {
				mock.ExpectQuery(".*").WillReturnError(fmt.Errorf("boom"))
				return readPostgreSQLIndexKeys(ctx, db, options, indexes, 150000)
			},
			want: "read PostgreSQL index keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer func() { _ = db.Close() }()

			err = tt.run(db, mock)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}
