package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/format"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

const (
	defaultMigrationPackageName  = "migrations"
	defaultMigrationUpFuncName   = "Up"
	defaultMigrationDownFuncName = "Down"
	ddlCodeQualifier             = "qddl."
	ddlImportPath                = "github.com/SennovE/qrafter/ddl"
)

// MigrationCodeOptions configures generated Go migration code.
type MigrationCodeOptions struct {
	PackageName  string
	UpFuncName   string
	DownFuncName string
}

// GenerateMigrationCodeFromDiff generates Go code with ddl-based Up and Down
// functions for the given schema diff.
func GenerateMigrationCodeFromDiff(diff SchemaDiff, options MigrationCodeOptions) ([]byte, error) {
	options = defaultMigrationCodeOptions(options)
	if err := validateMigrationCodeOptions(options); err != nil {
		return nil, err
	}

	var b strings.Builder
	steps := migrationSteps(diff)

	fmt.Fprintf(&b, "package %s\n\n", options.PackageName)
	fmt.Fprintf(&b, "import qddl %q\n\n", ddlImportPath)
	writeMigrationFunc(&b, options.UpFuncName, upStepCodes(steps))
	b.WriteString("\n")
	writeMigrationFunc(&b, options.DownFuncName, downStepCodes(steps))

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return nil, fmt.Errorf("format generated migration code: %w", err)
	}
	return src, nil
}

func defaultMigrationCodeOptions(options MigrationCodeOptions) MigrationCodeOptions {
	if options.PackageName == "" {
		options.PackageName = defaultMigrationPackageName
	}
	if options.UpFuncName == "" {
		options.UpFuncName = defaultMigrationUpFuncName
	}
	if options.DownFuncName == "" {
		options.DownFuncName = defaultMigrationDownFuncName
	}
	return options
}

func validateMigrationCodeOptions(options MigrationCodeOptions) error {
	if !validGoIdentifier(options.PackageName) {
		return fmt.Errorf("migrations: invalid package name %q", options.PackageName)
	}
	if !validGoIdentifier(options.UpFuncName) {
		return fmt.Errorf("migrations: invalid up function name %q", options.UpFuncName)
	}
	if !validGoIdentifier(options.DownFuncName) {
		return fmt.Errorf("migrations: invalid down function name %q", options.DownFuncName)
	}
	if options.UpFuncName == options.DownFuncName {
		return fmt.Errorf("migrations: up and down function names must differ")
	}
	return nil
}

func validGoIdentifier(name string) bool {
	return name != "_" && token.IsIdentifier(name) && !token.Lookup(name).IsKeyword()
}

func writeMigrationFunc(b *strings.Builder, name string, statements []string) {
	fmt.Fprintf(b, "func %s() %sStatements {\n", name, ddlCodeQualifier)
	if len(statements) == 0 {
		b.WriteString("\treturn nil\n")
		b.WriteString("}\n")
		return
	}

	fmt.Fprintf(b, "\treturn %sStatements{\n", ddlCodeQualifier)
	for i := range statements {
		b.WriteString("\t\t")
		b.WriteString(statements[i])
		b.WriteString(",\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
}

func upStepCodes(steps []migrationStep) []string {
	codes := make([]string, 0, len(steps))
	for i := range steps {
		if steps[i].upCode != "" {
			codes = append(codes, steps[i].upCode)
		}
	}
	return codes
}

func downStepCodes(steps []migrationStep) []string {
	codes := make([]string, 0, len(steps))
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].downCode != "" {
			codes = append(codes, steps[i].downCode)
		}
	}
	return codes
}

func createTableCode(table *Table) string {
	var code strings.Builder
	code.WriteString(ddlCodeQualifier)
	code.WriteString("CreateTable(")
	code.WriteString(quoteCode(table.Name))
	code.WriteString(")")
	if len(table.Columns) > 0 {
		code.WriteString(".Columns(")
		joinColumnCodes(&code, table.Columns)
		code.WriteString(")")
	}
	if len(table.Constraints) > 0 {
		code.WriteString(".Constraints(")
		joinConstraintCodes(&code, table.Name, table.Constraints)
		code.WriteString(")")
	}
	return code.String()
}

