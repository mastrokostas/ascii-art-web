package ascii

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// ===================  Test fixtures  ===================

// build_banner_block returns the eight-line graphical block for one character,
// where every row is the given label followed by its row index. Distinct rows
// make it obvious in a failure message which row came out in the wrong order.
func build_banner_block(label string) []string {
	block_rows := make([]string, 8)

	for row_index := 0; row_index < 8; row_index++ {
		block_rows[row_index] = label + string(rune('0'+row_index))
	}

	return block_rows
}

// build_test_banner_map returns a banner map covering just the characters the
// render tests need. Building the map directly (rather than going through
// LoadBanner) keeps the renderer tests independent of the file parser, so a
// parsing bug cannot disguise itself as a rendering bug.
func build_test_banner_map() map[rune][]string {
	return map[rune][]string{
		'A': build_banner_block("A"),
		'B': build_banner_block("B"),
	}
}

// ===================  ValidateHexColor  ===================

// TestValidateHexColor_AcceptsValidColor checks the ordinary success case.
func TestValidateHexColor_AcceptsValidColor(t *testing.T) {
	normalised_color, validation_error := ValidateHexColor("#ff8800")
	if validation_error != nil {
		t.Fatalf("valid color was rejected: %v", validation_error)
	}

	if normalised_color != "#ff8800" {
		t.Errorf("normalised color: got %q, want %q", normalised_color, "#ff8800")
	}
}

// TestValidateHexColor_NormalisesToLowercase checks that an uppercase color is
// returned lowercased. The value is echoed back into the page and into the HTML
// export, so a stable casing keeps those outputs comparable.
func TestValidateHexColor_NormalisesToLowercase(t *testing.T) {
	normalised_color, validation_error := ValidateHexColor("#FF8800")
	if validation_error != nil {
		t.Fatalf("uppercase color was rejected: %v", validation_error)
	}

	if normalised_color != "#ff8800" {
		t.Errorf("normalised color: got %q, want %q", normalised_color, "#ff8800")
	}
}

// TestValidateHexColor_TrimsSurroundingWhitespace checks that padding around the
// value does not make an otherwise valid color fail.
func TestValidateHexColor_TrimsSurroundingWhitespace(t *testing.T) {
	normalised_color, validation_error := ValidateHexColor("  #ff8800\n")
	if validation_error != nil {
		t.Fatalf("padded color was rejected: %v", validation_error)
	}

	if normalised_color != "#ff8800" {
		t.Errorf("normalised color: got %q, want %q", normalised_color, "#ff8800")
	}
}

// TestValidateHexColor_RejectsMissingHash checks that a value without the "#"
// prefix is refused.
func TestValidateHexColor_RejectsMissingHash(t *testing.T) {
	_, validation_error := ValidateHexColor("ff8800")
	if validation_error == nil {
		t.Error("color without a leading # was accepted")
	}
}

// TestValidateHexColor_RejectsWrongLength checks both sides of the six-digit
// rule: the three-digit CSS shorthand is not supported, and an over-long value
// is refused rather than being silently truncated.
func TestValidateHexColor_RejectsWrongLength(t *testing.T) {
	wrong_length_colors := []string{"#fff", "#ff88000", "#"}

	for _, candidate_color := range wrong_length_colors {
		_, validation_error := ValidateHexColor(candidate_color)
		if validation_error == nil {
			t.Errorf("color %q with a wrong digit count was accepted", candidate_color)
		}
	}
}

// TestValidateHexColor_RejectsNonHexCharacter checks the per-character check
// itself. These values carry the "#" prefix and the correct length, so they can
// only be caught by the digit loop — the handler-level tests use values that
// fail on the prefix check first and never reach it.
func TestValidateHexColor_RejectsNonHexCharacter(t *testing.T) {
	non_hex_colors := []string{"#gg0000", "#12345z", "#ff 800", "#-f0000"}

	for _, candidate_color := range non_hex_colors {
		_, validation_error := ValidateHexColor(candidate_color)
		if validation_error == nil {
			t.Errorf("color %q with a non-hex digit was accepted", candidate_color)
		}
	}
}

// ===================  LoadBanner  ===================

