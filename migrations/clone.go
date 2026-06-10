package migrations

func cloneNormalizedSchema(s Schema) Schema {
	out := Schema{Tables: make([]Table, len(s.Tables))}
	for i := range s.Tables {
		out.Tables[i] = cloneTable(&s.Tables[i])
	}
	out.normalize()
	return out
}

func cloneTable(t *Table) Table {
	out := *t
	out.Columns = append([]Column(nil), t.Columns...)

	out.Constraints = make([]Constraint, len(t.Constraints))
	for i := range t.Constraints {
		out.Constraints[i] = cloneConstraint(&t.Constraints[i])
	}

	out.Indexes = make([]Index, len(t.Indexes))
	for i := range t.Indexes {
		out.Indexes[i] = cloneIndex(&t.Indexes[i])
	}
	return out
}

func cloneConstraint(c *Constraint) Constraint {
	out := *c
	out.Columns = append([]string(nil), c.Columns...)
	out.Reference.Columns = append([]string(nil), c.Reference.Columns...)
	return out
}

func cloneIndex(i *Index) Index {
	out := *i
	out.Keys = append([]IndexKey(nil), i.Keys...)
	out.Include = append([]string(nil), i.Include...)
	return out
}
