SELECT
    n.nspname,
    tbl.relname,
    idx.relname,
    pos.ordinality::int,
    pos.ordinality::int <= {{KEY_LIMIT}},
    pg_catalog.pg_get_indexdef(i.indexrelid, pos.ordinality::int, true)
FROM pg_catalog.pg_index AS i
JOIN pg_catalog.pg_class AS tbl ON tbl.oid = i.indrelid
JOIN pg_catalog.pg_namespace AS n ON n.oid = tbl.relnamespace
JOIN pg_catalog.pg_class AS idx ON idx.oid = i.indexrelid
LEFT JOIN pg_catalog.pg_constraint AS con ON con.conindid = i.indexrelid
CROSS JOIN LATERAL pg_catalog.generate_series(1, i.indnatts::int) AS pos(ordinality)
WHERE con.oid IS NULL
  AND NOT i.indisprimary
  AND {{PREDICATE}}
ORDER BY n.nspname, tbl.relname, idx.relname, pos.ordinality;
