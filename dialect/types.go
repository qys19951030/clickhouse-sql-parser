package dialect

import (
	"sort"

	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// Remediation notes are fixed per category on purpose: every finding of
// the same category points at the same kind of migration work instead of
// pretending a generic suggestion could be a drop-in rewrite.
const (
	suggestionType     = "ClickHouse 专有类型写法，需按目标库语法重映射"
	suggestionClause   = "ClickHouse 专有子句，标准 SQL 无对应写法，需改写语句结构"
	suggestionFunction = "ClickHouse 专有函数，需改用标准函数或 CAST 表达式"
)

// Finding categories are matched against the exact, case-sensitive
// spelling used in ClickHouse grammar; lowercase or quoted lookalikes are
// identifiers, not the construct itself.
var proprietaryTypeNames = map[string]struct{}{
	"UInt8":          {},
	"UInt16":         {},
	"UInt32":         {},
	"UInt64":         {},
	"Int8":           {},
	"Int16":          {},
	"Int32":          {},
	"Int64":          {},
	"Float32":        {},
	"Float64":        {},
	"FixedString":    {},
	"DateTime":       {},
	"DateTime64":     {},
	"Decimal":        {},
	"LowCardinality": {},
	"Nullable":       {},
	"Array":          {},
	"Tuple":          {},
	"Map":            {},
	"Enum8":          {},
	"Enum16":         {},
	"IPv4":           {},
	"IPv6":           {},
	"UUID":           {},
}

// inspectTypes walks the statements and appends one Finding per
// ClickHouse-proprietary data type name. Nested types are reported
// separately because Walk visits every type node in the schema tree, so
// Nullable(UInt64) yields both an outer and an inner finding.
func inspectTypes(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	lines := newLineStarts(input)
	for _, stmt := range statements {
		clickhouse.Walk(stmt, func(node clickhouse.Expr) bool {
			name, pos, ok := typeNameAndPos(node)
			if !ok {
				return true
			}
			if _, ok := proprietaryTypeNames[name]; !ok {
				return true
			}
			line, column := lines.position(int(pos))
			findings = append(findings, Finding{
				Line:       line,
				Column:     column,
				Category:   CategoryType,
				Original:   input[int(pos) : int(pos)+len(name)],
				Suggestion: suggestionType,
			})
			return true
		})
	}
	return findings
}

// typeNameAndPos extracts the bare type name and the byte offset of its
// first character from a ColumnType syntax node. Every ColumnType
// implementation stores its name in an *Ident, including EnumType whose
// own Pos() points at the opening parenthesis rather than the name.
func typeNameAndPos(node clickhouse.Expr) (name string, pos clickhouse.Pos, ok bool) {
	switch n := node.(type) {
	case *clickhouse.ScalarType:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.TypeWithParams:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.TypeWithNamedParams:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.ComplexType:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.NestedType:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.EnumType:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.JSONType:
		return n.Name.Name, n.Name.NamePos, true
	case *clickhouse.PropertyType:
		return n.Name.Name, n.Name.NamePos, true
	}
	return "", 0, false
}

// sortFindings orders findings by source position, keeping every
// occurrence even when several constructs share the same fragment.
func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Column < findings[j].Column
	})
}
