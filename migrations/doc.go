// Package migrations reads database schemas and converts them into ddl builders.
//
// The first implementation targets PostgreSQL introspection. It returns a
// normalized Schema that preserves enough metadata for later diffing and can
// also be converted into ddl.Statements for rendering.
package migrations
