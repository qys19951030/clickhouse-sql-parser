package dialect

import (
	"sort"
	"strings"
	"testing"
)

// findColumns returns the 1-based columns of every standalone occurrence
// of fragment in sql: an occurrence counts only when the characters on
// both sides are not identifier characters, so "Int8" inside "UInt8" is
// not mistaken for a match.
func findColumns(t *testing.T, sql, fragment string) []int {
	t.Helper()
	isIdentPart := func(b byte) bool {
		return b == '_' || b == '$' ||
			('0' <= b && b <= '9') || ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
	}
	var columns []int
	for offset := 0; offset+len(fragment) <= len(sql); {
		idx := strings.Index(sql[offset:], fragment)
		if idx < 0 {
			break
		}
		idx += offset
		beforeOK := idx == 0 || !isIdentPart(sql[idx-1])
		end := idx + len(fragment)
		afterOK := end >= len(sql) || !isIdentPart(sql[end])
		if beforeOK && afterOK {
			columns = append(columns, idx+1)
		}
		offset = end
	}
	return columns
}

// TestTypeNames covers ClickHouse-proprietary data type detection across
// the parser's ColumnType node variants: bare scalar types, parenthesized
// parameterized and complex types, enums, nesting and multi-line input.
func TestTypeNames(t *testing.T) {
	allTypesSQL := "CREATE TABLE t (a UInt8, b UInt16, c UInt32, d Int8, e Int16, f Int32, " +
		"g Float32, h FixedString(4), i DateTime, j Decimal(10, 2), k LowCardinality(String), " +
		"l Array(UInt64), m Tuple(Int64), n Map(String, UInt32), o Enum8('x' = 1), " +
		"p Enum16('y' = 2), q IPv4, r IPv6, s UUID, t Float64) ENGINE = Memory"
	// String is intentionally absent: it is a standard SQL type, not in the
	// proprietary list, and must not be reported. Every name appears once
	// here; repeated SQL occurrences (the two UInt32 columns) are picked up
	// by findColumns scanning the whole statement.
	allTypeNames := []string{
		"UInt8", "UInt16", "UInt32", "Int8", "Int16", "Int32",
		"Float32", "FixedString", "DateTime", "Decimal", "LowCardinality",
		"Array", "UInt64", "Tuple", "Int64", "Map",
		"Enum8", "Enum16", "IPv4", "IPv6", "UUID", "Float64",
	}
	allExpected := make([]Finding, 0, len(allTypeNames)+1)
	for _, typeName := range allTypeNames {
		for _, column := range findColumns(t, allTypesSQL, typeName) {
			allExpected = append(allExpected, Finding{
				Line:       1,
				Column:     column,
				Category:   CategoryType,
				Original:   typeName,
				Suggestion: suggestionType,
			})
		}
	}
	// duplicates such as the two UInt32 occurrences stay in source order
	sort.Slice(allExpected, func(i, j int) bool { return allExpected[i].Column < allExpected[j].Column })

	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		{
			name: "scalar and parameterized types in CREATE TABLE",
			sql:  "CREATE TABLE db.t2 (a UInt64, b DateTime64) ENGINE = MergeTree ORDER BY a",
			expected: []Finding{
				{Line: 1, Column: 23, Category: CategoryType, Original: "UInt64", Suggestion: suggestionType},
				{Line: 1, Column: 33, Category: CategoryType, Original: "DateTime64", Suggestion: suggestionType},
			},
		},
		{
			name: "nested complex type reports every layer",
			sql:  "CREATE TABLE t5 (a Nullable(UInt64))",
			expected: []Finding{
				{Line: 1, Column: 20, Category: CategoryType, Original: "Nullable", Suggestion: suggestionType},
				{Line: 1, Column: 29, Category: CategoryType, Original: "UInt64", Suggestion: suggestionType},
			},
		},
		{
			name:     "all listed type names match once per occurrence",
			sql:      allTypesSQL,
			expected: allExpected,
		},
		{
			name: "type in CAST expression is a type position",
			sql:  "SELECT CAST(x AS UInt64) FROM t",
			expected: []Finding{
				{Line: 1, Column: 18, Category: CategoryType, Original: "UInt64", Suggestion: suggestionType},
			},
		},
		{
			name:     "type name inside a string literal is not reported",
			sql:      "SELECT 'UInt64' FROM t",
			expected: nil,
		},
		{
			name:     "matching spelling is case sensitive",
			sql:      "CREATE TABLE t (a uint64, b UInt64x) ENGINE = Memory",
			expected: nil,
		},
		{
			name: "repeated occurrences are each reported",
			sql:  "CREATE TABLE t (a Nullable(UInt64), b Array(UInt64)) ENGINE = Memory",
			expected: expectedTypeFindings(t,
				"CREATE TABLE t (a Nullable(UInt64), b Array(UInt64)) ENGINE = Memory",
				[]string{"Nullable", "UInt64", "Array"}),
		},
		{
			name: "multi line coordinates restart at each line",
			sql:  "CREATE TABLE t (\n  a UInt64,\n  b Nullable(String)\n) ENGINE = Memory",
			expected: []Finding{
				{Line: 2, Column: 5, Category: CategoryType, Original: "UInt64", Suggestion: suggestionType},
				{Line: 3, Column: 5, Category: CategoryType, Original: "Nullable", Suggestion: suggestionType},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := onlyCategory(Inspect(tc.sql, parseStatements(t, tc.sql)), CategoryType)
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
