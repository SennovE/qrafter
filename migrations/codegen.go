package migrations

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/SennovE/qrafter/ddl"
)

const ddlCodeQualifier = "qddl."

var simpleDDLTypeCodes = []struct {
	code string
	typ  ddl.Type
}{
	{"SmallInt", ddl.SmallInt()},
	{"Integer", ddl.Integer()},
	{"BigInt", ddl.BigInt()},
	{"Serial", ddl.Serial()},
	{"BigSerial", ddl.BigSerial()},
	{"Text", ddl.Text()},
	{"Boolean", ddl.Boolean()},
	{"Date", ddl.Date()},
	{"Time", ddl.Time()},
	{"Timestamp", ddl.Timestamp()},
	{"TimestampTZ", ddl.TimestampTZ()},
	{"UUID", ddl.UUID()},
	{"JSON", ddl.JSON()},
	{"JSONB", ddl.JSONB()},
	{"Binary", ddl.Binary()},
	{"Float", ddl.Float()},
	{"Double", ddl.Double()},
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
		code.WriteString(".\n\tColumns(\n")
		joinColumnCodes(&code, table.Columns)
		code.WriteString("\t)")
	}
	if len(table.Constraints) > 0 {
		code.WriteString(".\n\tConstraints(\n")
		joinConstraintCodes(&code, table.Name, table.Constraints)
		code.WriteString("\t)")
	}
	return code.String()
}

func joinColumnCodes(w *strings.Builder, columns []Column) {
	for i := range columns {
		writeIndentedCode(w, columnCode(&columns[i]), "\t\t")
		w.WriteString(",\n")
	}
}

func joinConstraintCodes(w *strings.Builder, table string, constraints []Constraint) {
	for i := range constraints {
		writeIndentedCode(w, constraintCode(table, &constraints[i]), "\t\t")
		w.WriteString(",\n")
	}
}

func columnCode(column *Column) string {
	var code strings.Builder
	w := &code
	w.WriteString(ddlCodeQualifier)
	w.WriteString("Column(")
	w.WriteString(quoteCode(column.Name))
	w.WriteString(", ")
	w.WriteString(typeCode(column.ddlType()))
	w.WriteString(")")
	switch column.Identity {
	case ddl.IdentityAlways:
		w.WriteString(".IdentityAlways()")
	case ddl.IdentityByDefault:
		w.WriteString(".IdentityByDefault()")
	}
	switch column.Generated {
	case ddl.GeneratedStored:
		w.WriteString(".GeneratedStored(")
		w.WriteString(quoteCode(column.GeneratedExpr))
		w.WriteString(")")
	case ddl.GeneratedVirtual:
		w.WriteString(".GeneratedVirtual(")
		w.WriteString(quoteCode(column.GeneratedExpr))
		w.WriteString(")")
	}
	if column.NotNull {
		w.WriteString(".NotNull()")
	}
	if column.HasDefault && column.Identity == ddl.IdentityNone && column.Generated == ddl.GeneratedNone {
		w.WriteString(".DefaultExpr(")
		w.WriteString(quoteCode(column.DefaultExpr))
		w.WriteString(")")
	}
	return code.String()
}

func typeCode(typ ddl.Type) string {
	if code, ok := simpleTypeCode(typ); ok {
		return code
	}
	if code, ok := parameterizedTypeCode(typ); ok {
		return code
	}

	keys := make([]string, 0, len(typ.DialectNames))
	for dialectName := range typ.DialectNames {
		keys = append(keys, dialectName)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ddlCodeQualifier + "SQLType(" + quoteCode(typ.Name) + ")"
	}

	var code strings.Builder
	code.WriteString(ddlCodeQualifier + "SQLType(" + quoteCode(typ.Name) + ")")
	for _, dialectName := range keys {
		code.WriteString(".ForDialect(" + quoteCode(dialectName) + ", " + quoteCode(typ.DialectNames[dialectName]) + ")")
	}
	return code.String()
}

