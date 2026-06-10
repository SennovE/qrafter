package migrations

import (
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

var postgreSQLSimpleTypes = map[string]func() ddl.Type{
	"smallint":                    ddl.SmallInt,
	"int2":                        ddl.SmallInt,
	"integer":                     ddl.Integer,
	"int":                         ddl.Integer,
	"int4":                        ddl.Integer,
	"bigint":                      ddl.BigInt,
	"int8":                        ddl.BigInt,
	"text":                        ddl.Text,
	"boolean":                     ddl.Boolean,
	"bool":                        ddl.Boolean,
	"date":                        ddl.Date,
	"time":                        ddl.Time,
	"time without time zone":      ddl.Time,
	"timestamp":                   ddl.Timestamp,
	"timestamp without time zone": ddl.Timestamp,
	"timestamp with time zone":    ddl.TimestampTZ,
	"timestamptz":                 ddl.TimestampTZ,
	"uuid":                        ddl.UUID,
	"json":                        ddl.JSON,
	"jsonb":                       ddl.JSONB,
	"bytea":                       ddl.Binary,
	"real":                        ddl.Float,
	"float4":                      ddl.Float,
	"double precision":            ddl.Double,
	"float8":                      ddl.Double,
}

func postgreSQLType(typeName string) (ddl.Type, bool) {
	normalized := normalizePostgreSQLType(typeName)
	if typeFunc, ok := postgreSQLSimpleTypes[normalized]; ok {
		return typeFunc(), true
	}
	if size, ok := parsePostgreSQLSingleArgType(normalized, "character varying"); ok {
		return ddl.VarChar(size), true
	}
	if size, ok := parsePostgreSQLSingleArgType(normalized, "varchar"); ok {
		return ddl.VarChar(size), true
	}
	if size, ok := parsePostgreSQLSingleArgType(normalized, "character"); ok {
		return ddl.Char(size), true
	}
	if size, ok := parsePostgreSQLSingleArgType(normalized, "char"); ok {
		return ddl.Char(size), true
	}
	if precision, scale, ok := parsePostgreSQLNumericType(normalized); ok {
		return ddl.Numeric(precision, scale), true
	}
	return ddl.SQLType(typeName), false
}

func normalizePostgreSQLType(typeName string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(typeName)))
	return strings.Join(fields, " ")
}

func parsePostgreSQLSingleArgType(typeName, prefix string) (int, bool) {
	inner, ok := parsePostgreSQLTypeArgs(typeName, prefix)
	if !ok {
		return 0, false
	}
	size, err := strconv.Atoi(strings.TrimSpace(inner))
	return size, err == nil && size > 0
}

func parsePostgreSQLNumericType(typeName string) (precision, scale int, ok bool) {
	var inner string
	inner, ok = parsePostgreSQLTypeArgs(typeName, "numeric")
	if !ok {
		return 0, 0, false
	}
	parts := strings.Split(inner, ",")
	if len(parts) == 0 || len(parts) > 2 {
		return 0, 0, false
	}
	precision, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || precision <= 0 {
		return 0, 0, false
	}
	if len(parts) == 1 {
		return precision, 0, true
	}
	scale, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || scale < 0 {
		return 0, 0, false
	}
	return precision, scale, true
}

func parsePostgreSQLTypeArgs(typeName, prefix string) (string, bool) {
	if !strings.HasPrefix(typeName, prefix+"(") || !strings.HasSuffix(typeName, ")") {
		return "", false
	}
	return typeName[len(prefix)+1 : len(typeName)-1], true
}