func joinColumnCodes(w *strings.Builder, columns []Column) {
	for _, col := range columns {
		columnCode(w, &col)
		w.WriteString(", ")
	}
}

func joinConstraintCodes(w *strings.Builder, table string, constraints []Constraint) {
	for _, constr := range constraints {
		constraintCode(w, table, &constr)
		w.WriteString(", ")
	}
}

func columnCode(w *strings.Builder, column *Column) {
	w.WriteString(ddlCodeQualifier)
	w.WriteString("Column(")
	w.WriteString(quoteCode(column.Name))
	w.WriteString(", ")
	w.WriteString(typeCode(column.ddlType()))
	w.WriteString(")")
	switch column.Identity {
	case IdentityAlways:
		w.WriteString(".IdentityAlways()")
	case IdentityByDefault:
		w.WriteString(".IdentityByDefault()")
	}
	switch column.Generated {
	case GeneratedStored:
		w.WriteString(".GeneratedStored(")
		w.WriteString(quoteCode(column.GeneratedExpr))
		w.WriteString(")")
	case GeneratedVirtual:
		w.WriteString(".GeneratedVirtual(")
		w.WriteString(quoteCode(column.GeneratedExpr))
		w.WriteString(")")
	}
	if column.NotNull {
		w.WriteString(".NotNull()")
	}
	if column.HasDefault && column.Identity == IdentityNone && column.Generated == GeneratedNone {
		w.WriteString(".DefaultExpr(")
		w.WriteString(quoteCode(column.DefaultExpr))
		w.WriteString(")")
	}
}

func typeCode(typ ddl.Type) string {
	code := ddlCodeQualifier + "SQLType(" + quoteCode(typ.Name) + ")"
	keys := make([]string, 0, len(typ.DialectNames))
	for dialectName := range typ.DialectNames {
		keys = append(keys, dialectName)
	}
	sort.Strings(keys)
	for _, dialectName := range keys {
		code += ".ForDialect(" + quoteCode(dialectName) + ", " + quoteCode(typ.DialectNames[dialectName]) + ")"
	}
	return code
}

func constraintCode(w *strings.Builder, table string, constraint *Constraint) {
	switch constraint.Kind {
	case ConstraintPrimaryKey:
		w.WriteString(ddlCodeQualifier)
		w.WriteString("PrimaryKey(")
		w.WriteString(stringArgsCode(constraint.Columns))
		w.WriteString(")")
	case ConstraintUnique:
		w.WriteString(ddlCodeQualifier)
		w.WriteString("Unique(")
		w.WriteString(stringArgsCode(constraint.Columns))
		w.WriteString(")")
	case ConstraintCheck:
		w.WriteString(ddlCodeQualifier)
		w.WriteString("Check(")
		w.WriteString(ddlCodeQualifier)
		w.WriteString("RawPred(")
		w.WriteString(quoteCode(constraint.CheckExpr))
		w.WriteString("))")
	case ConstraintForeignKey:
		foreignKeyCode(w, constraint)
	default:
		panic("unsupported constraint " + string(constraint.Kind))
	}
	if constraint.Name == "" {
		return
	}
	w.WriteString(".Named(")
	w.WriteString(quoteCode(constraintDropName(table, constraint)))
	w.WriteString(")")
}