// TestLoadBanner_MapsBlocksFromSpaceUpwards checks the core assumption of the
// parser: the first block in the file is ASCII 32 (space) and each following
// block is the next code point. A drift of one here shifts the entire alphabet.
func TestLoadBanner_MapsBlocksFromSpaceUpwards(t *testing.T) {
	// A miniature banner file in the real layout: a leading newline, then blocks
	// of eight rows separated by a blank line.
	banner_file_content := "\n" +
		strings.Join(build_banner_block("s"), "\n") + "\n\n" +
		strings.Join(build_banner_block("x"), "\n") + "\n\n" +
		strings.Join(build_banner_block("q"), "\n") + "\n"

	test_file_system := fstest.MapFS{
		"mini.txt": &fstest.MapFile{Data: []byte(banner_file_content)},
	}

	banner_map, load_error := LoadBanner(test_file_system, "mini.txt")
	if load_error != nil {
		t.Fatalf("loading the test banner failed: %v", load_error)
	}

	// Block 0 is ' ' (32), block 1 is '!' (33), block 2 is '"' (34).
	expected_labels := map[rune]string{' ': "s", '!': "x", '"': "q"}

	for expected_rune, expected_label := range expected_labels {
		block_rows, rune_is_present := banner_map[expected_rune]
		if !rune_is_present {
			t.Errorf("rune %q is missing from the banner map", expected_rune)
			continue
		}

		if block_rows[0] != expected_label+"0" {
			t.Errorf("rune %q first row: got %q, want %q", expected_rune, block_rows[0], expected_label+"0")
		}
	}
}

// TestLoadBanner_ReturnsErrorForMissingFile checks the read-error path. The
// handler distinguishes a missing banner (404) from any other read failure
// (500) by testing the returned error, so the error has to survive unwrapped.
func TestLoadBanner_ReturnsErrorForMissingFile(t *testing.T) {
	empty_file_system := fstest.MapFS{}

	banner_map, load_error := LoadBanner(empty_file_system, "does-not-exist.txt")
	if load_error == nil {
		t.Fatal("loading a missing banner returned no error")
	}

	if banner_map != nil {
		t.Error("a banner map was returned alongside the error, want nil")
	}

	if !errors.Is(load_error, fs.ErrNotExist) {
		t.Errorf("error does not unwrap to fs.ErrNotExist: %v", load_error)
	}
}

// TestLoadBanner_ReadsRealBannerFiles checks the three shipped banner files
// against the parser. This is the test that would catch a banner file being
// edited into a shape the parser no longer understands.
func TestLoadBanner_ReadsRealBannerFiles(t *testing.T) {
	// The tests run from the package directory, so the banners live two levels up.
	banner_file_system := os.DirFS("../../banners")
	banner_file_names := []string{"standard.txt", "shadow.txt", "thinkertoy.txt"}

	for _, banner_file_name := range banner_file_names {
		banner_map, load_error := LoadBanner(banner_file_system, banner_file_name)
		if load_error != nil {
			t.Errorf("loading %q failed: %v", banner_file_name, load_error)
			continue
		}

		// The printable ASCII range the site renders is 32 through 126.
		if len(banner_map) != 95 {
			t.Errorf("%q: got %d characters, want 95", banner_file_name, len(banner_map))
		}

		// Every character has to carry at least the eight rows the renderer indexes,
		// otherwise rendering that character panics on an out-of-range row.
		for character_rune := rune(32); character_rune <= 126; character_rune++ {
			block_rows, rune_is_present := banner_map[character_rune]
			if !rune_is_present {
				t.Errorf("%q: rune %q is missing", banner_file_name, character_rune)
				continue
			}

			if len(block_rows) < 8 {
				t.Errorf("%q: rune %q has %d rows, want at least 8", banner_file_name, character_rune, len(block_rows))
			}
		}
	}
}

// ===================  Render  ===================

// TestRender_ProducesEightRowsPerLine checks the exact output for a known input,
// pinning both the row order and the left-to-right character order.
func TestRender_ProducesEightRowsPerLine(t *testing.T) {
	rendered_output := Render([]string{"AB"}, build_test_banner_map())

	expected_output := "A0B0\nA1B1\nA2B2\nA3B3\nA4B4\nA5B5\nA6B6\nA7B7\n"
	if rendered_output != expected_output {
		t.Errorf("render output:\ngot  %q\nwant %q", rendered_output, expected_output)
	}
}

// TestRender_EmptyLineBecomesBlankLine checks that a blank input line produces a
// single blank output line rather than eight of them. This is how a paragraph
// break the user typed survives into the art.
func TestRender_EmptyLineBecomesBlankLine(t *testing.T) {
	rendered_output := Render([]string{"A", "", "A"}, build_test_banner_map())

	expected_output := "A0\nA1\nA2\nA3\nA4\nA5\nA6\nA7\n" +
		"\n" +
		"A0\nA1\nA2\nA3\nA4\nA5\nA6\nA7\n"
	if rendered_output != expected_output {
		t.Errorf("render output:\ngot  %q\nwant %q", rendered_output, expected_output)
	}
}

