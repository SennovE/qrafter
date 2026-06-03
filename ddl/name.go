package ddl

import (
	"fmt"
	"strings"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

func tableName(table any) string {
	switch v := table.(type) {
	case string:
		return requireName("table", v)
	case q.TableConfigProvider:
		ref := q.GetTableRef(v)
		return requireName("table", ref.Name)
	default:
		panic(fmt.Errorf("unsupported table identifier %T", table))
	}
}

func requireName(kind, name string) string {
	if name == "" {
		panic(fmt.Errorf("%s name is empty", kind))
	}
	return name
}

func renderColumnList(w *strings.Builder, d dialect.Renderer, columns []string) {
	for i, column := range columns {
		if i > 0 {
			w.WriteString(", ")
		}
		w.WriteString(d.QuoteIdent(column))
	}
}
