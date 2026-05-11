package expr

import (
	"fmt"
	"strings"

	"github.com/SennovE/qrafter/internal/core"
)

type ConstExpression struct {
	v string
}

var _ = (core.Selecter)(ConstExpression{})

func (c ConstExpression) Tables() core.TablesSet {
	return nil
}

func (c ConstExpression) Render() string {
	return c.v
}

func Const(value any) ConstExpression {
	var c ConstExpression

	switch v := value.(type) {
	case nil:
		c.v = "NULL"
	case bool:
		if v {
			c.v = "TRUE"
		} else {
			c.v = "FALSE"
		}
	case string:
		c.v = escapeSQLstring(v)
	default:
		c.v = fmt.Sprint(v)
	}

	return c
}

func escapeSQLstring(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
