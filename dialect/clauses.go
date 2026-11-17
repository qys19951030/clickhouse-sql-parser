package dialect

import (
	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// inspectClauses walks the statements and appends one Finding per
// ClickHouse-proprietary clause keyword (PREWHERE, FINAL, SAMPLE,
// ARRAY JOIN, SETTINGS, ENGINE, TTL) to findings. It is a no-op until the
// clause stage lands.
func inspectClauses(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	return findings
}
