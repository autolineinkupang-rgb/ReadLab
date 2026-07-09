package mdimport

import (
	"regexp"
	"strings"
)

var (
	multiBlank   = regexp.MustCompile(`\n{3,}`)
	trailingWS   = regexp.MustCompile(`[ \t]+\n`)
	headingSpace = regexp.MustCompile(`(?m)^(#{1,6})([^ #\s])`)
	bomPrefix    = []byte{0xEF, 0xBB, 0xBF}
)

func CleanMarkdown(input string) string {
	// Remove BOM
	if len(input) >= 3 && input[0] == bomPrefix[0] && input[1] == bomPrefix[1] && input[2] == bomPrefix[2] {
		input = input[3:]
	}

	// Normalize line endings
	input = strings.ReplaceAll(input, "\r\n", "\n")
	input = strings.ReplaceAll(input, "\r", "\n")

	// Trailing whitespace
	input = trailingWS.ReplaceAllString(input, "\n")

	// Collapse multiple blank lines to max 2
	input = multiBlank.ReplaceAllString(input, "\n\n")

	// Ensure space after heading markers
	input = headingSpace.ReplaceAllString(input, "$1 $2")

	// Normalize thematic breaks
	input = regexp.MustCompile(`(?m)^(\*\*\*|___)\s*$`).ReplaceAllString(input, "---")

	// Normalize unordered list markers: * and + to -
	input = regexp.MustCompile(`(?m)^([ \t]*)[*+](\s)`).ReplaceAllString(input, "${1}-${2}")

	// Trim leading/trailing whitespace
	input = strings.TrimSpace(input)

	return input
}
