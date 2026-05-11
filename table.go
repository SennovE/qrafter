package qrafter

import "github.com/SennovE/qrafter/internal/core"

type TableConfigProvider interface {
	TableConfig() TableConfig
}

type TableConfig struct {
	Name string
}

func TableAlias[T TableConfigProvider](table T, alias string) (T, error) {
	config := table.TableConfig()
	err := bindWithTableRef(&table, core.TableRef{Name: config.Name, Alias: alias})
	return table, err
}
