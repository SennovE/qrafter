package migrations

import (
	"reflect"
	"strings"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
)

// TableConfigToSchemaTable builds a normalized schema table snapshot from a
// qrafter table configuration and inferred column metadata.
func TableConfigToSchemaTable[T q.TableConfigProvider](d dialect.Renderer) Table {
	t := q.MustNewTable[T]()
	cfg := t.TableConfig()

	table := Table{
		Name:   cfg.Name,
		Schema: cfg.Schema,
	}

	columnsByKey, tagConstraints := infoFromColumnFields(t)
	table.Constraints = appendConstraintsUnique(table.Constraints, tagConstraints...)

	for key, colCfg := range cfg.Columns {
		col := columnsByKey[key.Name]

		if !colCfg.Type.IsZero() {
			col.Type = colCfg.Type
		}
		if colCfg.NotNull {
			col.NotNull = true
		}
		if !colCfg.Default.IsZero() {
			col.HasDefault = true
			sql, err := ddl.Render(d, colCfg.Default)
			if err != nil {
				panic(err)
			}
			col.DefaultExpr = sql
		}
		if colCfg.PrimaryKey {
			col.NotNull = true

			table.Constraints = appendConstraintsUnique(table.Constraints, Constraint{
				Schema:    table.Schema,
				TableName: table.Name,
				Kind:      ConstraintPrimaryKey,
				Columns:   []string{col.Name},
			})
		}

		columnsByKey[key.Name] = col
	}

	table.Columns = make([]Column, 0, len(columnsByKey))
	for _, col := range columnsByKey {
		col.Schema = table.Schema
		col.TableName = table.Name
		table.Columns = append(table.Columns, col)
	}

	table.normalize()
	return table
}

func infoFromColumnFields[T q.TableConfigProvider](table T) (map[string]Column, []Constraint) {
	rt := reflect.TypeOf(table)
	rv := reflect.ValueOf(table)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	result := make(map[string]Column)
	constraints := make([]Constraint, 0)

	columnRefType := reflect.TypeOf((*q.ColumnRef)(nil)).Elem()

	position := 1
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		if !field.Type.Implements(columnRefType) {
			continue
		}

		columnName := rv.Field(i).Interface().(q.ColumnRef).Name()
		qTag := parseQTag(field.Tag.Get("q"))

		col := Column{
			Position: position,
			Name:     columnName,
			Type:     ddlTypeFromGoType(field.Type),
			NotNull:  !isNullableQColumn(field.Type),
		}

		if typeName, ok := qTag["type"]; ok {
			col.Type = parseStringType(typeName)
		}
		if _, ok := qTag["nn"]; ok {
			col.NotNull = true
		}
		if expr, ok := qTag["default"]; ok {
			col.DefaultExpr = expr
			col.HasDefault = true
		}

		if _, ok := qTag["pk"]; ok {
			col.NotNull = true
			constraints = append(constraints, Constraint{
				Kind:    ConstraintPrimaryKey,
				Columns: []string{columnName},
			})
		}

		if _, ok := qTag["uq"]; ok {
			constraints = append(constraints, Constraint{
				Kind:    ConstraintUnique,
				Columns: []string{columnName},
			})
		}

		result[columnName] = col
		position++
	}

	return result, constraints
}

func parseQTag(tag string) map[string]string {
	out := make(map[string]string)
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val := part, ""
		if idx := strings.Index(part, ":"); idx >= 0 {
			key, val = part[:idx], part[idx+1:]
		}
		out[key] = val
	}
	return out
}

func ddlTypeFromGoType(t reflect.Type) ddl.Type {
	arg := strings.TrimPrefix(qColumnTypeArg(t), "*")

	switch arg {
	case "int64":
		return ddl.BigInt()
	case "int", "int32":
		return ddl.Integer()
	case "int16":
		return ddl.SmallInt()
	case "string":
		return ddl.Text()
	case "bool":
		return ddl.Boolean()
	case "[]uint8", "[]byte":
		return ddl.Binary()
	case "time.Time":
		return ddl.TimestampTZ()
	default:
		// enum/custom string aliases
		if strings.HasSuffix(arg, ".UserStatus") || strings.Contains(arg, "Status") {
			return ddl.Text()
		}
		return ddl.SQLType(arg)
	}
}

func isNullableQColumn(t reflect.Type) bool {
	arg := qColumnTypeArg(t)
	return strings.HasPrefix(arg, "*")
}

func parseStringType(typeName string) ddl.Type {
	parsed, ok := postgreSQLType(typeName)
	if ok {
		return parsed
	}
	return ddl.SQLType(typeName)
}

func qColumnTypeArg(t reflect.Type) string {
	s := t.String()
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end < 0 || end <= start {
		return ""
	}
	return s[start+1 : end]
}
