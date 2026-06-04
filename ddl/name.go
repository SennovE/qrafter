package ddl

import (
	"fmt"

	q "github.com/SennovE/qrafter"
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
