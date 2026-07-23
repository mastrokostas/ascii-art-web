package ascii

import (
	"fmt"
	"html"
	"strings"
)

// Render takes a slice of input lines and a banner map returned by LoadBanner,
// and returns the full ASCII art output as a plain string.
func Render(lines []string, bannerMap map[rune][]string) string {
	var outputBuilder strings.Builder

	for _, line := range lines {
		// An empty line in the input produces a blank line in the output
		if line == "" {
			outputBuilder.WriteString("\n")
			continue
		}

		// Each character is 8 rows tall — build one output line per row
		for row := 0; row < 8; row++ {
			var rowBuilder strings.Builder

			// For each character in the input line, append its row slice from the banner map
			for _, char := range line {
				if block, ok := bannerMap[char]; ok {
					rowBuilder.WriteString(block[row])
				}
			}

			outputBuilder.WriteString(rowBuilder.String())
			outputBuilder.WriteString("\n")
		}
	}

	return outputBuilder.String()
}

// RenderWithColor works like Render but wraps each row in an HTML span
// with the given hex color applied as an inline style.
// The returned string is HTML and must be passed to the template as template.HTML.
func RenderWithColor(lines []string, bannerMap map[rune][]string, hexColor string) string {
	var outputBuilder strings.Builder

	for _, line := range lines {
		// An empty line in the input produces a blank line in the output
		if line == "" {
			outputBuilder.WriteString("\n")
			continue
		}

		// Each character is 8 rows tall — build one output line per row
		for row := 0; row < 8; row++ {
			var rowBuilder strings.Builder

			// For each character in the input line, append its row slice from the banner map
			for _, char := range line {
				if block, ok := bannerMap[char]; ok {
					rowBuilder.WriteString(block[row])
				}
			}

			// Wrap the completed row in a span with the hex color.
			// The row is HTML-escaped first: some banner glyphs contain literal
			// '<' and '>' characters, which would otherwise corrupt the markup.
			escapedRow := html.EscapeString(rowBuilder.String())
			outputBuilder.WriteString(fmt.Sprintf("<span style=\"color: %s\">%s</span>\n", hexColor, escapedRow))
		}
	}

	return outputBuilder.String()
}
