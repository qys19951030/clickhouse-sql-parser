// Package dialect inspects parsed ClickHouse SQL and reports the
// ClickHouse-specific constructs that do not carry over as-is when the
// statement is moved to another SQL engine: proprietary type names,
// clause keywords and function names.
//
// The package never mutates the parser's AST and performs no SQL parsing
// of its own: callers parse the input with parser.Parser and hand the
// resulting statements to Inspect together with the original source text.
package dialect

import (
	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// Category classifies a ClickHouse-specific construct reported in a
// Finding.
type Category string

const (
	// CategoryType marks a ClickHouse-proprietary data type name.
	CategoryType Category = "类型"
	// CategoryClause marks a ClickHouse-proprietary clause keyword.
	CategoryClause Category = "子句"
	// CategoryFunction marks a ClickHouse-proprietary function name.
	CategoryFunction Category = "函数"
)

// Finding is a single ClickHouse-specific fragment located in the input.
//
// Line and Column are 1-based coordinates of the fragment's first byte.
// Original is the verbatim source text of the matched fragment and
// Suggestion is the fixed, per-category remediation note.
type Finding struct {
	Line       int
	Column     int
	Category   Category
	Original   string
	Suggestion string
}

// Inspect walks the parsed statements and returns every ClickHouse-specific
// type, clause and function usage, ordered by source position
// (line, then column). The input string must be the exact source text the
// statements were parsed from; it is used for coordinate conversion and to
// recover the original spelling of constructs whose AST only records a
// boolean marker.
func Inspect(input string, statements []clickhouse.Expr) []Finding {
	return nil
}