// TestRender_SkipsCharactersMissingFromBanner checks that a character with no
// entry in the banner map is dropped instead of panicking on the lookup.
func TestRender_SkipsCharactersMissingFromBanner(t *testing.T) {
	rendered_output := Render([]string{"AZB"}, build_test_banner_map())

	// 'Z' is not in the test banner map, so the output is the same as for "AB".
	expected_output := "A0B0\nA1B1\nA2B2\nA3B3\nA4B4\nA5B5\nA6B6\nA7B7\n"
	if rendered_output != expected_output {
		t.Errorf("render output:\ngot  %q\nwant %q", rendered_output, expected_output)
	}
}

// TestRender_NoInputProducesNoOutput checks the degenerate case of an empty
// slice, which is what a caller passing no lines at all would produce.
func TestRender_NoInputProducesNoOutput(t *testing.T) {
	rendered_output := Render([]string{}, build_test_banner_map())

	if rendered_output != "" {
		t.Errorf("render output for no input: got %q, want empty", rendered_output)
	}
}

// ===================  RenderWithColor  ===================

// TestRenderWithColor_WrapsEachRowInArtLineSpan checks that every row is wrapped
// in the class-based span. The class is what picks up the color from CSS.
func TestRenderWithColor_WrapsEachRowInArtLineSpan(t *testing.T) {
	rendered_output := RenderWithColor([]string{"A"}, build_test_banner_map(), "#ff0000")

	span_count := strings.Count(rendered_output, `<span class="art-line">`)
	if span_count != 8 {
		t.Errorf("span count: got %d, want 8", span_count)
	}

	if !strings.Contains(rendered_output, `<span class="art-line">A0</span>`) {
		t.Errorf("first row is not wrapped as expected: %q", rendered_output)
	}
}

// TestRenderWithColor_DoesNotEmitInlineStyle checks that the color is never
// written into the markup as a style attribute. An inline style would force the
// Content-Security-Policy to allow 'unsafe-inline' styles again, which is
// exactly what the class-based approach exists to avoid.
func TestRenderWithColor_DoesNotEmitInlineStyle(t *testing.T) {
	rendered_output := RenderWithColor([]string{"A"}, build_test_banner_map(), "#ff0000")

	if strings.Contains(rendered_output, "style=") {
		t.Errorf("output carries an inline style attribute: %q", rendered_output)
	}

	if strings.Contains(rendered_output, "#ff0000") {
		t.Errorf("the color was written into the markup: %q", rendered_output)
	}
}

// TestRenderWithColor_EscapesMarkupCharacters checks that glyph characters which
// are also HTML metacharacters are escaped. Several real banner glyphs (B, K, X,
// k, x, < and &) contain a literal '<'; unescaped, the browser parses them as
// tags and swallows part of the art.
func TestRenderWithColor_EscapesMarkupCharacters(t *testing.T) {
	// A banner map whose glyph rows are made of markup characters.
	markup_banner_map := map[rune][]string{
		'A': {"<&>", "<&>", "<&>", "<&>", "<&>", "<&>", "<&>", "<&>"},
	}

	rendered_output := RenderWithColor([]string{"A"}, markup_banner_map, "#ff0000")

	if !strings.Contains(rendered_output, "&lt;&amp;&gt;") {
		t.Errorf("glyph markup characters were not escaped: %q", rendered_output)
	}

	// The only raw '<' left must be the ones opening the wrapper spans: eight
	// opening tags plus eight closing tags.
	raw_angle_bracket_count := strings.Count(rendered_output, "<")
	if raw_angle_bracket_count != 16 {
		t.Errorf("raw '<' count: got %d, want 16 (span tags only)", raw_angle_bracket_count)
	}
}

// TestRenderWithColor_EmptyLineBecomesBlankLine checks that a blank input line
// produces a bare newline and not an empty span, matching Render's behaviour.
func TestRenderWithColor_EmptyLineBecomesBlankLine(t *testing.T) {
	rendered_output := RenderWithColor([]string{"", "A"}, build_test_banner_map(), "#ff0000")

	if !strings.HasPrefix(rendered_output, "\n<span") {
		t.Errorf("blank line did not produce a bare newline: %q", rendered_output)
	}

	span_count := strings.Count(rendered_output, `<span class="art-line">`)
	if span_count != 8 {
		t.Errorf("span count: got %d, want 8 (the blank line must not emit a span)", span_count)
	}
}
