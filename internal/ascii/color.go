package ascii

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidateHexColor checks that the input is a valid #rrggbb hex color string.
// Returns the normalized lowercase hex string, or an error if invalid.
func ValidateHexColor(color string) (string, error) {
	color = strings.TrimSpace(color)

	if !strings.HasPrefix(color, "#") {
		return "", fmt.Errorf("invalid color: %q — must start with #", color)
	}

	hex := color[1:]
	if len(hex) != 6 {
		return "", fmt.Errorf("invalid color: %q — must be #rrggbb", color)
	}

	for i := 0; i < 6; i++ {
		_, err := strconv.ParseUint(string(hex[i]), 16, 8)
		if err != nil {
			return "", fmt.Errorf("invalid color: %q — non-hex character at position %d", color, i+1)
		}
	}

	return strings.ToLower(color), nil
}
