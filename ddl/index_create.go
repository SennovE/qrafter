package ddl

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

type IndexMethod string

const (
	IndexDefault     IndexMethod = ""
	IndexBTree       IndexMethod = "btree"
	IndexHash        IndexMethod = "hash"
	IndexGin         IndexMethod = "gin"
	IndexGist        IndexMethod = "gist"
	IndexSpGist      IndexMethod = "spgist"
	IndexBrin        IndexMethod = "brin"
	IndexFullText    IndexMethod = "fulltext"
	IndexSpatial     IndexMethod = "spatial"
	IndexColumnstore IndexMethod = "columnstore"
	IndexBitmap      IndexMethod = "bitmap"
)

type IndexOption struct {
	name  string
	value any
}

type CreateIndexStmt struct {
	name  string
	table string

	keys    []IndexKey
	include []Expression

	unique       bool
	ifNotExists  bool
	concurrently bool

	method IndexMethod

	pred *Predicate

	tablespace string
	options    []IndexOption

	// dialect-specific common flags
	clustered        *bool
	invisible        bool
	nullsNotDistinct bool
}

func CreateIndex(name string) CreateIndexStmt {
	return CreateIndexStmt{name: name}
}

func (s CreateIndexStmt) On(table string, keys ...IndexKey) CreateIndexStmt {
	s.table = table
	s.keys = keys
	return s
}

func (s CreateIndexStmt) OnCols(table string, cols ...string) CreateIndexStmt {
	s.table = table
	s.keys = make([]IndexKey, 0, len(cols))
	for _, col := range cols {
		s.keys = append(s.keys, KeyCol(col))
	}
	return s
}

func (s CreateIndexStmt) Unique() CreateIndexStmt {
	s.unique = true
	return s
}

func (s CreateIndexStmt) IfNotExists() CreateIndexStmt {
	s.ifNotExists = true
	return s
}

func (s CreateIndexStmt) Concurrently() CreateIndexStmt {
	s.concurrently = true
	return s
}

func (s CreateIndexStmt) Using(method IndexMethod) CreateIndexStmt {
	s.method = method
	return s
}

func (s CreateIndexStmt) Where(pred Predicate) CreateIndexStmt {
	s.pred = &pred
	return s
}

func (s CreateIndexStmt) Include(exprs ...Expression) CreateIndexStmt {
	s.include = exprs
	return s
}

func (s CreateIndexStmt) Tablespace(name string) CreateIndexStmt {
	s.tablespace = name
	return s
}

func (s CreateIndexStmt) With(name string, value any) CreateIndexStmt {
	s.options = append(s.options, IndexOption{
		name:  name,
		value: value,
	})
	return s
}

func (s CreateIndexStmt) Clustered() CreateIndexStmt {
	v := true
	s.clustered = &v
	return s
}

func (s CreateIndexStmt) NonClustered() CreateIndexStmt {
	v := false
	s.clustered = &v
	return s
}

func (s CreateIndexStmt) Invisible() CreateIndexStmt {
	s.invisible = true
	return s
}

func (s CreateIndexStmt) NullsNotDistinct() CreateIndexStmt {
	s.nullsNotDistinct = true
	return s
}

// Render renders the CREATE INDEX operations.
func (s CreateIndexStmt) Render(d dialect.Renderer) (string, error) {
	return render(d, s.renderDDL)
}

// MustRender is like Render but panics if rendering fails.
func (s CreateIndexStmt) MustRender(d dialect.Renderer) string {
	return mustRender(d, s.renderDDL)
}

func (s CreateIndexStmt) renderDDL(w *strings.Builder, d dialect.Renderer) {
	if len(s.keys) == 0 {
		panic(fmt.Errorf("CREATE INDEX %q must include at least one key", s.name))
	}

	w.WriteString("CREATE ")

	if s.unique {
		w.WriteString("UNIQUE ")
	}

	if s.clustered != nil {
		if *s.clustered {
			w.WriteString("CLUSTERED ")
		} else {
			w.WriteString("NONCLUSTERED ")
		}
	}

	w.WriteString("INDEX ")

	if s.concurrently {
		w.WriteString("CONCURRENTLY ")
	}

	if s.ifNotExists {
		w.WriteString("IF NOT EXISTS ")
	}

	w.WriteString(d.QuoteIdent(s.name))
	w.WriteString(" ON ")
	w.WriteString(d.QuoteIdent(s.table))

	if s.method != IndexDefault {
		w.WriteString(" USING ")
		w.WriteString(string(s.method))
	}

	w.WriteString(" (")
	for i, key := range s.keys {
		if i > 0 {
			w.WriteString(", ")
		}
		key.Render(w, d)
	}
	w.WriteString(")")

	if len(s.include) > 0 {
		w.WriteString(" INCLUDE (")
		for i, expr := range s.include {
			if i > 0 {
				w.WriteString(", ")
			}
			expr.Render(w, d)
		}
		w.WriteString(")")
	}

	if s.nullsNotDistinct {
		w.WriteString(" NULLS NOT DISTINCT")
	}

	if len(s.options) > 0 {
		w.WriteString(" WITH (")
		for i, opt := range s.options {
			if i > 0 {
				w.WriteString(", ")
			}
			w.WriteString(opt.name)
			w.WriteString(" = ")
			w.WriteString(renderIndexOptionValue(d, opt.value))
		}
		w.WriteString(")")
	}

	if s.tablespace != "" {
		w.WriteString(" TABLESPACE ")
		w.WriteString(d.QuoteIdent(s.tablespace))
	}

	if s.pred != nil {
		if isMySQL(d) {
			unsupported(d, "PARTIAL INDEX")
		}
		w.WriteString(" WHERE ")
		s.pred.Render(w, d)
	}

	if s.invisible {
		w.WriteString(" INVISIBLE")
	}
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
