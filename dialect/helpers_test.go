package dialect

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// requireEqualFindings asserts the full finding list, including order,
// line/column coordinates, category, original fragment and suggestion.
func requireEqualFindings(t *testing.T, expected, actual []Finding) {
	t.Helper()
	require.Equal(t, expected, actual)
}
