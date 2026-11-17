package dialect

import (
	"unicode/utf8"

	clickhouse "github.com/AfterShip/clickhouse-sql-parser/parser"
)

// keywordToken is a bare (unquoted) word token located in the input.
// Words inside string literals, quoted identifiers and comments are
// skipped, mirroring the parser lexer's structural rules.
type keywordToken struct {
	pos  int
	text string
}

// scanKeywordTokens returns every bare word token of the input in
// order. It is a deliberately small structural scan: clause positions
// already come from the AST, so this scanner only has to recover the
// source spelling for the two clauses the AST records as a boolean
// (FINAL) or a modifier list (ARRAY JOIN), and to keep words inside
// strings, quoted identifiers and comments out of the result.
func scanKeywordTokens(input string) []keywordToken {
	var tokens []keywordToken
	for i := 0; i < len(input); {
		c := input[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
		case c == '-' && i+1 < len(input) && input[i+1] == '-':
			i += 2
			for i < len(input) && input[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(input) && input[i+1] == '*':
			i += 2
			for i+1 < len(input) && !(input[i] == '*' && input[i+1] == '/') {
				i++
			}
			if i+1 < len(input) {
				i += 2
			} else {
				i = len(input)
			}
		case c == '\'':
			i++
			for i < len(input) {
				if input[i] == '\\' {
					i += 2
					continue
				}
				if input[i] == '\'' {
					if i+1 < len(input) && input[i+1] == '\'' {
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
		case c == '`' || c == '"':
			quote := c
			i++
			for i < len(input) && input[i] != quote {
				i++
			}
			if i < len(input) {
				i++
			}
		case isASCIIIdentStart(c):
			start := i
			for i < len(input) && isASCIIIdentPart(input[i]) {
				i++
			}
			tokens = append(tokens, keywordToken{pos: start, text: input[start:i]})
		default:
			// Non-ASCII runes cannot start an unquoted keyword; skip the
			// whole rune so multi-byte input stays byte aligned.
			if c >= utf8.RuneSelf {
				_, size := utf8.DecodeRuneInString(input[i:])
				i += size
				continue
			}
			i++
		}
	}
	return tokens
}

func isASCIIIdentStart(c byte) bool {
	return c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isASCIIIdentPart(c byte) bool {
	return isASCIIIdentStart(c) || ('0' <= c && c <= '9') || c == '$'
}

// findKeywordAtOrAfter returns the first token at or after start whose
// spelling equals keyword exactly (case sensitive), or -1 when no such
// token exists.
func findKeywordAtOrAfter(tokens []keywordToken, start int, keyword string) int {
	for _, token := range tokens {
		if token.pos < start {
			continue
		}
		if token.text == keyword {
			return token.pos
		}
	}
	return -1
}

// findArrayJoinSpan locates an ARRAY keyword at or after start that is
// immediately followed (the token list already excludes whitespace and
// comments) by a JOIN keyword, returning the span [pos, end) of the
// verbatim "ARRAY ... JOIN" source text. ok is false when no such pair
// exists. Only the canonical, case-sensitive spelling qualifies.
func findArrayJoinSpan(tokens []keywordToken, start int) (pos, end int, ok bool) {
	for i := range tokens {
		if tokens[i].pos < start || tokens[i].text != "ARRAY" {
			continue
		}
		if i+1 < len(tokens) && tokens[i+1].text == "JOIN" {
			return tokens[i].pos, tokens[i+1].pos + len("JOIN"), true
		}
	}
	return 0, 0, false
}

// inspectClauses walks the statements and appends one Finding per
// ClickHouse-proprietary clause keyword: PREWHERE, FINAL, SAMPLE,
// ARRAY JOIN, SETTINGS, ENGINE and TTL. Each construct is reported at
// the keyword's first byte; nested clauses (e.g. a column-level TTL
// inside a table that also carries a table-level TTL) are reported
// separately.
func inspectClauses(input string, statements []clickhouse.Expr, findings []Finding) []Finding {
	lines := newLineStarts(input)
	tokens := scanKeywordTokens(input)
	emit := func(pos int, original string) {
		line, column := lines.position(pos)
		findings = append(findings, Finding{
			Line:       line,
			Column:     column,
			Category:   CategoryClause,
			Original:   original,
			Suggestion: suggestionClause,
		})
	}

	for _, stmt := range statements {
		clickhouse.Walk(stmt, func(node clickhouse.Expr) bool {
			switch n := node.(type) {
			case *clickhouse.PrewhereClause:
				if hasCanonicalSpelling(input, int(n.PrewherePos), "PREWHERE") {
					emit(int(n.PrewherePos), "PREWHERE")
				}
			case *clickhouse.SampleClause:
				if hasCanonicalSpelling(input, int(n.SamplePos), "SAMPLE") {
					emit(int(n.SamplePos), "SAMPLE")
				}
			case *clickhouse.SampleByClause:
				if hasCanonicalSpelling(input, int(n.SamplePos), "SAMPLE") {
					emit(int(n.SamplePos), "SAMPLE")
				}
			case *clickhouse.SettingsClause:
				if hasCanonicalSpelling(input, int(n.SettingsPos), "SETTINGS") {
					emit(int(n.SettingsPos), "SETTINGS")
				}
			case *clickhouse.EngineExpr:
				if hasCanonicalSpelling(input, int(n.EnginePos), "ENGINE") {
					emit(int(n.EnginePos), "ENGINE")
				}
			case *clickhouse.TTLClause:
				if hasCanonicalSpelling(input, int(n.TTLPos), "TTL") {
					emit(int(n.TTLPos), "TTL")
				}
			case *clickhouse.TableExpr:
				if n.HasFinal {
					if pos := findKeywordAtOrAfter(tokens, int(n.End()), "FINAL"); pos >= 0 {
						emit(pos, "FINAL")
					}
				}
			case *clickhouse.JoinTableExpr:
				if n.HasFinal {
					if pos := findKeywordAtOrAfter(tokens, int(n.End()), "FINAL"); pos >= 0 {
						emit(pos, "FINAL")
					}
				}
			case *clickhouse.OptimizeStmt:
				if n.HasFinal {
					if pos := findKeywordAtOrAfter(tokens, 0, "FINAL"); pos >= 0 {
						emit(pos, "FINAL")
					}
				}
			case *clickhouse.JoinExpr:
				if !joinHasArrayModifier(n.Modifiers) {
					return true
				}
				if pos, end, ok := findArrayJoinSpan(tokens, int(n.JoinPos)); ok {
					emit(pos, input[pos:end])
				}
			}
			return true
		})
	}
	return findings
}

// hasCanonicalSpelling reports whether input[offset:] starts with the
// exact, case-sensitive keyword.
func hasCanonicalSpelling(input string, offset int, keyword string) bool {
	return offset+len(keyword) <= len(input) && input[offset:offset+len(keyword)] == keyword
}

// joinHasArrayModifier reports whether the parsed join modifiers mark an
// ARRAY JOIN, accepting the parser's own canonical spelling.
func joinHasArrayModifier(modifiers []string) bool {
	for _, modifier := range modifiers {
		if modifier == "ARRAY" || modifier == "ARRAY JOIN" {
			return true
		}
	}
	return false
}
