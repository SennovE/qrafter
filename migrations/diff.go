package migrations

import "github.com/SennovE/qrafter/ddl"

// SchemaDiff describes changes needed to move Current database schema to
// Desired user schema.
type SchemaDiff struct {
	AddedTables   []Table
	RemovedTables []Table
	ChangedTables []TableDiff
}

// IsEmpty reports whether the schema snapshots are equal for diff purposes.
func (d SchemaDiff) IsEmpty() bool {
	return len(d.AddedTables) == 0 &&
		len(d.RemovedTables) == 0 &&
		len(d.ChangedTables) == 0
}

// TableDiff describes changes inside one table.
type TableDiff struct {
	Schema string
	Name   string

	AddedColumns   []Column
	RemovedColumns []Column
	ChangedColumns []ColumnDiff

	AddedConstraints   []Constraint
	RemovedConstraints []Constraint
	ChangedConstraints []ConstraintDiff

	AddedIndexes   []Index
	RemovedIndexes []Index
	ChangedIndexes []IndexDiff
}

// IsEmpty reports whether the table snapshots are equal for diff purposes.
func (d *TableDiff) IsEmpty() bool {
	return len(d.AddedColumns) == 0 &&
		len(d.RemovedColumns) == 0 &&
		len(d.ChangedColumns) == 0 &&
		len(d.AddedConstraints) == 0 &&
		len(d.RemovedConstraints) == 0 &&
		len(d.ChangedConstraints) == 0 &&
		len(d.AddedIndexes) == 0 &&
		len(d.RemovedIndexes) == 0 &&
		len(d.ChangedIndexes) == 0
}

// ColumnDiff stores the current and desired versions of a changed column.
type ColumnDiff struct {
	Current Column
	Desired Column
}

// ConstraintDiff stores the current and desired versions of a changed
// constraint.
type ConstraintDiff struct {
	Current Constraint
	Desired Constraint
}

// IndexDiff stores the current and desired versions of a changed index.
type IndexDiff struct {
	Current Index
	Desired Index
}

// Diff compares the receiver as the current schema against desired.
func (s Schema) Diff(desired Schema) SchemaDiff {
	return DiffSchemas(s, desired)
}

// DiffSchemas compares current database schema against desired user schema.
func DiffSchemas(current, desired Schema) SchemaDiff {
	current = cloneNormalizedSchema(current)
	desired = cloneNormalizedSchema(desired)

	currentTables := tablesByKey(current.Tables)
	desiredTables := tablesByKey(desired.Tables)

	diff := SchemaDiff{}
	for i := range desired.Tables {
		table := &desired.Tables[i]
		key := tableKey{schema: table.Schema, table: table.Name}
		currentTable, ok := currentTables[key]
		if !ok {
			diff.AddedTables = append(diff.AddedTables, *table)
			continue
		}

		tableDiff := diffTables(currentTable, table)
		if !tableDiff.IsEmpty() {
			diff.ChangedTables = append(diff.ChangedTables, tableDiff)
		}
	}
	for i := range current.Tables {
		table := &current.Tables[i]
		key := tableKey{schema: table.Schema, table: table.Name}
		if _, ok := desiredTables[key]; !ok {
			diff.RemovedTables = append(diff.RemovedTables, *table)
		}
	}
	return diff
}

func diffTables(current, desired *Table) TableDiff {
	diff := TableDiff{
		Schema: desired.Schema,
		Name:   desired.Name,
	}
	diff.AddedColumns, diff.RemovedColumns, diff.ChangedColumns = diffColumns(current.Columns, desired.Columns)
	diff.AddedConstraints, diff.RemovedConstraints, diff.ChangedConstraints = diffConstraints(
		current.Constraints,
		desired.Constraints,
	)
	diff.AddedIndexes, diff.RemovedIndexes, diff.ChangedIndexes = diffIndexes(current.Indexes, desired.Indexes)
	return diff
}

func diffColumns(current, desired []Column) (added, removed []Column, changed []ColumnDiff) {
	currentByName := make(map[string]int, len(current))
	desiredByName := make(map[string]int, len(desired))
	for i := range current {
		currentByName[current[i].Name] = i
	}
	for i := range desired {
		desiredByName[desired[i].Name] = i
	}

	for i := range desired {
		currentIdx, ok := currentByName[desired[i].Name]
		if !ok {
			added = append(added, desired[i])
			continue
		}
		if !columnsEqual(&current[currentIdx], &desired[i]) {
			changed = append(changed, ColumnDiff{Current: current[currentIdx], Desired: desired[i]})
		}
	}

	for i := range current {
		if _, ok := desiredByName[current[i].Name]; !ok {
			removed = append(removed, current[i])
		}
	}
	return added, removed, changed
}

func diffConstraints(
	current, desired []Constraint,
) (added, removed []Constraint, changed []ConstraintDiff) {
	currentUsed := make([]bool, len(current))
	desiredUsed := make([]bool, len(desired))

	changed = matchConstraintsBySemantic(current, desired, currentUsed, desiredUsed, changed)
	changed = matchConstraintsByName(current, desired, currentUsed, desiredUsed, changed)

	added = unmatchedDesiredConstraints(desired, desiredUsed)
	removed = unmatchedCurrentConstraints(current, currentUsed)
	return added, removed, changed
}

