package qrafter

import (
	"reflect"

	"github.com/SennovE/qrafter/internal/core"
)

// TableConfigProvider is implemented by table model structs with a TableConfig method.
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

// Table can be embedded into table models to provide table configuration.
type Table struct {
	config TableConfig
}

// TableConfig returns the embedded table configuration.
func (t Table) TableConfig() TableConfig {
	return t.config
}

// TableAlias returns a copy of a table model bound to a SQL alias.
func TableAlias[T TableConfigProvider](table T, alias string) (T, error) {
	config := getTableConfig(table)
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

	return core.TableRef{Name: getTableConfig(table).Name}
}

func getTableConfig(table TableConfigProvider) TableConfig {
	config := table.TableConfig()
	if config.Name != "" {
		return config
	}

	embeddedConfig, hasEmbedded := embeddedTableConfig(table)
	if hasEmbedded && embeddedConfig.Name != "" {
		return embeddedConfig
	}

	return config
}

func embeddedTableConfig(table any) (TableConfig, bool) {
	v := reflect.ValueOf(table)
	if !v.IsValid() {
		return TableConfig{}, false
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return TableConfig{}, false
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return TableConfig{}, false
	}

	return embeddedTableConfigFromValue(v)
}

func embeddedTableConfigFromValue(v reflect.Value) (TableConfig, bool) {
	tableType := reflect.TypeOf(Table{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.Anonymous || !isTableField(sf.Type, tableType) {
			continue
		}

		config := tableConfigFromFieldValue(v.Field(i))
		if config.Name == "" {
			config.Name = sf.Tag.Get("table")
		}

		return config, true
	}

	return TableConfig{}, false
}

func tableConfigFromFieldValue(v reflect.Value) TableConfig {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return TableConfig{}
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Type() != reflect.TypeOf(Table{}) {
		return TableConfig{}
	}
	return v.Interface().(Table).TableConfig()
}

func setEmbeddedTableConfig(v reflect.Value, config TableConfig) {
	tableType := reflect.TypeOf(Table{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.Anonymous || !isTableField(sf.Type, tableType) {
			continue
		}

		setTableConfigField(v.Field(i), config)
		return
	}
}

func setTableConfigField(v reflect.Value, config TableConfig) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.IsValid() && v.CanSet() && v.Type() == reflect.TypeOf(Table{}) {
		v.Set(reflect.ValueOf(Table{config: config}))
	}
}

func isTableField(fieldType, tableType reflect.Type) bool {
	return fieldType == tableType ||
		(fieldType.Kind() == reflect.Pointer && fieldType.Elem() == tableType)
}
