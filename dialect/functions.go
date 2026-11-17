package dialect

import (
	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// proprietaryFunctionNames lists the exact, case-sensitive function
// spellings treated as ClickHouse-specific. A match requires an
// unquoted call site: quoted identifiers (`f` or "f") are names chosen
// by the writer, not the ClickHouse grammar token.
var proprietaryFunctionNames = map[string]struct{}{
	"toString":          {},
	"toUInt64":          {},
	"toDate":            {},
	"toDateTime":        {},
	"countDistinct":     {},
	"uniq":              {},
	"anyLast":           {},
	"groupArray":        {},
	"JSONExtractString": {},
}

// inspectFunctions walks the statements and appends one Finding per
// ClickHouse-proprietary function call. Nested calls are each visited
// as their own FunctionExpr, so toString(uniq(a)) reports both
// functions. Function names appearing without a call argument list are
// plain identifiers and are not reported.
func inspectFunctions(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	lines := newLineStarts(input)
	for _, stmt := range statements {
		clickhouse.Walk(stmt, func(node clickhouse.Expr) bool {
			fn, ok := node.(*clickhouse.FunctionExpr)
			if !ok || fn.Name == nil || fn.Name.QuoteType != clickhouse.Unquoted {
				return true
			}
			name := fn.Name.Name
			if _, ok := proprietaryFunctionNames[name]; !ok {
				return true
			}
			line, column := lines.position(int(fn.Name.NamePos))
			findings = append(findings, Finding{
				Line:       line,
				Column:     column,
				Category:   CategoryFunction,
				Original:   name,
				Suggestion: suggestionFunction,
			})
			return true
		})
	}
	return findings
}
