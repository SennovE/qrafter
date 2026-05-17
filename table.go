package qrafter

import (
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
)

// TableConfigProvider is implemented by table model structs.
type TableConfigProvider interface {
	TableConfig() TableConfig
}

// TableRefProvider is implemented by values that expose an explicit table reference.
type TableRefProvider interface {
	TableRef() core.TableRef
}

// TableConfig describes a SQL table.
type TableConfig struct {
	// Name is the SQL table name.
	Name string
}

// TableAlias returns a copy of a table model bound to a SQL alias.
func TableAlias[T TableConfigProvider](table T, alias string) (T, error) {
	config := table.TableConfig()
	aliasedTable, err := bindWithTableRef[T](core.TableRef{Name: config.Name, Alias: alias})
	return aliasedTable, err
}

// GetTableRef returns the SQL table reference for a table model.
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

			if col, ok := f.Interface().(TableRefer); ok {
				return col.TableRef()
			}
		}
	}

	return core.TableRef{Name: table.TableConfig().Name}
}