func matchConstraintsBySemantic(
	current, desired []Constraint,
	currentUsed, desiredUsed []bool,
	changed []ConstraintDiff,
) []ConstraintDiff {
	semantic := make(map[string]int, len(current))
	for i := range current {
		semantic[constraintSemanticKey(&current[i])] = i
	}

	for i := range desired {
		j, ok := semantic[constraintSemanticKey(&desired[i])]
		if !ok || currentUsed[j] {
			continue
		}
		currentUsed[j] = true
		desiredUsed[i] = true
		if constraintChanged(&current[j], &desired[i]) {
			changed = append(changed, ConstraintDiff{Current: current[j], Desired: desired[i]})
		}
	}
	return changed
}

func matchConstraintsByName(
	current, desired []Constraint,
	currentUsed, desiredUsed []bool,
	changed []ConstraintDiff,
) []ConstraintDiff {
	byName := make(map[string]int, len(current))
	for i := range current {
		if !currentUsed[i] && current[i].Name != "" {
			byName[current[i].Name] = i
		}
	}
	for i := range desired {
		if desiredUsed[i] || desired[i].Name == "" {
			continue
		}
		j, ok := byName[desired[i].Name]
		if !ok || currentUsed[j] {
			continue
		}
		currentUsed[j] = true
		desiredUsed[i] = true
		if constraintChanged(&current[j], &desired[i]) {
			changed = append(changed, ConstraintDiff{Current: current[j], Desired: desired[i]})
		}
	}
	return changed
}

func unmatchedDesiredConstraints(desired []Constraint, desiredUsed []bool) []Constraint {
	var added []Constraint
	for i := range desired {
		if !desiredUsed[i] {
			added = append(added, desired[i])
		}
	}
	return added
}

func unmatchedCurrentConstraints(current []Constraint, currentUsed []bool) []Constraint {
	var removed []Constraint
	for i := range current {
		if !currentUsed[i] {
			removed = append(removed, current[i])
		}
	}
	return removed
}

func diffIndexes(current, desired []Index) (added, removed []Index, changed []IndexDiff) {
	currentByName := make(map[string]int, len(current))
	desiredByName := make(map[string]int, len(desired))
	for i := range current {
		currentByName[current[i].Name] = i
	}
	for i := range desired {
		desiredByName[desired[i].Name] = i
	}

	for i := range desired {
		currentIdx, ok := currentByName[desired[i].Name]
		if !ok {
			added = append(added, desired[i])
			continue
		}
		if !indexesEqual(&current[currentIdx], &desired[i]) {
			changed = append(changed, IndexDiff{Current: current[currentIdx], Desired: desired[i]})
		}
	}

	for i := range current {
		if _, ok := desiredByName[current[i].Name]; !ok {
			removed = append(removed, current[i])
		}
	}
	return added, removed, changed
}

func columnsEqual(current, desired *Column) bool {
	return typesEqual(current.ddlType(), desired.ddlType()) &&
		current.NotNull == desired.NotNull &&
		current.HasDefault == desired.HasDefault &&
		normalizeSQL(current.DefaultExpr) == normalizeSQL(desired.DefaultExpr) &&
		current.Identity == desired.Identity &&
		current.Generated == desired.Generated &&
		normalizeSQL(current.GeneratedExpr) == normalizeSQL(desired.GeneratedExpr)
}

func typesEqual(current, desired ddl.Type) bool {
	if normalizeSQL(current.Name) != normalizeSQL(desired.Name) {
		return false
	}
	if len(current.DialectNames) != len(desired.DialectNames) {
		return false
	}
	for dialectName, currentName := range current.DialectNames {
		desiredName, ok := desired.DialectNames[dialectName]
		if !ok || normalizeSQL(currentName) != normalizeSQL(desiredName) {
			return false
		}
	}
	return true
}

func constraintChanged(current, desired *Constraint) bool {
	if constraintSemanticKey(current) != constraintSemanticKey(desired) {
		return true
	}
	return desired.Name != "" && current.Name != desired.Name
}

func constraintSemanticKey(c *Constraint) string {
	normalized := *c
	normalized.OnDelete = normalizeReferenceAction(normalized.OnDelete)
	normalized.OnUpdate = normalizeReferenceAction(normalized.OnUpdate)
	return constraintKey(&normalized)
}

func indexesEqual(current, desired *Index) bool {
	return current.Name == desired.Name &&
		current.TableName == desired.TableName &&
		current.Unique == desired.Unique &&
		current.Method == desired.Method &&
		current.Tablespace == desired.Tablespace &&
		current.NullsNotDistinct == desired.NullsNotDistinct &&
		normalizeSQL(current.Predicate) == normalizeSQL(desired.Predicate) &&
		stringSlicesEqualNormalized(indexKeyExpressions(current.Keys), indexKeyExpressions(desired.Keys)) &&
		stringSlicesEqualNormalized(current.Include, desired.Include)
}

func indexKeyExpressions(keys []IndexKey) []string {
	out := make([]string, len(keys))
	for i := range keys {
		out[i] = keys[i].Expression
	}
	return out
}

func stringSlicesEqualNormalized(current, desired []string) bool {
	if len(current) != len(desired) {
		return false
	}
	for i := range current {
		if normalizeSQL(current[i]) != normalizeSQL(desired[i]) {
			return false
		}
	}
	return true
}
