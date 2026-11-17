package dialect

import (
	"testing"

	"github.com/stretchr/testify/require"

	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// parseStatements parses sql with the project parser, failing the test on
// a parse error so the dialect tests stay focused on inspection output.
func parseStatements(t *testing.T, sql string) []clickhouse.Expr {
	t.Helper()
	stmts, err := clickhouse.NewParser(sql).ParseStmts()
	require.NoError(t, err)
	return stmts
}

// TestInspectNoFindings covers statements that use none of the listed
// ClickHouse-specific constructs: inspection returns an empty list rather
// than a nil-dependent placeholder, and the CLI then prints nothing.
func TestInspectNoFindings(t *testing.T) {
	cases := []string{
		"SELECT a FROM t4 WHERE a = 2",
		"SELECT 'toString', 'ENGINE' FROM t6 -- PREWHERE FINAL",
	}
	for _, sql := range cases {
		t.Run(sql, func(t *testing.T) {
			require.Empty(t, Inspect(sql, parseStatements(t, sql)))
		})
	}
}

// TestInspectMixedCategoriesAndOrdering asserts the end-to-end report
// across all three categories, including the global line/column ordering
// across multiple statements and every fixed suggestion text.
func TestInspectMixedCategoriesAndOrdering(t *testing.T) {
	sql := "SELECT toString(a) FROM t; " +
		"CREATE TABLE u (x Nullable(UInt64)) ENGINE = Memory; " +
		"SELECT id FROM v FINAL"
	findings := Inspect(sql, parseStatements(t, sql))

	// Columns are derived from the source text so the assertion cannot
	// drift from hand-counted offsets.
	col := func(fragment string) int {
		columns := findColumns(t, sql, fragment)
		require.Len(t, columns, 1)
		return columns[0]
	}
	expected := []Finding{
		{Line: 1, Column: col("toString"), Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
		{Line: 1, Column: col("Nullable"), Category: CategoryType, Original: "Nullable", Suggestion: suggestionType},
		{Line: 1, Column: col("UInt64"), Category: CategoryType, Original: "UInt64", Suggestion: suggestionType},
		{Line: 1, Column: col("ENGINE"), Category: CategoryClause, Original: "ENGINE", Suggestion: suggestionClause},
		{Line: 1, Column: col("FINAL"), Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
	}
	requireEqualFindings(t, expected, findings)

	// ordering invariant: non-decreasing line, then column
	for i := 1; i < len(findings); i++ {
		prev, cur := findings[i-1], findings[i]
		require.True(t,
			prev.Line < cur.Line || (prev.Line == cur.Line && prev.Column <= cur.Column),
			"findings not sorted at index %d: %+v after %+v", i, cur, prev)
	}
}

// TestInspectMultiLineCoordinates checks line tracking across newlines
// for every category.
func TestInspectMultiLineCoordinates(t *testing.T) {
	sql := "SELECT\n  toString(a),\n  uniq(b)\nFROM t\nPREWHERE c = 2"
	findings := Inspect(sql, parseStatements(t, sql))
	expected := []Finding{
		{Line: 2, Column: 3, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
		{Line: 3, Column: 3, Category: CategoryFunction, Original: "uniq", Suggestion: suggestionFunction},
		{Line: 5, Column: 1, Category: CategoryClause, Original: "PREWHERE", Suggestion: suggestionClause},
	}
	requireEqualFindings(t, expected, findings)
}
