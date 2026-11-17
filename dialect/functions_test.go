package dialect

import "testing"

// TestFunctions holds the cases for ClickHouse-proprietary function name
// detection. The table is populated in the function stage.
func TestFunctions(t *testing.T) {
	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		// cases are added with the function implementation.
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := Inspect(tc.sql, parseStatements(t, tc.sql))
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
