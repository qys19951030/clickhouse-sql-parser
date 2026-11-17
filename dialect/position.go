package dialect

// lineStarts records the byte offset at which each line of the input
// begins, so byte offsets from parser positions can be converted to
// 1-based line/column coordinates without rescanning the input per
// finding.
type lineStarts []int

func newLineStarts(input string) lineStarts {
	starts := lineStarts{0}
	for i := 0; i < len(input); i++ {
		if input[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// position converts a byte offset into 1-based line and column numbers.
func (s lineStarts) position(offset int) (line, column int) {
	lo, hi := 0, len(s)
	for lo < hi {
		mid := (lo + hi) / 2
		if s[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	lineStart := s[lo-1]
	return lo, offset - lineStart + 1
}
