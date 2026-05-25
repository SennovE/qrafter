// Package ddl builds SQL data definition statements.
//
// The package is intentionally separate from the root qrafter query DSL:
// SELECT/INSERT/UPDATE/DELETE live in qrafter, while schema statements such as
// CREATE TABLE, ALTER TABLE, and CREATE INDEX live here.
//
// DDL statements render SQL text for a selected dialect. Render returns an error
// when a dialect cannot safely render the requested feature; MustRender is
// available for examples and tests that should fail fast.
package ddl
