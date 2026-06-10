package ddl

// SortDirection is an ASC or DESC index key direction.
type SortDirection string

const (
	// SortAsc renders ASC.
	SortAsc SortDirection = "ASC"
	// SortDesc renders DESC.
	SortDesc SortDirection = "DESC"
)

// NullsOrder describes NULLS FIRST or NULLS LAST index key ordering.
type NullsOrder string

const (
	// NullsFirstOrder renders NULLS FIRST.
	NullsFirstOrder NullsOrder = "FIRST"
	// NullsLastOrder renders NULLS LAST.
	NullsLastOrder NullsOrder = "LAST"
)

// IndexKey describes one expression or column inside an index key list.
type IndexKey struct {
	Expr    Expression
	Options *IndexKeyOptions
}

// IndexKeyOptions stores optional index key clauses.
type IndexKeyOptions struct {
	Sort      SortDirection
	Nulls     NullsOrder
	Collation string
	OpClass   string
	Length    int
}

// Key creates an index key from an expression.
func Key(expr Expression) IndexKey {
	return IndexKey{Expr: expr}
}

// KeyCol creates an index key from an unqualified column name.
func KeyCol(name string) IndexKey {
	return IndexKey{Expr: Col(name)}
}

// Asc marks the key as ascending.
func (k IndexKey) Asc() IndexKey {
	options := k.cloneOptions()
	options.Sort = SortAsc
	k.Options = options
	return k
}

// Desc marks the key as descending.
func (k IndexKey) Desc() IndexKey {
	options := k.cloneOptions()
	options.Sort = SortDesc
	k.Options = options
	return k
}

// NullsFirst adds NULLS FIRST ordering.
func (k IndexKey) NullsFirst() IndexKey {
	options := k.cloneOptions()
	options.Nulls = NullsFirstOrder
	k.Options = options
	return k
}

// NullsLast adds NULLS LAST ordering.
func (k IndexKey) NullsLast() IndexKey {
	options := k.cloneOptions()
	options.Nulls = NullsLastOrder
	k.Options = options
	return k
}

// Collate adds a collation name to the key.
func (k IndexKey) Collate(name string) IndexKey {
	options := k.cloneOptions()
	options.Collation = name
	k.Options = options
	return k
}

// OpClass adds an operator class to the key.
func (k IndexKey) OpClass(name string) IndexKey {
	options := k.cloneOptions()
	options.OpClass = name
	k.Options = options
	return k
}

// PrefixLength adds a prefix length for dialects that support it.
func (k IndexKey) PrefixLength(n int) IndexKey {
	if n <= 0 {
		panic("ddl: index prefix length must be positive")
	}
	options := k.cloneOptions()
	options.Length = n
	k.Options = options
	return k
}

func (k IndexKey) cloneOptions() *IndexKeyOptions {
	if k.Options == nil {
		return &IndexKeyOptions{}
	}
	options := *k.Options
	return &options
}
