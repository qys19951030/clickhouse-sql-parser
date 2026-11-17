package dialect

import (
	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// inspectFunctions walks the statements and appends one Finding per
// ClickHouse-proprietary function name to findings. It is a no-op until
// the function stage lands.
func inspectFunctions(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	return findings
}
