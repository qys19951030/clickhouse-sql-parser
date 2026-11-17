package dialect

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// requireEqualFindings asserts the full finding list, including order,
// line/column coordinates, category, original fragment and suggestion.
func requireEqualFindings(t *testing.T, expected, actual []Finding) {
	t.Helper()
	require.Equal(t, expected, actual)
}

// onlyCategory drops findings of other categories so a stage-specific
// test table can assert against its own findings regardless of what the
// other detection stages report on the same input.
func onlyCategory(findings []Finding, category Category) []Finding {
	var out []Finding
	for _, finding := range findings {
		if finding.Category == category {
			out = append(out, finding)
		}
	}
	return out
}

// expectedTypeFindings builds the expected single-line type findings for
// names in source order, deriving columns from the SQL itself so test
// tables do not hand-count offsets.
func expectedTypeFindings(t *testing.T, sql string, names []string) []Finding {
	t.Helper()
	var out []Finding
	for _, name := range names {
		for _, column := range findColumns(t, sql, name) {
			out = append(out, Finding{
				Line:       1,
				Column:     column,
				Category:   CategoryType,
				Original:   name,
				Suggestion: suggestionType,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Column < out[j].Column })
	return out
}
