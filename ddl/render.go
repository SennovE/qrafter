package ddl

import (
	"fmt"

	"github.com/SennovE/qrafter/dialect"
)

// Renderer renders a DDL statement.
type Renderer interface {
	Render(d dialect.Renderer) (string, error)
	MustRender(d dialect.Renderer) string
}

// Statements renders a group of DDL statements separated by semicolons.
type Statements []Renderer

// Render renders all statements separated by semicolons.
func (s Statements) Render(d dialect.Renderer) (string, error) {
	return render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s Statements) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}

func render(d dialect.Renderer, node any) (sql string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			sql = ""
			err = panicToError(recovered)
		}
	}()

	compiler := newCompiler(d)
	compiler.Compile(node)
	return compiler.SQL(), nil
}

func mustRender(d dialect.Renderer, node any) string {
	sql, err := render(d, node)
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
