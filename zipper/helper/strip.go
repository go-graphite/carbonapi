package helper

import (
	"strings"
	"unicode/utf8"
)

// Aggressively strips HTML tags from a string.
// It will only keep anything between `>` and `<`.
func stripHtmlTags(s string, maxLen int) string {
	var n int
	if !strings.Contains(s, "<html>") {
		if maxLen == 0 || maxLen > len(s) {
			return s
		}
		return s[:maxLen]
	}
	// Setup a string builder and allocate enough memory for the new string.
	var builder strings.Builder
	if maxLen == 0 {
		n = len(s) + utf8.UTFMax
	} else {
		n = min(len(s), maxLen)
	}

	builder.Grow(n)

	in := false // True if we are inside an HTML tag.
	start := 0  // The index of the previous start tag character `<`
	end := 0    // The index of the previous end tag character `>`

	for i, c := range s {
		// If this is the last character and we are not in an HTML tag, save it.
		if (i+1) == len(s) && end >= start {
			builder.WriteString(s[end:])
		}

		if c == htmlTagStart {
			// Only update the start if we are not in a tag.
			// This make sure we strip out `<<br>` not just `<br>`
			if !in {
				start = i
			}
			in = true

			// Write the valid string between the close and start of the two tags.
			builder.WriteString(s[end:start])
			end = i + 1
		} else if c == htmlTagEnd {
			in = false
			end = i + 1
		}
	}
	s = strings.Trim(builder.String(), "\r\n")
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	htmlTagStart = 60 // Unicode `<`
	htmlTagEnd   = 62 // Unicode `>`
)
