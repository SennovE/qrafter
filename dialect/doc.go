// Package dialect contains SQL rendering dialects for qrafter queries.
//
// The included dialects cover low-level rendering such as identifier quoting,
// literals, placeholders, and LIMIT/OFFSET syntax. Rendering itself lives in
// qrafter's compilers; dialects override only the focused compiler nodes that
// differ across databases, such as INSERT default rows, UPDATE source tables,
// DELETE source tables, JOIN support, RETURNING clauses, NULL ordering, partial
// indexes, ALTER TABLE operations, and unsupported syntax.
package dialect
