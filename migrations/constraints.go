package migrations

func appendConstraintsUnique(dst []Constraint, constraints ...Constraint) []Constraint {
	seen := make(map[string]int, len(dst))

	for i := range dst {
		seen[constraintKey(&dst[i])] = i
	}

	for i := range constraints {
		key := constraintKey(&constraints[i])

		if existingIdx, ok := seen[key]; ok {
			mergeConstraint(&dst[existingIdx], &constraints[i])
			continue
		}

		seen[key] = len(dst)
		dst = append(dst, constraints[i])
	}

	return dst
}

func mergeConstraint(existing, incoming *Constraint) {
	if existing.Schema == "" {
		existing.Schema = incoming.Schema
	}
	if existing.TableName == "" {
		existing.TableName = incoming.TableName
	}

	if existing.Name == "" {
		existing.Name = incoming.Name
		return
	}

	if incoming.Name == "" || existing.Name == incoming.Name {
		return
	}

	panic("migrations: duplicate constraint with different names: " +
		existing.Name + " and " + incoming.Name)
}

func qualifyTableConstraints(table *Table) {
	for i := range table.Constraints {
		if table.Constraints[i].Schema == "" {
			table.Constraints[i].Schema = table.Schema
		}
		if table.Constraints[i].TableName == "" {
			table.Constraints[i].TableName = table.Name
		}
	}
}
