package stringsx

import (
	"fmt"
	"regexp"
	"strings"
)

func ReduceNewlines(s string, maxNewlines int) string {
	if maxNewlines < 1 {
		return s
	}

	// Create the replacement string based on maxNewlines.
	replacement := strings.Repeat("\n", maxNewlines)

	// This pattern will match sequences of (maxNewlines+1) or more newlines, with spaces or tabs in between.
	pattern := regexp.MustCompile(fmt.Sprintf(`(\n[ \t]*){%d,}`, maxNewlines+1))
	s = pattern.ReplaceAllString(s, replacement)

	return s
}
