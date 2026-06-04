package core

import (
	"sort"
)

type ColumnBinder interface {
	Bind(name string, table TableRef)
}

type TableRef struct {
	Name  string
	Alias string
	CTE   *CTERef
}

type TablesSet = map[TableRef]struct{}

type CTERef struct {
	Name      string
	Columns   []string
	Query     QueryExpression
	Recursive bool
}

func (t TableRef) SQLName() string {
	if t.Alias == "" {
		return t.Name
	}
	return t.Alias
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
