package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

// Type describes a SQL column type.
type Type struct {
	name        string
	dialectName map[string]string
}

// SQLType creates a custom SQL type name.
func SQLType(name string) Type {
	return Type{name: requireName("type", name)}
}

// ForDialect returns a copy of the type with a dialect-specific name.
func (t Type) ForDialect(dialectName, name string) Type {
	if t.dialectName == nil {
		t.dialectName = make(map[string]string)
	} else {
		copied := make(map[string]string, len(t.dialectName)+1)
		for key, value := range t.dialectName {
			copied[key] = value
		}
		t.dialectName = copied
	}
	t.dialectName[requireName("dialect", dialectName)] = requireName("type", name)
	return t
}

func (t Type) render(d dialect.Renderer) string {
	if t.name == "" {
		panic(fmt.Errorf("ddl type is empty"))
	}

	name := dialectName(d)
	for dialectName, typeName := range t.dialectName {
		if strings.EqualFold(dialectName, name) {
			return typeName
		}
	}
	return t.name
}

// SmallInt returns a SMALLINT type.
func SmallInt() Type {
	return SQLType("SMALLINT")
}

// Integer returns an INTEGER type.
func Integer() Type {
	return SQLType("INTEGER")
}

// BigInt returns a BIGINT type.
func BigInt() Type {
	return SQLType("BIGINT")
}

// Serial returns an auto-incrementing integer type.
func Serial() Type {
	return SQLType("INTEGER").
		ForDialect("PostgreSQL", "SERIAL").
		ForDialect("MySQL", "INT AUTO_INCREMENT").
		ForDialect("SQLite", "INTEGER")
}

// BigSerial returns an auto-incrementing big integer type.
func BigSerial() Type {
	return SQLType("BIGINT").
		ForDialect("PostgreSQL", "BIGSERIAL").
		ForDialect("MySQL", "BIGINT AUTO_INCREMENT").
		ForDialect("SQLite", "INTEGER")
}

// Text returns a TEXT type.
func Text() Type {
	return SQLType("TEXT")
}

// VarChar returns a VARCHAR type.
func VarChar(size int) Type {
	if size <= 0 {
		panic(fmt.Errorf("varchar size must be positive"))
	}
	return SQLType(fmt.Sprintf("VARCHAR(%d)", size))
}

// Char returns a CHAR type.
func Char(size int) Type {
	if size <= 0 {
		panic(fmt.Errorf("char size must be positive"))
	}
	return SQLType(fmt.Sprintf("CHAR(%d)", size))
}

// Boolean returns a BOOLEAN type.
func Boolean() Type {
	return SQLType("BOOLEAN")
}

// Date returns a DATE type.
func Date() Type {
	return SQLType("DATE")
}

// Time returns a TIME type.
func Time() Type {
	return SQLType("TIME")
}

// Timestamp returns a TIMESTAMP type.
func Timestamp() Type {
	return SQLType("TIMESTAMP")
}

// TimestampTZ returns a timestamp-with-time-zone type.
func TimestampTZ() Type {
	return SQLType("TIMESTAMP WITH TIME ZONE").
		ForDialect("PostgreSQL", "TIMESTAMPTZ").
		ForDialect("MySQL", "TIMESTAMP").
		ForDialect("SQLite", "TEXT")
}

// Numeric returns a NUMERIC type.
func Numeric(precision, scale int) Type {
	if precision <= 0 {
		panic(fmt.Errorf("numeric precision must be positive"))
	}
	if scale < 0 {
		panic(fmt.Errorf("numeric scale cannot be negative"))
	}
	if scale == 0 {
		return SQLType(fmt.Sprintf("NUMERIC(%d)", precision))
	}
	return SQLType(fmt.Sprintf("NUMERIC(%d, %d)", precision, scale))
}

// UUID returns a UUID type.
func UUID() Type {
	return SQLType("UUID").
		ForDialect("MySQL", "CHAR(36)").
		ForDialect("SQLite", "TEXT")
}

// JSON returns a JSON type.
func JSON() Type {
	return SQLType("JSON").
		ForDialect("SQLite", "TEXT")
}

// JSONB returns a JSONB type.
func JSONB() Type {
	return SQLType("JSONB").
		ForDialect("MySQL", "JSON").
		ForDialect("SQLite", "TEXT")
}

// Binary returns a binary large object type.
func Binary() Type {
	return SQLType("BYTEA").
		ForDialect("MySQL", "BLOB").
		ForDialect("SQLite", "BLOB")
}

// Float returns a single-precision floating point type.
func Float() Type {
	return SQLType("REAL").
		ForDialect("MySQL", "FLOAT")
}

// Double returns a double-precision floating point type.
func Double() Type {
	return SQLType("DOUBLE PRECISION").
		ForDialect("MySQL", "DOUBLE").
		ForDialect("SQLite", "REAL")
}
