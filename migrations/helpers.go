package migrations

import "strings"

func appendConstraintsUnique(dst []Constraint, constraints ...Constraint) []Constraint {
	seen := make(map[string]int, len(dst))

	for i := range dst {
		seen[constraintKey(dst[i])] = i
	}

	for _, c := range constraints {
		key := constraintKey(c)

		if existingIdx, ok := seen[key]; ok {
			dst[existingIdx] = mergeConstraint(dst[existingIdx], c)
			continue
		}

		seen[key] = len(dst)
		dst = append(dst, c)
	}

	return dst
}

func mergeConstraint(existing, incoming Constraint) Constraint {
	if existing.Name == "" {
		existing.Name = incoming.Name
		return existing
	}

	if incoming.Name == "" || existing.Name == incoming.Name {
		return existing
	}

	panic("migrations: duplicate constraint with different names: " +
		existing.Name + " and " + incoming.Name)
}

func constraintKey(c Constraint) string {
	var b strings.Builder

	b.WriteString(string(c.Kind))
	b.WriteString("|cols=")
	b.WriteString(strings.Join(c.Columns, ","))

	switch c.Kind {
	case ConstraintCheck:
		b.WriteString("|check=")
		b.WriteString(normalizeSQL(c.CheckExpr))

	case ConstraintForeignKey:
		b.WriteString("|ref=")
		b.WriteString(c.Reference.Schema)
		b.WriteByte('.')
		b.WriteString(c.Reference.TableName)
		b.WriteByte('(')
		b.WriteString(strings.Join(c.Reference.Columns, ","))
		b.WriteByte(')')
		b.WriteString("|del=")
		b.WriteString(string(c.OnDelete))
		b.WriteString("|upd=")
		b.WriteString(string(c.OnUpdate))
	}

	return b.String()
}

func normalizeSQL(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}
