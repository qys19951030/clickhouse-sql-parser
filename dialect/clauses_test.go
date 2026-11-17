package dialect

import "testing"

// TestClauses covers the seven ClickHouse-proprietary clause keywords:
// PREWHERE, FINAL, SAMPLE, ARRAY JOIN, SETTINGS, ENGINE and TTL,
// including boolean/marker-only clauses (FINAL, ARRAY JOIN), nested
// clauses and case-sensitive matching.
func TestClauses(t *testing.T) {
	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		{
			name: "FINAL after a table in a select",
			sql:  "SELECT id FROM t1 FINAL WHERE a = 1",
			expected: []Finding{
				{Line: 1, Column: 19, Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
			},
		},
		{
			name: "PREWHERE and ENGINE sit at their keyword positions",
			sql:  "CREATE TABLE db.t2 (a UInt64, b DateTime64) ENGINE = MergeTree ORDER BY a",
			expected: []Finding{
				{Line: 1, Column: 45, Category: CategoryClause, Original: "ENGINE", Suggestion: suggestionClause},
			},
		},
		{
			name: "PREWHERE in a select",
			sql:  "SELECT toString(a), uniq(b) FROM t3 PREWHERE c = 2",
			expected: []Finding{
				{Line: 1, Column: 37, Category: CategoryClause, Original: "PREWHERE", Suggestion: suggestionClause},
			},
		},
		{
			name: "ARRAY JOIN spans both keywords verbatim",
			sql:  "SELECT * FROM t1 ARRAY JOIN arr",
			expected: []Finding{
				{Line: 1, Column: 18, Category: CategoryClause, Original: "ARRAY JOIN", Suggestion: suggestionClause},
			},
		},
		{
			name: "ARRAY JOIN keeps the source text when a comment separates the keywords",
			sql:  "SELECT * FROM t ARRAY/* c */JOIN arr",
			expected: []Finding{
				{Line: 1, Column: 17, Category: CategoryClause, Original: "ARRAY/* c */JOIN", Suggestion: suggestionClause},
			},
		},
		{
			name: "SAMPLE ratio clause",
			sql:  "SELECT * FROM t SAMPLE 1/10",
			expected: []Finding{
				{Line: 1, Column: 17, Category: CategoryClause, Original: "SAMPLE", Suggestion: suggestionClause},
			},
		},
		{
			name: "SAMPLE BY in an engine clause",
			sql:  "CREATE TABLE t (a UInt64) ENGINE = MergeTree SAMPLE BY a ORDER BY a",
			expected: []Finding{
				{Line: 1, Column: 27, Category: CategoryClause, Original: "ENGINE", Suggestion: suggestionClause},
				{Line: 1, Column: 46, Category: CategoryClause, Original: "SAMPLE", Suggestion: suggestionClause},
			},
		},
		{
			name: "SETTINGS in a select",
			sql:  "SELECT * FROM t SETTINGS max_threads = 1",
			expected: []Finding{
				{Line: 1, Column: 17, Category: CategoryClause, Original: "SETTINGS", Suggestion: suggestionClause},
			},
		},
		{
			name: "column and table level TTL are both reported",
			sql:  "CREATE TABLE t (a UInt64 TTL now() + INTERVAL 1 DAY) ENGINE = Memory TTL a + INTERVAL 1 MONTH",
			expected: []Finding{
				{Line: 1, Column: 26, Category: CategoryClause, Original: "TTL", Suggestion: suggestionClause},
				{Line: 1, Column: 54, Category: CategoryClause, Original: "ENGINE", Suggestion: suggestionClause},
				{Line: 1, Column: 70, Category: CategoryClause, Original: "TTL", Suggestion: suggestionClause},
			},
		},
		{
			name: "FINAL in OPTIMIZE TABLE",
			sql:  "OPTIMIZE TABLE t FINAL",
			expected: []Finding{
				{Line: 1, Column: 18, Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
			},
		},
		{
			name: "FINAL coordinates restart on the second line",
			sql:  "SELECT * FROM t\nFINAL",
			expected: []Finding{
				{Line: 2, Column: 1, Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
			},
		},
		{
			name: "each statement's FINAL is reported separately",
			sql:  "SELECT * FROM t FINAL; SELECT * FROM u FINAL",
			expected: []Finding{
				{Line: 1, Column: 17, Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
				{Line: 1, Column: 40, Category: CategoryClause, Original: "FINAL", Suggestion: suggestionClause},
			},
		},
		{
			name:     "lowercase clause spelling does not match",
			sql:      "SELECT * FROM t final",
			expected: nil,
		},
		{
			name:     "clause keywords in strings and comments are not reported",
			sql:      "SELECT 'toString', 'ENGINE' FROM t6 -- PREWHERE FINAL",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := onlyCategory(Inspect(tc.sql, parseStatements(t, tc.sql)), CategoryClause)
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
