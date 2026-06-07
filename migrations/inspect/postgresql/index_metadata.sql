SELECT
    n.nspname,
    tbl.relname,
    idx.relname,
    am.amname,
    i.indisunique,
    COALESCE(pg_catalog.pg_get_expr(i.indpred, i.indrelid, true), ''),
    COALESCE(ts.spcname, ''),
    {{NULLS_NOT_DISTINCT}}
FROM pg_catalog.pg_index AS i
JOIN pg_catalog.pg_class AS tbl ON tbl.oid = i.indrelid
JOIN pg_catalog.pg_namespace AS n ON n.oid = tbl.relnamespace
JOIN pg_catalog.pg_class AS idx ON idx.oid = i.indexrelid
JOIN pg_catalog.pg_am AS am ON am.oid = idx.relam
LEFT JOIN pg_catalog.pg_tablespace AS ts ON ts.oid = idx.reltablespace
LEFT JOIN pg_catalog.pg_constraint AS con ON con.conindid = i.indexrelid
WHERE con.oid IS NULL
  AND NOT i.indisprimary
  AND {{PREDICATE}}
ORDER BY n.nspname, tbl.relname, idx.relname;
