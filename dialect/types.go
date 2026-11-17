package dialect

import (
	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// inspectTypes walks the statements and appends one Finding per
// ClickHouse-proprietary data type name to findings. It is a no-op until
// the type-name stage lands.
func inspectTypes(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	return findings
}