func simpleTypeCode(typ ddl.Type) (string, bool) {
	for i := range simpleDDLTypeCodes {
		if typesEqual(typ, simpleDDLTypeCodes[i].typ) {
			return ddlCodeQualifier + simpleDDLTypeCodes[i].code + "()", true
		}
	}
	return "", false
}

func parameterizedTypeCode(typ ddl.Type) (string, bool) {
	if len(typ.DialectNames) > 0 {
		return "", false
	}
	if size, ok := parseWrappedIntType(typ.Name, "VARCHAR"); ok {
		return ddlCodeQualifier + "VarChar(" + strconv.Itoa(size) + ")", true
	}
	if size, ok := parseWrappedIntType(typ.Name, "CHAR"); ok {
		return ddlCodeQualifier + "Char(" + strconv.Itoa(size) + ")", true
	}
	if precision, scale, ok := parseNumericTypeCodeArgs(typ.Name); ok {
		return ddlCodeQualifier + "Numeric(" + strconv.Itoa(precision) + ", " + strconv.Itoa(scale) + ")", true
	}
	return "", false
}

func parseWrappedIntType(name, prefix string) (int, bool) {
	inner, ok := parseTypeCodeArgs(name, prefix)
	if !ok {
		return 0, false
	}
	size, err := strconv.Atoi(strings.TrimSpace(inner))
	return size, err == nil && size > 0
}

func parseNumericTypeCodeArgs(name string) (precision, scale int, ok bool) {
	inner, ok := parseTypeCodeArgs(name, "NUMERIC")
	if !ok {
		return 0, 0, false
	}
	parts := strings.Split(inner, ",")
	if len(parts) == 0 || len(parts) > 2 {
		return 0, 0, false
	}
	precision, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || precision <= 0 {
		return 0, 0, false
	}
	if len(parts) == 1 {
		return precision, 0, true
	}
	scale, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || scale < 0 {
		return 0, 0, false
	}
	return precision, scale, true
}

func parseTypeCodeArgs(name, prefix string) (string, bool) {
	if !strings.HasPrefix(name, prefix+"(") || !strings.HasSuffix(name, ")") {
		return "", false
	}
	return name[len(prefix)+1 : len(name)-1], true
}

func constraintCode(table string, constraint *Constraint) string {
	var code strings.Builder
	w := &code
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
	if constraint.Name != "" {
		w.WriteString(".Named(")
		w.WriteString(quoteCode(constraintDropName(table, constraint)))
		w.WriteString(")")
	}
	return code.String()
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
	code.WriteString(").On(\n\t\t")
	code.WriteString(quoteCode(index.TableName))
	for i := range index.Keys {
		code.WriteString(",\n")
		code.WriteString(indexKeyCode(&index.Keys[i]))
	}
	code.WriteString(",\n\t)")

	if index.Unique {
		code.WriteString(".\n\tUnique()")
	}
	if index.Method != "" {
		code.WriteString(".\n\tUsing(" + indexMethodCode(index.Method) + ")")
	}
	if len(index.Include) > 0 {
		code.WriteString(".\n\tInclude(\n")
		for _, expr := range index.Include {
			writeIndentedCode(&code, rawExpressionCode(expr), "\t\t")
			code.WriteString(",\n")
		}
		code.WriteString("\t)")
	}
	if index.NullsNotDistinct {
		code.WriteString(".\n\tNullsNotDistinct()")
	}
	if index.Tablespace != "" {
		code.WriteString(".\n\tTablespace(" + quoteCode(index.Tablespace) + ")")
	}
	if index.Predicate != "" {
		code.WriteString(".\n\tWhere(" + ddlCodeQualifier + "RawPred(" + quoteCode(index.Predicate) + "))")
	}
	return code.String()
}

