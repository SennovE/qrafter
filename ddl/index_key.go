package ddl

type sortDirection string

const (
	sortAsc  sortDirection = "ASC"
	sortDesc sortDirection = "DESC"
)

type nullsOrder string

const (
	nullsFirst nullsOrder = "FIRST"
	nullsLast  nullsOrder = "LAST"
)

// IndexKey describes one expression or column inside an index key list.
type IndexKey struct {
	expr    Expression
	options *indexKeyOptions
}

type indexKeyOptions struct {
	sort      sortDirection
	nulls     nullsOrder
	collation string
	opclass   string
	length    int
}

// Key creates an index key from an expression.
func Key(expr Expression) IndexKey {
	return IndexKey{expr: expr}
}

// KeyCol creates an index key from an unqualified column name.
func KeyCol(name string) IndexKey {
	return IndexKey{expr: Col(name)}
}

// Asc marks the key as ascending.
func (k IndexKey) Asc() IndexKey {
	options := k.cloneOptions()
	options.sort = sortAsc
	k.options = options
	return k
}

// Desc marks the key as descending.
func (k IndexKey) Desc() IndexKey {
	options := k.cloneOptions()
	options.sort = sortDesc
	k.options = options
	return k
}

// NullsFirst adds NULLS FIRST ordering.
func (k IndexKey) NullsFirst() IndexKey {
	options := k.cloneOptions()
	options.nulls = nullsFirst
	k.options = options
	return k
}

// NullsLast adds NULLS LAST ordering.
func (k IndexKey) NullsLast() IndexKey {
	options := k.cloneOptions()
	options.nulls = nullsLast
	k.options = options
	return k
}

// Collate adds a collation name to the key.
func (k IndexKey) Collate(name string) IndexKey {
	options := k.cloneOptions()
	options.collation = requireName("collation", name)
	k.options = options
	return k
}

// OpClass adds an operator class to the key.
func (k IndexKey) OpClass(name string) IndexKey {
	options := k.cloneOptions()
	options.opclass = requireName("operator class", name)
	k.options = options
	return k
}

// PrefixLength adds a prefix length for dialects that support it.
func (k IndexKey) PrefixLength(n int) IndexKey {
	if n <= 0 {
		panic("ddl: index prefix length must be positive")
	}
	options := k.cloneOptions()
	options.length = n
	k.options = options
	return k
}

func (k IndexKey) cloneOptions() *indexKeyOptions {
	if k.options == nil {
		return &indexKeyOptions{}
	}
	options := *k.options
	return &options
}
