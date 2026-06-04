package ddl

import "github.com/SennovE/qrafter/dialect"

// IndexMethod names an index access method for dialects that support USING.
type IndexMethod string

const (
	// IndexDefault leaves the index method unspecified.
	IndexDefault IndexMethod = ""
	// IndexBTree renders USING btree.
	IndexBTree IndexMethod = "btree"
	// IndexHash renders USING hash.
	IndexHash IndexMethod = "hash"
	// IndexGin renders USING gin.
	IndexGin IndexMethod = "gin"
	// IndexGist renders USING gist.
	IndexGist IndexMethod = "gist"
	// IndexSpGist renders USING spgist.
	IndexSpGist IndexMethod = "spgist"
	// IndexBrin renders USING brin.
	IndexBrin IndexMethod = "brin"
	// IndexFullText renders USING fulltext.
	IndexFullText IndexMethod = "fulltext"
	// IndexSpatial renders USING spatial.
	IndexSpatial IndexMethod = "spatial"
	// IndexColumnstore renders USING columnstore.
	IndexColumnstore IndexMethod = "columnstore"
	// IndexBitmap renders USING bitmap.
	IndexBitmap IndexMethod = "bitmap"
)

// IndexOption stores an index storage option added with CreateIndexStmt.With.
type IndexOption struct {
	name  string
	value any
}

// CreateIndexStmt builds a CREATE INDEX statement.
type CreateIndexStmt struct {
	name  string
	table string

	keys []IndexKey

	options *createIndexOptions
}

type createIndexOptions struct {
	include []Expression

	unique       bool
	ifNotExists  bool
	concurrently bool

	method IndexMethod

	pred *Predicate

	tablespace string
	with       []IndexOption

	clustered        *bool
	invisible        bool
	nullsNotDistinct bool
}

// CreateIndex starts a CREATE INDEX statement.
func CreateIndex(name string) CreateIndexStmt {
	return CreateIndexStmt{name: name}
}

// On sets the indexed table and key expressions.
func (s CreateIndexStmt) On(table string, keys ...IndexKey) CreateIndexStmt {
	s.table = table
	s.keys = append([]IndexKey(nil), keys...)
	return s
}

// OnCols sets the indexed table and simple column keys.
func (s CreateIndexStmt) OnCols(table string, cols ...string) CreateIndexStmt {
	s.table = table
	s.keys = make([]IndexKey, 0, len(cols))
	for _, col := range cols {
		s.keys = append(s.keys, KeyCol(col))
	}
	return s
}

// Unique marks the index as UNIQUE.
func (s CreateIndexStmt) Unique() CreateIndexStmt {
	options := s.cloneOptions()
	options.unique = true
	s.options = options
	return s
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateIndexStmt) IfNotExists() CreateIndexStmt {
	options := s.cloneOptions()
	options.ifNotExists = true
	s.options = options
	return s
}

// Concurrently adds CONCURRENTLY for dialects that support it.
func (s CreateIndexStmt) Concurrently() CreateIndexStmt {
	options := s.cloneOptions()
	options.concurrently = true
	s.options = options
	return s
}

// Using sets the index access method.
func (s CreateIndexStmt) Using(method IndexMethod) CreateIndexStmt {
	options := s.cloneOptions()
	options.method = method
	s.options = options
	return s
}

// Where adds a partial-index predicate.
func (s CreateIndexStmt) Where(pred Predicate) CreateIndexStmt {
	options := s.cloneOptions()
	options.pred = &pred
	s.options = options
	return s
}

// Include adds non-key expressions to INCLUDE.
func (s CreateIndexStmt) Include(exprs ...Expression) CreateIndexStmt {
	options := s.cloneOptions()
	options.include = append([]Expression(nil), exprs...)
	s.options = options
	return s
}

// Tablespace sets a TABLESPACE clause.
func (s CreateIndexStmt) Tablespace(name string) CreateIndexStmt {
	options := s.cloneOptions()
	options.tablespace = name
	s.options = options
	return s
}

// With adds an index storage option.
func (s CreateIndexStmt) With(name string, value any) CreateIndexStmt {
	options := s.cloneOptions()
	options.with = append(options.with, IndexOption{
		name:  name,
		value: value,
	})
	s.options = options
	return s
}

// Clustered marks the index as CLUSTERED.
func (s CreateIndexStmt) Clustered() CreateIndexStmt {
	v := true
	options := s.cloneOptions()
	options.clustered = &v
	s.options = options
	return s
}

// NonClustered marks the index as NONCLUSTERED.
func (s CreateIndexStmt) NonClustered() CreateIndexStmt {
	v := false
	options := s.cloneOptions()
	options.clustered = &v
	s.options = options
	return s
}

// Invisible marks the index as INVISIBLE.
func (s CreateIndexStmt) Invisible() CreateIndexStmt {
	options := s.cloneOptions()
	options.invisible = true
	s.options = options
	return s
}

// NullsNotDistinct adds NULLS NOT DISTINCT.
func (s CreateIndexStmt) NullsNotDistinct() CreateIndexStmt {
	options := s.cloneOptions()
	options.nullsNotDistinct = true
	s.options = options
	return s
}

// Render renders the CREATE INDEX operations.
func (s CreateIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}

func (s CreateIndexStmt) cloneOptions() *createIndexOptions {
	if s.options == nil {
		return &createIndexOptions{}
	}
	options := *s.options
	options.include = append([]Expression(nil), s.options.include...)
	options.with = append([]IndexOption(nil), s.options.with...)
	return &options
}

func renderIndexOptionValue(d dialect.Renderer, v any) string {
	switch v := v.(type) {
	case bool:
		if v {
			return "ON"
		}
		return "OFF"
	case string:
		return d.Literal(v)
	default:
		return d.Literal(v)
	}
}