func foreignKeyCode(w *strings.Builder, constraint *Constraint) {
	w.WriteString(ddlCodeQualifier)
	w.WriteString("ForeignKey(")
	w.WriteString(stringArgsCode(constraint.Columns))
	w.WriteString(")")
	w.WriteString(".References(")
	w.WriteString(quoteCode(constraint.Reference.TableName))
	if len(constraint.Reference.Columns) > 0 {
		w.WriteString(", ")
		w.WriteString(stringArgsCode(constraint.Reference.Columns))
	}
	w.WriteString(")")
	if constraint.OnDelete != "" && constraint.OnDelete != ddl.NoAction {
		w.WriteString(".OnDelete(")
		w.WriteString(referenceActionCode(constraint.OnDelete))
		w.WriteString(")")
	}
	if constraint.OnUpdate != "" && constraint.OnUpdate != ddl.NoAction {
		w.WriteString(".OnUpdate(")
		w.WriteString(referenceActionCode(constraint.OnUpdate))
		w.WriteString(")")
	}
}

func referenceActionCode(action ddl.ReferenceAction) string {
	switch action {
	case ddl.Restrict:
		return ddlCodeQualifier + "Restrict"
	case ddl.Cascade:
		return ddlCodeQualifier + "Cascade"
	case ddl.SetNull:
		return ddlCodeQualifier + "SetNull"
	case ddl.SetDefault:
		return ddlCodeQualifier + "SetDefault"
	case ddl.NoAction:
		return ddlCodeQualifier + "NoAction"
	default:
		return ddlCodeQualifier + "ReferenceAction(" + quoteCode(string(action)) + ")"
	}
}

func createIndexCode(index *Index) string {
	var code strings.Builder
	code.WriteString(ddlCodeQualifier)
	code.WriteString("CreateIndex(")
	code.WriteString(quoteCode(index.Name))
	code.WriteString(").On(")
	code.WriteString(quoteCode(index.TableName))
	for i := range index.Keys {
		code.WriteString(", ")
		code.WriteString(ddlCodeQualifier)
		code.WriteString("Key(")
		code.WriteString(ddlCodeQualifier)
		code.WriteString("RawExpr(")
		code.WriteString(quoteCode(index.Keys[i].Expression))
		code.WriteString("))")
	}
	code.WriteString(")")

	if index.Unique {
		code.WriteString(".Unique()")
	}
	if index.Method != "" {
		code.WriteString(".Using(" + indexMethodCode(index.Method) + ")")
	}
	if len(index.Include) > 0 {
		code.WriteString(".Include(" + rawExpressionArgsCode(index.Include) + ")")
	}
	if index.NullsNotDistinct {
		code.WriteString(".NullsNotDistinct()")
	}
	if index.Tablespace != "" {
		code.WriteString(".Tablespace(" + quoteCode(index.Tablespace) + ")")
	}
	if index.Predicate != "" {
		code.WriteString(".Where(" + ddlCodeQualifier + "RawPred(" + quoteCode(index.Predicate) + "))")
	}
	return code.String()
}

func indexMethodCode(method ddl.IndexMethod) string {
	switch method {
	case ddl.IndexBTree:
		return ddlCodeQualifier + "IndexBTree"
	case ddl.IndexHash:
		return ddlCodeQualifier + "IndexHash"
	case ddl.IndexGin:
		return ddlCodeQualifier + "IndexGin"
	case ddl.IndexGist:
		return ddlCodeQualifier + "IndexGist"
	case ddl.IndexSpGist:
		return ddlCodeQualifier + "IndexSpGist"
	case ddl.IndexBrin:
		return ddlCodeQualifier + "IndexBrin"
	case ddl.IndexFullText:
		return ddlCodeQualifier + "IndexFullText"
	case ddl.IndexSpatial:
		return ddlCodeQualifier + "IndexSpatial"
	case ddl.IndexColumnstore:
		return ddlCodeQualifier + "IndexColumnstore"
	case ddl.IndexBitmap:
		return ddlCodeQualifier + "IndexBitmap"
	default:
		return ddlCodeQualifier + "IndexMethod(" + quoteCode(string(method)) + ")"
	}
}

func dropTableCode(table string) string {
	return ddlCodeQualifier + "DropTable(" + quoteCode(table) + ")"
}

func dropIndexCode(index string) string {
	return ddlCodeQualifier + "DropIndex(" + quoteCode(index) + ")"
}

