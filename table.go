package qrafter

import (
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
)

type TableConfigProvider interface {
	TableConfig() TableConfig
}

type TableRefProvider interface {
	TableRef() core.TableRef
}

type TableConfig struct {
	Name string
}

func TableAlias[T TableConfigProvider](table T, alias string) (T, error) {
	config := table.TableConfig()
	err := bindWithTableRef(&table, core.TableRef{Name: config.Name, Alias: alias})
	return table, err
}

func GetTableRef(table TableConfigProvider) core.TableRef {
	if refProvider, ok := table.(TableRefProvider); ok {
		return refProvider.TableRef()
	}

	v := reflect.ValueOf(table)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			sf := t.Field(i)
			if !sf.IsExported() {
				continue
			}

			f := v.Field(i)

			if col, ok := f.Interface().(ColumnMarker); ok {
				return col.TableRef()
			}
		}
	}

	return core.TableRef{Name: table.TableConfig().Name}
}
