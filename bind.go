package qrafter

import (
	"fmt"
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
	"github.com/SennovE/qrafter/internal/utils"
)

func NewTable[T TableConfigProvider]() (T, error) {
	var tmp T
	config := tmp.TableConfig() 
	table, err := bindWithTableRef[T](core.TableRef{Name: config.Name})
	return table, err
}

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
		return table, fmt.Errorf("T must be a struct, got %s", v.Kind())
	}
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

	return table, nil
}
