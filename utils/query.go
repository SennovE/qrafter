package utils

import "fmt"

func QuoteIdent(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}
