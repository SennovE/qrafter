package migrations

import (
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

func normalizeSQL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func normalizeReferenceAction(action ddl.ReferenceAction) ddl.ReferenceAction {
	if action == "" {
		return ddl.NoAction
	}
	return action
}
