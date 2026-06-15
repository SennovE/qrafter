package migrations

import "github.com/SennovE/qrafter/ddl"

type migrationStep struct {
	up       ddl.Renderer
	down     ddl.Renderer
	upCode   string
	downCode string
}

func migrationSteps(diff SchemaDiff) []migrationStep {
	var steps []migrationStep
	steps = appendAddedTableSteps(steps, diff.AddedTables)
	steps = appendRemovedTableSteps(steps, diff.RemovedTables)
	for i := range diff.ChangedTables {
		steps = appendTableDiffSteps(steps, &diff.ChangedTables[i])
	}
	return steps
}

func appendAddedTableSteps(steps []migrationStep, tables []Table) []migrationStep {
	for i := range tables {
		table := &tables[i]
		steps = append(steps, createTableStep(table))
		for j := range table.Indexes {
			steps = append(steps, createIndexStep(&table.Indexes[j]))
		}
	}
	return steps
}

func appendRemovedTableSteps(steps []migrationStep, tables []Table) []migrationStep {
	for i := range tables {
		table := &tables[i]
		for j := range table.Indexes {
			steps = append(steps, restoreIndexStep(&table.Indexes[j]))
		}
		steps = append(steps, dropTableStep(table))
	}
	return steps
}

func appendTableDiffSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	steps = appendDroppedIndexSteps(steps, diff)
	steps = appendDroppedConstraintSteps(steps, diff)
	steps = appendRemovedColumnSteps(steps, diff)
	steps = appendAddedColumnSteps(steps, diff)
	steps = appendChangedColumnSteps(steps, diff)
	steps = appendAddedConstraintSteps(steps, diff)
	steps = appendCreatedIndexSteps(steps, diff)
	return steps
}

func appendDroppedIndexSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.RemovedIndexes {
		steps = append(steps, dropIndexStep(&diff.RemovedIndexes[i]))
	}
	for i := range diff.ChangedIndexes {
		steps = append(steps, dropIndexStep(&diff.ChangedIndexes[i].Current))
	}
	return steps
}

func appendDroppedConstraintSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.RemovedConstraints {
		steps = append(steps, dropConstraintStep(diff.Name, &diff.RemovedConstraints[i]))
	}
	for i := range diff.ChangedConstraints {
		steps = append(steps, dropConstraintStep(diff.Name, &diff.ChangedConstraints[i].Current))
	}
	return steps
}

func appendRemovedColumnSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.RemovedColumns {
		steps = append(steps, dropColumnStep(diff.Name, &diff.RemovedColumns[i]))
	}
	return steps
}

func appendAddedColumnSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.AddedColumns {
		steps = append(steps, addColumnStep(diff.Name, &diff.AddedColumns[i]))
	}
	return steps
}

func appendChangedColumnSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.ChangedColumns {
		steps = appendColumnChangeSteps(steps, diff.Name, &diff.ChangedColumns[i])
	}
	return steps
}

func appendAddedConstraintSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.ChangedConstraints {
		steps = append(steps, addConstraintStep(diff.Name, &diff.ChangedConstraints[i].Desired))
	}
	for i := range diff.AddedConstraints {
		steps = append(steps, addConstraintStep(diff.Name, &diff.AddedConstraints[i]))
	}
	return steps
}

func appendCreatedIndexSteps(steps []migrationStep, diff *TableDiff) []migrationStep {
	for i := range diff.ChangedIndexes {
		steps = append(steps, createIndexStep(&diff.ChangedIndexes[i].Desired))
	}
	for i := range diff.AddedIndexes {
		steps = append(steps, createIndexStep(&diff.AddedIndexes[i]))
	}
	return steps
}

func appendColumnChangeSteps(steps []migrationStep, table string, diff *ColumnDiff) []migrationStep {
	if columnDefinitionModeChanged(&diff.Current, &diff.Desired) {
		return append(steps, replaceColumnStep(table, diff))
	}
	if !typesEqual(diff.Current.ddlType(), diff.Desired.ddlType()) {
		steps = append(steps, alterColumnTypeStep(table, diff))
	}
	if diff.Current.NotNull != diff.Desired.NotNull {
		steps = append(steps, changeNotNullStep(table, diff))
	}
	if diff.Current.HasDefault != diff.Desired.HasDefault ||
		normalizeSQL(diff.Current.DefaultExpr) != normalizeSQL(diff.Desired.DefaultExpr) {
		steps = append(steps, changeDefaultStep(table, diff))
	}
	return steps
}

func columnDefinitionModeChanged(current, desired *Column) bool {
	return current.Identity != desired.Identity ||
		current.Generated != desired.Generated ||
		normalizeSQL(current.GeneratedExpr) != normalizeSQL(desired.GeneratedExpr)
}

