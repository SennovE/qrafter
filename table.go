package qrafter

import (
	"github.com/SennovE/qrafter/utils"
)

type TableConfig struct {
	Name string
}

type TableConfigProvider interface {
	TableConfig() TableConfig
}

type TableRef struct {
	Name  string
	Alias string
}

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

func TableAlias[T TableConfigProvider](table T, alias string) (T, error) {
	config := table.TableConfig()
	err := bindWithTableRef(&table, TableRef{Name: config.Name, Alias: alias})
	return table, err
}
