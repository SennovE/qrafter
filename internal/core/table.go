package core

import (
	"sort"

	"github.com/SennovE/qrafter/internal/utils"
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

func (t TableRef) BuildSQL() string {
	if t.Alias == "" {
		return utils.QuoteIdent(t.Name)
	}
	return utils.QuoteIdent(t.Name) + " AS " + utils.QuoteIdent(t.Alias)
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
