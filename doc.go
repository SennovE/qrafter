// Package qrafter builds dialect-aware SQL queries from typed Go table structs.
//
// A table model is a struct with qrafter.Column fields and a TableConfig method.
// NewTable binds those fields to SQL column names, and the query builders render
// SQL plus driver arguments for a selected dialect.
//
// Columns can also scan values from database/sql rows, which lets the same
// struct describe both query construction and result destinations.
package qrafter
