package ddl

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/internal/utils"
)

const (
	ddlTagName       = "ddl"
	ddlSkipTag       = "-"
	typeTagPrefix    = "type:"
	typeTagEquals    = "type="
	columnValueField = "value"
)

var (
	bytesType       = reflect.TypeOf([]byte(nil))
	timeType        = reflect.TypeOf(time.Time{})
	nullStringType  = reflect.TypeOf(sql.NullString{})
	nullBoolType    = reflect.TypeOf(sql.NullBool{})
	nullInt64Type   = reflect.TypeOf(sql.NullInt64{})
	nullInt32Type   = reflect.TypeOf(sql.NullInt32{})
	nullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
	nullTimeType    = reflect.TypeOf(sql.NullTime{})
)

// TypeFor returns the default DDL type for a Go type.
func TypeFor[T any]() Type {
	return typeFromGoType(reflect.TypeOf((*T)(nil)).Elem())
}

func columnsFromModel(model any) []ColumnDef {
	v := reflect.ValueOf(model)
	if !v.IsValid() {
		panic(fmt.Errorf("CREATE TABLE model is nil"))
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			panic(fmt.Errorf("CREATE TABLE model is nil"))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("CREATE TABLE model must be a struct, got %s", v.Kind()))
	}

	columns := make([]ColumnDef, 0, v.NumField())
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() || shouldSkipDDLField(&sf) {
			continue
		}

		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}

		column, ok := field.Interface().(q.ColumnRef)
		if !ok {
			continue
		}

		columns = append(columns, ColumnDef{
			name: columnNameFromField(&sf, column),
			typ:  typeFromField(&sf),
		})
	}
	return columns
}

func shouldSkipDDLField(sf *reflect.StructField) bool {
	return strings.TrimSpace(sf.Tag.Get(ddlTagName)) == ddlSkipTag
}

func columnNameFromField(sf *reflect.StructField, column q.ColumnRef) string {
	if name := column.ColumnName(); name != "" {
		return name
	}
	if name := sf.Tag.Get("db"); name != "" {
		return name
	}
	if q.NameMapper == nil {
		return utils.ToSnake(sf.Name)
	}
	return q.NameMapper(sf.Name)
}

func typeFromField(sf *reflect.StructField) Type {
	if tagType, ok := typeFromDDLTag(sf.Tag.Get(ddlTagName)); ok {
		return tagType
	}
	return typeFromColumnType(sf.Type)
}

func typeFromColumnType(t reflect.Type) Type {
	t = indirectType(t)
	if t.Kind() == reflect.Struct {
		if valueField, ok := t.FieldByName(columnValueField); ok {
			return typeFromGoType(valueField.Type)
		}
	}
	panic(fmt.Errorf("cannot infer ddl type from %s; pass ddl.Column(name, type)", t))
}

func typeFromDDLTag(tag string) (Type, bool) {
	tag = strings.TrimSpace(tag)
	if tag == "" || tag == ddlSkipTag {
		return Type{}, false
	}

	typeName := tag
	lowerTypeName := strings.ToLower(typeName)
	if strings.HasPrefix(lowerTypeName, typeTagPrefix) {
		typeName = strings.TrimSpace(typeName[len(typeTagPrefix):])
	}
	if strings.HasPrefix(lowerTypeName, typeTagEquals) {
		typeName = strings.TrimSpace(typeName[len(typeTagEquals):])
	}
	if typeName == "" {
		return Type{}, false
	}
	return SQLType(typeName), true
}

func typeFromGoType(t reflect.Type) Type {
	t = indirectType(t)

	if typ, ok := typeFromKnownGoType(t); ok {
		return typ
	}
	if typ, ok := typeFromGoKind(t); ok {
		return typ
	}
	panic(fmt.Errorf("cannot infer ddl type from Go type %s", t))
}

func typeFromKnownGoType(t reflect.Type) (Type, bool) {
	switch t {
	case bytesType:
		return Binary(), true
	case timeType, nullTimeType:
		return TimestampTZ(), true
	case nullStringType:
		return Text(), true
	case nullBoolType:
		return Boolean(), true
	case nullInt64Type:
		return BigInt(), true
	case nullInt32Type:
		return Integer(), true
	case nullFloat64Type:
		return Double(), true
	}
	return Type{}, false
}

func typeFromGoKind(t reflect.Type) (Type, bool) {
	switch t.Kind() {
	case reflect.Bool:
		return Boolean(), true
	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
		return Integer(), true
	case reflect.Int8, reflect.Int16, reflect.Uint8, reflect.Uint16:
		return SmallInt(), true
	case reflect.Int64, reflect.Uint64, reflect.Uintptr:
		return BigInt(), true
	case reflect.Float32:
		return Float(), true
	case reflect.Float64:
		return Double(), true
	case reflect.String:
		return Text(), true
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return Binary(), true
		}
	}
	return Type{}, false
}

func indirectType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
