package ddl

import (
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/dialect"
)

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

type IndexKey struct {
	expr Expression

	sort      sortDirection
	nulls     nullsOrder
	collation string
	opclass   string
	length    int
}

func Key(expr Expression) IndexKey {
	return IndexKey{expr: expr}
}

func KeyCol(name string) IndexKey {
	return IndexKey{expr: Col(name)}
}

func (k IndexKey) Asc() IndexKey {
	k.sort = sortAsc
	return k
}

func (k IndexKey) Desc() IndexKey {
	k.sort = sortDesc
	return k
}

func (k IndexKey) NullsFirst() IndexKey {
	k.nulls = nullsFirst
	return k
}

func (k IndexKey) NullsLast() IndexKey {
	k.nulls = nullsLast
	return k
}

func (k IndexKey) Collate(name string) IndexKey {
	return k
}

func (k IndexKey) OpClass(name string) IndexKey {
	return k
}

func (k IndexKey) PrefixLength(n int) IndexKey {
	if n <= 0 {
		panic("ddl: index prefix length must be positive")
	}
	k.length = n
	return k
}

func (k IndexKey) Render(w *strings.Builder, d dialect.Renderer) {
	k.expr.Render(w, d)

	if k.length > 0 {
		w.WriteString("(")
		w.WriteString(strconv.Itoa(k.length))
		w.WriteString(")")
	}

	if k.collation != "" {
		w.WriteString(" COLLATE ")
		w.WriteString(d.QuoteIdent(k.collation))
	}

	if k.opclass != "" {
		w.WriteString(" ")
		w.WriteString(k.opclass)
	}

	if k.sort != "" {
		w.WriteString(" ")
		w.WriteString(string(k.sort))
	}

	if k.nulls != "" {
		w.WriteString(" NULLS ")
		w.WriteString(string(k.nulls))
	}
}
