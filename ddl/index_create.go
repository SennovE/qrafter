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
	Name  string
	Value any
}

// CreateIndexStmt builds a CREATE INDEX statement.
type CreateIndexStmt struct {
	Name  string
	Table string

	Keys []IndexKey

	Options *CreateIndexOptions
}

// CreateIndexOptions stores optional CREATE INDEX clauses.
type CreateIndexOptions struct {
	Include []Expression

	Unique       bool
	IfNotExists  bool
	Concurrently bool

	Method IndexMethod

	Predicate *Predicate

	Tablespace string
	With       []IndexOption

	Clustered        *bool
	Invisible        bool
	NullsNotDistinct bool
}

// Index starts a table-detached CREATE INDEX statement.
func Index(name string, keys ...IndexKey) CreateIndexStmt {
	return CreateIndex(name).On("", keys...)
}

// IndexCols starts a table-detached CREATE INDEX statement on simple columns.
func IndexCols(name string, cols ...string) CreateIndexStmt {
	return CreateIndex(name).OnCols("", cols...)
}

// CreateIndex starts a CREATE INDEX statement.
func CreateIndex(name string) CreateIndexStmt {
	return CreateIndexStmt{Name: name}
}

// On sets the indexed table and key expressions.
func (s CreateIndexStmt) On(table string, keys ...IndexKey) CreateIndexStmt {
	s.Table = table
	s.Keys = append([]IndexKey(nil), keys...)
	return s
}

// OnCols sets the indexed table and simple column keys.
func (s CreateIndexStmt) OnCols(table string, cols ...string) CreateIndexStmt {
	s.Table = table
	s.Keys = make([]IndexKey, 0, len(cols))
	for _, col := range cols {
		s.Keys = append(s.Keys, KeyCol(col))
	}
	return s
}

// Unique marks the index as UNIQUE.
func (s CreateIndexStmt) Unique() CreateIndexStmt {
	options := s.cloneOptions()
	options.Unique = true
	s.Options = options
	return s
}

// IfNotExists adds IF NOT EXISTS.
func (s CreateIndexStmt) IfNotExists() CreateIndexStmt {
	options := s.cloneOptions()
	options.IfNotExists = true
	s.Options = options
	return s
}

// Concurrently adds CONCURRENTLY for dialects that support it.
func (s CreateIndexStmt) Concurrently() CreateIndexStmt {
	options := s.cloneOptions()
	options.Concurrently = true
	s.Options = options
	return s
}

// Using sets the index access method.
func (s CreateIndexStmt) Using(method IndexMethod) CreateIndexStmt {
	options := s.cloneOptions()
	options.Method = method
	s.Options = options
	return s
}

// Where adds a partial-index predicate.
func (s CreateIndexStmt) Where(pred Predicate) CreateIndexStmt {
	options := s.cloneOptions()
	options.Predicate = &pred
	s.Options = options
	return s
}

// Include adds non-key expressions to INCLUDE.
func (s CreateIndexStmt) Include(exprs ...Expression) CreateIndexStmt {
	options := s.cloneOptions()
	options.Include = append([]Expression(nil), exprs...)
	s.Options = options
	return s
}

// Tablespace sets a TABLESPACE clause.
func (s CreateIndexStmt) Tablespace(name string) CreateIndexStmt {
	options := s.cloneOptions()
	options.Tablespace = name
	s.Options = options
	return s
}

// With adds an index storage option.
func (s CreateIndexStmt) With(name string, value any) CreateIndexStmt {
	options := s.cloneOptions()
	options.With = append(options.With, IndexOption{
		Name:  name,
		Value: value,
	})
	s.Options = options
	return s
}

// Clustered marks the index as CLUSTERED.
func (s CreateIndexStmt) Clustered() CreateIndexStmt {
	v := true
	options := s.cloneOptions()
	options.Clustered = &v
	s.Options = options
	return s
}

// NonClustered marks the index as NONCLUSTERED.
func (s CreateIndexStmt) NonClustered() CreateIndexStmt {
	v := false
	options := s.cloneOptions()
	options.Clustered = &v
	s.Options = options
	return s
}

// Invisible marks the index as INVISIBLE.
func (s CreateIndexStmt) Invisible() CreateIndexStmt {
	options := s.cloneOptions()
	options.Invisible = true
	s.Options = options
	return s
}

// NullsNotDistinct adds NULLS NOT DISTINCT.
func (s CreateIndexStmt) NullsNotDistinct() CreateIndexStmt {
	options := s.cloneOptions()
	options.NullsNotDistinct = true
	s.Options = options
	return s
}

// Render renders the CREATE INDEX operations.
func (s CreateIndexStmt) Render(d dialect.Renderer) (string, error) {
	return Render(d, s)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s)
}

func (s CreateIndexStmt) cloneOptions() *CreateIndexOptions {
	if s.Options == nil {
		return &CreateIndexOptions{}
	}
	options := *s.Options
	options.Include = append([]Expression(nil), s.Options.Include...)
	options.With = append([]IndexOption(nil), s.Options.With...)
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
