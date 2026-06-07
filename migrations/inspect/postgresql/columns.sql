SELECT
    n.nspname,
    c.relname,
    a.attnum::int,
    a.attname,
    pg_catalog.format_type(a.atttypid, a.atttypmod),
    a.attnotnull,
    pg_catalog.pg_get_expr(ad.adbin, ad.adrelid, true),
    a.attidentity::text,
    {{GENERATED}}
FROM pg_catalog.pg_attribute AS a
JOIN pg_catalog.pg_class AS c ON c.oid = a.attrelid
JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
LEFT JOIN pg_catalog.pg_attrdef AS ad ON ad.adrelid = a.attrelid AND ad.adnum = a.attnum
WHERE c.relkind IN ('r', 'p')
  AND a.attnum > 0
  AND NOT a.attisdropped
  AND {{PREDICATE}}
ORDER BY n.nspname, c.relname, a.attnum;
