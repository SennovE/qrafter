package qrafter

import (
	"fmt"
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

// NameMapper maps Go struct field names to SQL column names when a field does
// not have a db tag. Set it before calling NewTable or MustNewTable.
var NameMapper = utils.ToSnake

// NewTable creates a table model and binds its exported Column fields.
// Table configuration can come from a TableConfig method or an embedded Table
// field tagged with table:"table_name".
// Column names come from the field's db tag when present; otherwise the Go
// field name is mapped through NameMapper.
func NewTable[T TableConfigProvider]() (T, error) {
	var tmp T
	config := getTableConfig(tmp)
	return bindWithTableRef[T](core.TableRef{Name: config.Name})
}

// MustNewTable is like NewTable but panics if the table cannot be bound.
func MustNewTable[T TableConfigProvider]() T {
	table, err := NewTable[T]()
	if err != nil {
		panic(err)
	}
	return table
}

func bindWithTableRef[T any](tableRef core.TableRef) (T, error) {
	var table T
	if tableRef.Name == "" {
		return table, fmt.Errorf("table name is empty")
	}
	v := reflect.ValueOf(&table).Elem()
	if v.Kind() != reflect.Struct {
		return table, fmt.Errorf("type T must be a struct, got %s", v.Kind())
	}
	setEmbeddedTableConfig(v, &TableConfig{Name: tableRef.Name})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		f := v.Field(i)
		if !f.CanAddr() {
			continue
		}

		col, ok := f.Addr().Interface().(core.ColumnBinder)
		if !ok {
			continue
		}

		col.Bind(columnNameForField(&sf), tableRef)
	}

	return table, nil
}

func columnNameForField(sf *reflect.StructField) string {
	name := sf.Tag.Get("db")
	if name != "" {
		return name
	}
	return NameMapper(sf.Name)
}
