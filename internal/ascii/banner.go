package ascii

import (
	"io/fs"
	"strings"
)

// LoadBanner reads a banner .txt file from the given filesystem and returns
// a map where each rune (ASCII 32–126) maps to its 8-line graphical representation.
func LoadBanner(fsys fs.FS, filename string) (map[rune][]string, error) {
	// Read the entire banner file as a byte slice
	byteSlice, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return nil, err
	}

	// Convert to string and trim the leading newline that precedes the first character (space)
	content := strings.TrimPrefix(string(byteSlice), "\n")

	// Split the file into character blocks — each block is separated by a blank line
	characters := strings.Split(content, "\n\n")

	// Build the map: ASCII 32 (space) is the first character, incrementing by 1 for each block
	bannerMap := make(map[rune][]string)
	for i, block := range characters {
		lines := strings.Split(block, "\n")
		bannerMap[rune(32+i)] = lines
	}

	return bannerMap, nil
}
