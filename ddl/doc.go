// Package ddl builds SQL data definition statements.
//
// The package is intentionally separate from the root qrafter query DSL:
// SELECT/INSERT/UPDATE/DELETE live in qrafter, while schema statements such as
// CREATE TABLE, ALTER TABLE, and CREATE INDEX live here.
//
// DDL builders keep statement state only. A centralized compiler renders the
// SQL text and lets dialects override focused nodes such as partial indexes,
// DROP TABLE behavior, index renames, and ALTER TABLE operations. Render returns
// an error when a dialect cannot safely render the requested feature; MustRender
// is available for examples and tests that should fail fast.
package ddl