func indexKeyCode(key *IndexKey) string {
	if column, ok := simpleColumnExpression(key.Expression); ok {
		return ddlCodeQualifier + "KeyCol(" + quoteCode(column) + ")"
	}
	return ddlCodeQualifier + "Key(" + rawExpressionCode(key.Expression) + ")"
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
	code.WriteString(".AddColumn(\n")
	writeIndentedCode(&code, columnCode(column), "\t\t")
	code.WriteString(",\n\t)")
	return code.String()
}

func dropColumnCode(table, column string) string {
	return alterTableCode(table) + ".DropColumn(" + quoteCode(column) + ")"
}

func replaceColumnCode(table, dropColumn string, addColumn *Column) string {
	var code strings.Builder
	code.WriteString(alterTableCode(table))
	code.WriteString(".\n\tDropColumn(")
	code.WriteString(quoteCode(dropColumn))
	code.WriteString(")")
	code.WriteString(".\n\tAddColumn(\n")
	writeIndentedCode(&code, columnCode(addColumn), "\t\t")
	code.WriteString(",\n\t)")
	return code.String()
}

func alterColumnTypeCode(table string, column *Column) string {
	var code strings.Builder
	code.WriteString(alterTableCode(table))
	code.WriteString(".AlterColumnType(")
	code.WriteString(quoteCode(column.Name))
	code.WriteString(", ")
	code.WriteString(typeCode(column.ddlType()))
	code.WriteString(")")
	return code.String()
}

func notNullCode(table string, column *Column) string {
	if column.NotNull {
		return alterTableCode(table) + ".\n\tSetNotNull(" + quoteCode(column.Name) + ")"
	}
	return alterTableCode(table) + ".\n\tDropNotNull(" + quoteCode(column.Name) + ")"
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
	code.WriteString(".AddConstraint(\n")
	code.WriteString(constraintCode(table, constraint))
	code.WriteString(",\n\t)")
	return code.String()
}

func dropConstraintCode(table, name string) string {
	return alterTableCode(table) + ".DropConstraint(" + quoteCode(name) + ")"
}

func alterTableCode(table string) string {
	return ddlCodeQualifier + "AlterTable(" + quoteCode(table) + ")"
}

func rawExpressionCode(expression string) string {
	if column, ok := simpleColumnExpression(expression); ok {
		return ddlCodeQualifier + "Col(" + quoteCode(column) + ")"
	}
	return ddlCodeQualifier + "RawExpr(" + quoteCode(expression) + ")"
}

func writeIndentedCode(w *strings.Builder, code, indent string) {
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		if line != "" {
			w.WriteString(indent)
			w.WriteString(line)
		}
		if i < len(lines)-1 {
			w.WriteString("\n")
		}
	}
}

func simpleColumnExpression(expression string) (string, bool) {
	expression = strings.TrimSpace(expression)
	if expression == "" {
		return "", false
	}
	if strings.HasPrefix(expression, `"`) {
		return parseQuotedIdentifierExpression(expression)
	}
	if !isUnquotedIdentifier(expression) {
		return "", false
	}
	return expression, true
}

func parseQuotedIdentifierExpression(expression string) (string, bool) {
	var name strings.Builder
	for i := 1; i < len(expression); i++ {
		if expression[i] != '"' {
			name.WriteByte(expression[i])
			continue
		}
		if i+1 < len(expression) && expression[i+1] == '"' {
			name.WriteByte('"')
			i++
			continue
		}
		if i != len(expression)-1 || name.Len() == 0 {
			return "", false
		}
		return name.String(), true
	}
	return "", false
}

func isUnquotedIdentifier(expression string) bool {
	for i, r := range expression {
		if i == 0 && !isIdentifierStart(r) {
			return false
		}
		if i > 0 && !isIdentifierContinue(r) {
			return false
		}
	}
	return true
}

func isIdentifierStart(r rune) bool {
	return r == '_' || isASCIILetter(r)
}

func isIdentifierContinue(r rune) bool {
	return isIdentifierStart(r) || r == '$' || isASCIIDigit(r)
}

func isASCIILetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
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
