package core

import (
	"sort"

	"github.com/SennovE/qrafter/dialect"
)

type ColumnBinder interface {
	Bind(name string, table TableRef)
}

type TableRef struct {
	Name  string
	Alias string
}

type TablesSet = map[TableRef]struct{}

func (t TableRef) SQLName() string {
	if t.Alias == "" {
		return t.Name
	}
	return t.Alias
}

func (t TableRef) Render(d dialect.DialectRenderer) string {
	if t.Alias == "" {
		return d.QuoteIdent(t.Name)
	}
	return d.QuoteIdent(t.Name) + " AS " + d.QuoteIdent(t.Alias)
}

func GetSortedTables(tables TablesSet) []TableRef {
	sortedTables := make([]TableRef, 0, len(tables))
	for table := range tables {
		sortedTables = append(sortedTables, table)
	}
	sort.Slice(sortedTables, func(i, j int) bool {
		if sortedTables[i].Name == sortedTables[j].Name {
			return sortedTables[i].Alias < sortedTables[j].Alias
		}
		return sortedTables[i].Name < sortedTables[j].Name
	})
	return sortedTables
}
