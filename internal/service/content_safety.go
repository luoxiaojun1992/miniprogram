package service

import (
	"regexp"
	"strings"
)

var htmlTagPattern = regexp.MustCompile(`(?i)<\s*/?\s*[a-z][^>]*>`)
var richTextUnsafePattern = regexp.MustCompile(`(?i)<\s*(script|iframe)\b|javascript\s*:`)

func hasHTMLTag(text string) bool {
	return htmlTagPattern.MatchString(strings.TrimSpace(text))
}

func hasUnsafeRichText(text string) bool {
	return richTextUnsafePattern.MatchString(strings.TrimSpace(text))
}

