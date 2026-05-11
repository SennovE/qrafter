package qrafter

import (
	"fmt"
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

func Bind[T TableConfigProvider](table T) error {
	config := table.TableConfig()
	return bindWithTableRef(table, core.TableRef{Name: config.Name})
}

func bindWithTableRef[T any](table T, tableRef core.TableRef) error {
	if tableRef.Name == "" {
		return fmt.Errorf("table name is empty")
	}
	v := reflect.ValueOf(table)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("table must be a pointer to a struct")
	}

	v = v.Elem()
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

		name := sf.Tag.Get("db")
		if name == "" {
			name = utils.ToSnake(sf.Name)
		}

		col.Bind(name, tableRef)
	}

	return nil
}
