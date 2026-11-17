package dialect

import "testing"

// TestTypeNames holds the cases for ClickHouse-proprietary data type
// detection. The table is populated in the type-name stage.
func TestTypeNames(t *testing.T) {
	cases := []struct {
		name     string
		sql      string
		expected []Finding
	}{
		// cases are added with the type-name implementation.
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			findings := Inspect(tc.sql, parseStatements(t, tc.sql))
			requireEqualFindings(t, tc.expected, findings)
		})
	}
}