func createTableStep(table *Table) migrationStep {
	return migrationStep{
		up:       table.CreateTable(),
		down:     ddl.DropTable(table.Name),
		upCode:   createTableCode(table),
		downCode: dropTableCode(table.Name),
	}
}

func dropTableStep(table *Table) migrationStep {
	return migrationStep{
		up:       ddl.DropTable(table.Name),
		down:     table.CreateTable(),
		upCode:   dropTableCode(table.Name),
		downCode: createTableCode(table),
	}
}

func createIndexStep(index *Index) migrationStep {
	return migrationStep{
		up:       index.CreateIndex(),
		down:     ddl.DropIndex(index.Name),
		upCode:   createIndexCode(index),
		downCode: dropIndexCode(index.Name),
	}
}

func restoreIndexStep(index *Index) migrationStep {
	return migrationStep{
		down:     index.CreateIndex(),
		downCode: createIndexCode(index),
	}
}

func dropIndexStep(index *Index) migrationStep {
	return migrationStep{
		up:       ddl.DropIndex(index.Name),
		down:     index.CreateIndex(),
		upCode:   dropIndexCode(index.Name),
		downCode: createIndexCode(index),
	}
}

func addColumnStep(table string, column *Column) migrationStep {
	return migrationStep{
		up:       ddl.AlterTable(table).AddColumn(column.ColumnDef()),
		down:     ddl.AlterTable(table).DropColumn(column.Name),
		upCode:   addColumnCode(table, column),
		downCode: dropColumnCode(table, column.Name),
	}
}

func dropColumnStep(table string, column *Column) migrationStep {
	return migrationStep{
		up:       ddl.AlterTable(table).DropColumn(column.Name),
		down:     ddl.AlterTable(table).AddColumn(column.ColumnDef()),
		upCode:   dropColumnCode(table, column.Name),
		downCode: addColumnCode(table, column),
	}
}

func replaceColumnStep(table string, diff *ColumnDiff) migrationStep {
	return migrationStep{
		up: ddl.AlterTable(table).
			DropColumn(diff.Current.Name).
			AddColumn(diff.Desired.ColumnDef()),
		down: ddl.AlterTable(table).
			DropColumn(diff.Desired.Name).
			AddColumn(diff.Current.ColumnDef()),
		upCode:   replaceColumnCode(table, diff.Current.Name, &diff.Desired),
		downCode: replaceColumnCode(table, diff.Desired.Name, &diff.Current),
	}
}

func alterColumnTypeStep(table string, diff *ColumnDiff) migrationStep {
	return migrationStep{
		up:       ddl.AlterTable(table).AlterColumnType(diff.Desired.Name, diff.Desired.ddlType()),
		down:     ddl.AlterTable(table).AlterColumnType(diff.Current.Name, diff.Current.ddlType()),
		upCode:   alterColumnTypeCode(table, &diff.Desired),
		downCode: alterColumnTypeCode(table, &diff.Current),
	}
}

func changeNotNullStep(table string, diff *ColumnDiff) migrationStep {
	return migrationStep{
		up:       notNullStatement(table, &diff.Desired),
		down:     notNullStatement(table, &diff.Current),
		upCode:   notNullCode(table, &diff.Desired),
		downCode: notNullCode(table, &diff.Current),
	}
}

func changeDefaultStep(table string, diff *ColumnDiff) migrationStep {
	return migrationStep{
		up:       defaultStatement(table, &diff.Desired),
		down:     defaultStatement(table, &diff.Current),
		upCode:   defaultCode(table, &diff.Desired),
		downCode: defaultCode(table, &diff.Current),
	}
}

func addConstraintStep(table string, constraint *Constraint) migrationStep {
	name := constraintDropName(table, constraint)
	return migrationStep{
		up:       ddl.AlterTable(table).AddConstraint(constraint.TableConstraint()),
		down:     ddl.AlterTable(table).DropConstraint(name),
		upCode:   addConstraintCode(table, constraint),
		downCode: dropConstraintCode(table, name),
	}
}

func dropConstraintStep(table string, constraint *Constraint) migrationStep {
	name := constraintDropName(table, constraint)
	return migrationStep{
		up:       ddl.AlterTable(table).DropConstraint(name),
		down:     ddl.AlterTable(table).AddConstraint(constraint.TableConstraint()),
		upCode:   dropConstraintCode(table, name),
		downCode: addConstraintCode(table, constraint),
	}
}

func notNullStatement(table string, column *Column) ddl.AlterTableStmt {
	stmt := ddl.AlterTable(table)
	if column.NotNull {
		return stmt.SetNotNull(column.Name)
	}
	return stmt.DropNotNull(column.Name)
}

func defaultStatement(table string, column *Column) ddl.AlterTableStmt {
	stmt := ddl.AlterTable(table)
	if column.HasDefault {
		return stmt.SetDefaultExpr(column.Name, column.DefaultExpr)
	}
	return stmt.DropDefault(column.Name)
}
