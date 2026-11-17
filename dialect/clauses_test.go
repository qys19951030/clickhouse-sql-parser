package dialect

import "testing"

// TestClauses holds the cases for ClickHouse-proprietary clause keyword
// detection. The table is populated in the clause stage.
func TestClauses(t *testing.T) {
	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		// cases are added with the clause implementation.
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := Inspect(tc.sql, parseStatements(t, tc.sql))
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
