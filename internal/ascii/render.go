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

// RenderWithColor works like Render but wraps each row in an HTML span carrying
// the .art-line class. The class takes its color from the --art-color custom
// property, which the page sets from the validated hex color — the color is
// deliberately not written as an inline style="color:..." attribute, because
// that would force the Content-Security-Policy to allow 'unsafe-inline' styles.
//
// hexColor is therefore no longer written into the markup, but it is kept in the
// signature: callers pass the value they validated, and an empty string here
// still means "no color was chosen".
//
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

			// Wrap the completed row in a span that picks up the art color from
			// CSS. The row is HTML-escaped first: some banner glyphs contain
			// literal '<' and '>' characters, which would otherwise corrupt the
			// markup.
			escapedRow := html.EscapeString(rowBuilder.String())
			outputBuilder.WriteString(fmt.Sprintf("<span class=\"art-line\">%s</span>\n", escapedRow))
		}
	}

	return outputBuilder.String()
}
