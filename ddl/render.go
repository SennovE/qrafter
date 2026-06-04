package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// Renderer renders a DDL statement.
type Renderer interface {
	Render(d dialect.Renderer) (string, error)
	MustRender(d dialect.Renderer) string
}

// Statements renders a group of DDL statements separated by semicolons.
type Statements []Renderer

type ddlRenderer func(w *strings.Builder, d dialect.Renderer)

type dialectNamer interface {
	DialectName() string
}

// Render renders all statements separated by semicolons.
func (s Statements) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDLs)
}

// MustRender is like Render but panics if rendering fails.
func (s Statements) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDLs)
}

func (s Statements) renderDDLs(w *strings.Builder, d dialect.Renderer) {
	for _, stmt := range s {
		sql := stmt.MustRender(d)
		if sql != "" {
			w.WriteString(sql)
			w.WriteString(";\n")
		}
	}
}

func render(d dialect.Renderer, fn ddlRenderer) (sql string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			sql = ""
			err = panicToError(recovered)
		}
	}()

	var w strings.Builder
	fn(&w, d)
	return w.String(), nil
}

func mustRender(d dialect.Renderer, fn ddlRenderer) string {
	sql, err := render(d, fn)
	if err != nil {
		panic(err)
	}
	return sql
}

func panicToError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}
	return fmt.Errorf("%v", value)
}

func unsupported(d dialect.Renderer, feature string) {
	panic(dialect.UnsupportedFeatureError{Dialect: dialectName(d), Feature: feature})
}

func dialectName(d dialect.Renderer) string {
	unwrapped := dialect.UnwrapRenderer(d)
	if named, ok := unwrapped.(dialectNamer); ok {
		return named.DialectName()
	}

	switch unwrapped.(type) {
	case dialect.PostgreSQL:
		return "PostgreSQL"
	case dialect.MySQL:
		return "MySQL"
	case dialect.SQLite:
		return "SQLite"
	default:
		return "SQL"
	}
}

func isMySQL(d dialect.Renderer) bool {
	_, ok := dialect.UnwrapRenderer(d).(dialect.MySQL)
	return ok
}

func isSQLite(d dialect.Renderer) bool {
	_, ok := dialect.UnwrapRenderer(d).(dialect.SQLite)
	return ok
}
