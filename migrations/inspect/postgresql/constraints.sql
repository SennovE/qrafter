SELECT
    n.nspname,
    tbl.relname,
    con.conname,
    con.contype::text,
    COALESCE((
        SELECT pg_catalog.string_agg(att.attname, pg_catalog.chr(31) ORDER BY keys.ordinality)
        FROM pg_catalog.unnest(con.conkey) WITH ORDINALITY AS keys(attnum, ordinality)
        JOIN pg_catalog.pg_attribute AS att ON att.attrelid = con.conrelid AND att.attnum = keys.attnum
    ), ''),
    COALESCE(refn.nspname, ''),
    COALESCE(reft.relname, ''),
    COALESCE((
        SELECT pg_catalog.string_agg(att.attname, pg_catalog.chr(31) ORDER BY keys.ordinality)
        FROM pg_catalog.unnest(con.confkey) WITH ORDINALITY AS keys(attnum, ordinality)
        JOIN pg_catalog.pg_attribute AS att ON att.attrelid = con.confrelid AND att.attnum = keys.attnum
    ), ''),
    con.confdeltype::text,
    con.confupdtype::text,
    CASE WHEN con.contype = 'c'
        THEN pg_catalog.pg_get_expr(con.conbin, con.conrelid, true)
        ELSE ''
    END
FROM pg_catalog.pg_constraint AS con
JOIN pg_catalog.pg_class AS tbl ON tbl.oid = con.conrelid
JOIN pg_catalog.pg_namespace AS n ON n.oid = tbl.relnamespace
LEFT JOIN pg_catalog.pg_class AS reft ON reft.oid = con.confrelid
LEFT JOIN pg_catalog.pg_namespace AS refn ON refn.oid = reft.relnamespace
WHERE con.contype IN ('p', 'u', 'f', 'c')
  AND {{PREDICATE}}
ORDER BY n.nspname, tbl.relname, con.contype, con.conname;
