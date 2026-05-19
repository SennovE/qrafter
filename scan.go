package qrafter

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
)

var (
	_ sql.Scanner   = (*Column[int])(nil)
	_ driver.Valuer = Column[int]{}
)

// Get returns the column's scanned or assigned value.
func (c Column[T]) Get() T {
	return c.value
}

// Set assigns the column's value.
func (c *Column[T]) Set(value T) {
	c.value = value
}

// Ptr returns a pointer to the column's value.
func (c *Column[T]) Ptr() *T {
	return &c.value
}

// Scan implements sql.Scanner for Column.
func (c *Column[T]) Scan(src any) error {
	var value T
	if err := assignScannedValue(&value, src); err != nil {
		if c.name == "" {
			return err
		}
		return fmt.Errorf("scan column %q: %w", c.name, err)
	}
	c.value = value
	return nil
}

// Value implements driver.Valuer for Column.
func (c Column[T]) Value() (driver.Value, error) {
	if valuer, ok := any(c.value).(driver.Valuer); ok {
		return valuer.Value()
	}
	if valuer, ok := any(&c.value).(driver.Valuer); ok {
		return valuer.Value()
	}
	return driver.DefaultParameterConverter.ConvertValue(c.value)
}

// ScanDest returns scan destinations for exported Column fields in a struct pointer.
func ScanDest(table any) ([]any, error) {
	v := reflect.ValueOf(table)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("scan destination must be a pointer to a struct")
	}

	v = v.Elem()
	destinations := make([]any, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		sf := v.Type().Field(i)
		if !sf.IsExported() {
			continue
		}

		f := v.Field(i)
		if !f.CanAddr() {
			continue
		}

		if scanner, ok := f.Addr().Interface().(sql.Scanner); ok {
			destinations = append(destinations, scanner)
		}
	}

	return destinations, nil
}

func assignScannedValue(dest, src any) error {
	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}
	v := reflect.ValueOf(dest)
	return assignScannedReflectValue(v.Elem(), src)
}

func assignScannedReflectValue(dest reflect.Value, src any) error {
	if src == nil {
		return assignNull(dest)
	}

	if dest.Kind() == reflect.Pointer {
		return assignPointer(dest, src)
	}

	srcValue := reflect.ValueOf(src)
	if assignDirect(dest, src, srcValue) {
		return nil
	}

	if ok, err := assignByKind(dest, src); ok {
		return err
	}

	if srcValue.Type().ConvertibleTo(dest.Type()) {
		dest.Set(srcValue.Convert(dest.Type()))
		return nil
	}

	return fmt.Errorf("unsupported scan: storing driver.Value type %T into type %s", src, dest.Type())
}

func assignPointer(dest reflect.Value, src any) error {
	dest.Set(reflect.New(dest.Type().Elem()))
	return assignScannedValue(dest.Interface(), src)
}

func assignDirect(dest reflect.Value, src any, srcValue reflect.Value) bool {
	if !srcValue.Type().AssignableTo(dest.Type()) {
		return false
	}

	if srcBytes, ok := src.([]byte); ok {
		dest.Set(reflect.ValueOf(cloneBytes(srcBytes)))
		return true
	}

	dest.Set(srcValue)
	return true
}

func assignByKind(dest reflect.Value, src any) (bool, error) {
	switch dest.Kind() {
	case reflect.String:
		return true, assignString(dest, src)
	case reflect.Bool:
		return true, assignBool(dest, src)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, assignInt(dest, src)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true, assignUint(dest, src)
	case reflect.Float32, reflect.Float64:
		return true, assignFloat(dest, src)
	case reflect.Slice:
		return true, assignSlice(dest, src)
	default:
		return false, nil
	}
}

func assignNull(dest reflect.Value) error {
	switch dest.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map:
		dest.Set(reflect.Zero(dest.Type()))
		return nil
	default:
		return fmt.Errorf("cannot scan NULL into %s", dest.Type())
	}
}

func assignString(dest reflect.Value, src any) error {
	switch value := src.(type) {
	case string:
		dest.SetString(value)
		return nil
	case []byte:
		dest.SetString(string(value))
		return nil
	default:
		return fmt.Errorf("unsupported scan: storing driver.Value type %T into type %s", src, dest.Type())
	}
}

func assignBool(dest reflect.Value, src any) error {
	value, err := strconv.ParseBool(asString(src))
	if err != nil {
		return conversionError(src, dest.Type(), err)
	}
	dest.SetBool(value)
	return nil
}

func assignInt(dest reflect.Value, src any) error {
	value, err := strconv.ParseInt(asString(src), 10, dest.Type().Bits())
	if err != nil {
		return conversionError(src, dest.Type(), err)
	}
	dest.SetInt(value)
	return nil
}

func assignUint(dest reflect.Value, src any) error {
	value, err := strconv.ParseUint(asString(src), 10, dest.Type().Bits())
	if err != nil {
		return conversionError(src, dest.Type(), err)
	}
	dest.SetUint(value)
	return nil
}

func assignFloat(dest reflect.Value, src any) error {
	value, err := strconv.ParseFloat(asString(src), dest.Type().Bits())
	if err != nil {
		return conversionError(src, dest.Type(), err)
	}
	dest.SetFloat(value)
	return nil
}

func assignSlice(dest reflect.Value, src any) error {
	if dest.Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("unsupported scan: storing driver.Value type %T into type %s", src, dest.Type())
	}

	switch value := src.(type) {
	case []byte:
		dest.SetBytes(cloneBytes(value))
		return nil
	case string:
		dest.SetBytes([]byte(value))
		return nil
	default:
		return fmt.Errorf("unsupported scan: storing driver.Value type %T into type %s", src, dest.Type())
	}
}

func asString(src any) string {
	switch value := src.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}

	value := reflect.ValueOf(src)
	switch value.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'g', -1, value.Type().Bits())
	default:
		return fmt.Sprintf("%v", src)
	}
}

func cloneBytes(value []byte) []byte {
	clone := make([]byte, len(value))
	copy(clone, value)
	return clone
}

func conversionError(src any, dest reflect.Type, err error) error {
	return fmt.Errorf("converting driver.Value type %T (%q) to %s: %w", src, asString(src), dest, err)
}
