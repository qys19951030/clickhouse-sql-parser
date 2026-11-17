package dialect

import "testing"

// TestFunctions covers ClickHouse-proprietary function call detection:
// exact case-sensitive spelling at unquoted FunctionExpr call sites,
// repeated and nested calls, and exclusion of strings, bare identifiers,
// quoted names and differently cased spellings.
func TestFunctions(t *testing.T) {
	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		{
			name: "two different functions plus a clause",
			sql:  "SELECT toString(a), uniq(b) FROM t3 PREWHERE c = 2",
			expected: []Finding{
				{Line: 1, Column: 8, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
				{Line: 1, Column: 21, Category: CategoryFunction, Original: "uniq", Suggestion: suggestionFunction},
			},
		},
		{
			name: "the same function twice is reported at each call",
			sql:  "SELECT toString(a), toString(b) FROM t7",
			expected: []Finding{
				{Line: 1, Column: 8, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
				{Line: 1, Column: 21, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
			},
		},
		{
			name: "nested calls each produce a finding",
			sql:  "SELECT toString(uniq(a)) FROM t",
			expected: []Finding{
				{Line: 1, Column: 8, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
				{Line: 1, Column: 17, Category: CategoryFunction, Original: "uniq", Suggestion: suggestionFunction},
			},
		},
		{
			name: "every listed function name matches at a call site",
			sql: "SELECT countDistinct(a), anyLast(b), groupArray(c), JSONExtractString(d, 'k'), " +
				"toUInt64(e), toDate(f), toDateTime(g), toString(h), uniq(i) FROM t",
			expected: expectedFunctionFindings(t,
				"SELECT countDistinct(a), anyLast(b), groupArray(c), JSONExtractString(d, 'k'), "+
					"toUInt64(e), toDate(f), toDateTime(g), toString(h), uniq(i) FROM t",
				[]string{
					"countDistinct", "anyLast", "groupArray", "JSONExtractString",
					"toUInt64", "toDate", "toDateTime", "toString", "uniq",
				}),
		},
		{
			name:     "a function name without a call is a plain identifier",
			sql:      "SELECT toString FROM t",
			expected: nil,
		},
		{
			name:     "a function name inside a string literal is not reported",
			sql:      "SELECT 'toString' FROM t",
			expected: nil,
		},
		{
			name:     "matching is case sensitive",
			sql:      "SELECT TOSTRING(a), ToString(b) FROM t",
			expected: nil,
		},
		{
			name:     "quoted function names are identifiers chosen by the writer",
			sql:      "SELECT `toString`(a) FROM t",
			expected: nil,
		},
		{
			name: "coordinates restart at each line",
			sql:  "SELECT\n  toString(a)\nFROM t",
			expected: []Finding{
				{Line: 2, Column: 3, Category: CategoryFunction, Original: "toString", Suggestion: suggestionFunction},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := onlyCategory(Inspect(tc.sql, parseStatements(t, tc.sql)), CategoryFunction)
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
