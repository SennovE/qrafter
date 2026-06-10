package migrations

import (
	"sort"
	"strings"
)

type tableKey struct {
	schema string
	table  string
}

type indexKey struct {
	schema string
	table  string
	index  string
}

func schemaFromTableMap(tables map[tableKey]*Table) Schema {
	keys := make([]tableKey, 0, len(tables))
	for key := range tables {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].schema == keys[j].schema {
			return keys[i].table < keys[j].table
		}
		return keys[i].schema < keys[j].schema
	})

	schema := Schema{Tables: make([]Table, 0, len(keys))}
	for _, key := range keys {
		schema.Tables = append(schema.Tables, *tables[key])
	}
	return schema
}

func tablesByKey(tables []Table) map[tableKey]*Table {
	out := make(map[tableKey]*Table, len(tables))
	for i := range tables {
		table := &tables[i]
		out[tableKey{schema: table.Schema, table: table.Name}] = table
	}
	return out
}

func constraintKey(c *Constraint) string {
	var b strings.Builder

	b.WriteString(string(c.Kind))
	b.WriteString("|cols=")
	b.WriteString(strings.Join(c.Columns, ","))

	switch c.Kind {
	case ConstraintCheck:
		b.WriteString("|check=")
		b.WriteString(normalizeSQL(c.CheckExpr))

	case ConstraintForeignKey:
		b.WriteString("|ref=")
		b.WriteString(c.Reference.Schema)
		b.WriteByte('.')
		b.WriteString(c.Reference.TableName)
		b.WriteByte('(')
		b.WriteString(strings.Join(c.Reference.Columns, ","))
		b.WriteByte(')')
		b.WriteString("|del=")
		b.WriteString(string(normalizeReferenceAction(c.OnDelete)))
		b.WriteString("|upd=")
		b.WriteString(string(normalizeReferenceAction(c.OnUpdate)))
	}

	return b.String()
}
