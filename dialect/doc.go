// Package dialect contains SQL rendering dialects for qrafter queries.
//
// The included dialects cover low-level rendering such as identifier quoting,
// literals, placeholders, and LIMIT/OFFSET syntax. Dialects can also override
// focused feature hooks for clauses that differ across databases, such as
// INSERT default rows, UPDATE source tables, DELETE source tables, JOIN support,
// RETURNING clauses, NULL ordering, and unsupported syntax.
package dialect
