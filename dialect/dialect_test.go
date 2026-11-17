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

// TestSkeleton verifies the empty package end to end: inspection of a
// statement containing no recognized construct returns no findings and
// stays ordered. Concrete assertions are added alongside each detection
// stage.
func TestSkeleton(t *testing.T) {
	findings := Inspect("SELECT a FROM t WHERE a = 2",
		parseStatements(t, "SELECT a FROM t WHERE a = 2"))
	require.Empty(t, findings)
}