func addColumnCode(table string, column *Column) string {
	var code strings.Builder
	code.WriteString(alterTableCode(table))
	code.WriteString(".AddColumn(")
	columnCode(&code, column)
	code.WriteString(")")
	return code.String()
}

func dropColumnCode(table, column string) string {
	return alterTableCode(table) + ".DropColumn(" + quoteCode(column) + ")"
}

func replaceColumnCode(table, dropColumn string, addColumn *Column) string {
	var code strings.Builder
	code.WriteString(alterTableCode(table))
	code.WriteString(".DropColumn(")
	code.WriteString(quoteCode(dropColumn))
	code.WriteString(")")
	code.WriteString(".AddColumn(")
	columnCode(&code, addColumn)
	code.WriteString(")")
	return code.String()
}

func alterColumnTypeCode(table string, column *Column) string {
	return alterTableCode(table) +
		".AlterColumnType(" + quoteCode(column.Name) + ", " + typeCode(column.ddlType()) + ")"
}

func notNullCode(table string, column *Column) string {
	if column.NotNull {
		return alterTableCode(table) + ".SetNotNull(" + quoteCode(column.Name) + ")"
	}
	return alterTableCode(table) + ".DropNotNull(" + quoteCode(column.Name) + ")"
}

func defaultCode(table string, column *Column) string {
	if column.HasDefault {
		return alterTableCode(table) +
			".SetDefaultExpr(" + quoteCode(column.Name) + ", " + quoteCode(column.DefaultExpr) + ")"
	}
	return alterTableCode(table) + ".DropDefault(" + quoteCode(column.Name) + ")"
}

func addConstraintCode(table string, constraint *Constraint) string {
	var code strings.Builder
	code.WriteString(alterTableCode(table))
	code.WriteString(".AddConstraint(")
	constraintCode(&code, table, constraint)
	code.WriteString(")")
	return code.String()
}

func dropConstraintCode(table, name string) string {
	return alterTableCode(table) + ".DropConstraint(" + quoteCode(name) + ")"
}

func alterTableCode(table string) string {
	return ddlCodeQualifier + "AlterTable(" + quoteCode(table) + ")"
}

func rawExpressionArgsCode(expressions []string) string {
	items := make([]string, len(expressions))
	for i := range expressions {
		items[i] = ddlCodeQualifier + "RawExpr(" + quoteCode(expressions[i]) + ")"
	}
	return strings.Join(items, ", ")
}

func stringArgsCode(items []string) string {
	quoted := make([]string, len(items))
	for i := range items {
		quoted[i] = quoteCode(items[i])
	}
	return strings.Join(quoted, ", ")
}

func constraintDropName(table string, constraint *Constraint) string {
	if constraint.Name != "" {
		return constraint.Name
	}

	switch constraint.Kind {
	case ConstraintPrimaryKey:
		return generatedConstraintName("pk", append([]string{table}, constraint.Columns...)...)
	case ConstraintUnique:
		return generatedConstraintName("uq", append([]string{table}, constraint.Columns...)...)
	case ConstraintCheck:
		return generatedCheckName(table, constraint.CheckExpr)
	case ConstraintForeignKey:
		parts := append([]string{table}, constraint.Columns...)
		parts = append(parts, constraint.Reference.TableName)
		parts = append(parts, constraint.Reference.Columns...)
		return generatedConstraintName("fk", parts...)
	default:
		return generatedConstraintName(string(constraint.Kind), table)
	}
}

func generatedConstraintName(prefix string, parts ...string) string {
	return fmt.Sprintf("%s_%s", prefix, strings.Join(parts, "_"))
}

func generatedCheckName(table, expr string) string {
	fields := strings.Fields(strings.ToLower(expr))
	normalized := strings.Join(fields, " ")
	sum := sha256.Sum256([]byte(normalized))
	hash := hex.EncodeToString(sum[:])[:8]
	return fmt.Sprintf("chk_%s_%s", table, hash)
}

func quoteCode(value string) string {
	return strconv.Quote(value)
}
