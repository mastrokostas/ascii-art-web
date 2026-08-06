package handlers

import (
	"errors"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain runs before all tests in this package.
// It changes the working directory to the project root so that
// handlers can resolve relative paths like "templates/" and "banners/".
func TestMain(m *testing.M) {
	os.Chdir("../..")
	os.Exit(m.Run())
}

// TestHomeHandler_Status200 checks that GET / returns 200 OK.
func TestHomeHandler_Status200(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(HomeHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("HomeHandler: got %v, want %v", status, http.StatusOK)
	}
}

// TestHomeHandler_UnknownPath checks that any path other than "/" returns 404.
func TestHomeHandler_UnknownPath(t *testing.T) {
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(HomeHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("HomeHandler unknown path: got %v, want %v", status, http.StatusNotFound)
	}
}

// TestAsciiArtHandler_RejectGET checks that GET /ascii-art returns 400 Bad Request.
func TestAsciiArtHandler_RejectGET(t *testing.T) {
	req, err := http.NewRequest("GET", "/ascii-art", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("AsciiArtHandler GET: got %v, want %v", status, http.StatusBadRequest)
	}
}

// TestAsciiArtHandler_EmptyText checks that a POST with empty text returns 400 Bad Request.
func TestAsciiArtHandler_EmptyText(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "")
	formData.Set("banner", "standard")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("AsciiArtHandler empty text: got %v, want %v", status, http.StatusBadRequest)
	}
}

// TestAsciiArtHandler_InvalidBanner checks that a POST with an unknown banner returns 400 Bad Request.
func TestAsciiArtHandler_InvalidBanner(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "Hello")
	formData.Set("banner", "nonexistent")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("AsciiArtHandler invalid banner: got %v, want %v", status, http.StatusBadRequest)
	}
}

// TestAsciiArtHandler_ValidPOST checks that a valid POST returns 200 OK.
func TestAsciiArtHandler_ValidPOST(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "Hello")
	formData.Set("banner", "standard")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("AsciiArtHandler valid POST: got %v, want %v", status, http.StatusOK)
	}
}

// TestAsciiArtHandler_ValidColor checks that a POST with a valid hex color returns 200 OK.
func TestAsciiArtHandler_ValidColor(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "Hello")
	formData.Set("banner", "standard")
	formData.Set("color", "#ff0000")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("AsciiArtHandler valid color: got %v, want %v", status, http.StatusOK)
	}
}

// TestAsciiArtHandler_InvalidColor checks that a POST with an invalid hex color returns 400 Bad Request.
func TestAsciiArtHandler_InvalidColor(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "Hello")
	formData.Set("banner", "standard")
	formData.Set("color", "notacolor")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("AsciiArtHandler invalid color: got %v, want %v", status, http.StatusBadRequest)
	}
}

// TestAsciiArtHandler_NoColor checks that a POST with no color returns 200 OK and renders plain.
func TestAsciiArtHandler_NoColor(t *testing.T) {
	formData := url.Values{}
	formData.Set("text", "Hello")
	formData.Set("banner", "standard")
	req, err := http.NewRequest("POST", "/ascii-art", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("AsciiArtHandler no color: got %v, want %v", status, http.StatusOK)
	}
}

// ===================  Download: format selection  ===================

// postForm builds a form-encoded POST request to the given path.
func postForm(t *testing.T, path string, form url.Values) *http.Request {
	t.Helper()
	req, err := http.NewRequest("POST", path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// TestDownloadHandler_DefaultsToTxt checks that a POST with no format field
// still returns the plain-text .txt download (backwards compatible).
func TestDownloadHandler_DefaultsToTxt(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "hello")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Errorf("Content-Type: got %q, want text/plain", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "ascii-art.txt") {
		t.Errorf("Content-Disposition: got %q, want ascii-art.txt", cd)
	}
	if body := rr.Body.String(); body != "hello" {
		t.Errorf("body: got %q, want %q", body, "hello")
	}
}

// TestDownloadHandler_HtmlFormat checks that format=html returns a self-contained
// HTML document as an .html attachment.
func TestDownloadHandler_HtmlFormat(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "AB")
	form.Set("format", "html")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "ascii-art.html") {
		t.Errorf("Content-Disposition: got %q, want ascii-art.html", cd)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("body missing doctype: %q", body)
	}
	if !strings.Contains(body, "<pre") {
		t.Errorf("body missing <pre>: %q", body)
	}
	if !strings.Contains(body, "AB") {
		t.Errorf("body missing art content: %q", body)
	}
}

// TestDownloadHandler_HtmlAppliesColor checks that a validated color is applied
// to the exported art via an inline style.
func TestDownloadHandler_HtmlAppliesColor(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "AB")
	form.Set("format", "html")
	form.Set("color", "#ff0000")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if body := rr.Body.String(); !strings.Contains(body, "color: #ff0000") {
		t.Errorf("body missing color style: %q", body)
	}
}

// TestDownloadHandler_HtmlEscapesContent checks that HTML metacharacters in the
// art are escaped, so the exported file cannot contain injected markup.
func TestDownloadHandler_HtmlEscapesContent(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "a<b>c")
	form.Set("format", "html")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	body := rr.Body.String()
	if !strings.Contains(body, "a&lt;b&gt;c") {
		t.Errorf("content not escaped: %q", body)
	}
	if strings.Contains(body, "a<b>c") {
		t.Errorf("raw unescaped content present: %q", body)
	}
}

// TestDownloadHandler_HtmlIgnoresInvalidColor checks that an invalid color is not
// written into the document — the export falls back to plain, never emitting an
// unvalidated value into the style attribute.
func TestDownloadHandler_HtmlIgnoresInvalidColor(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "AB")
	form.Set("format", "html")
	form.Set("color", "bogus")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", ct)
	}
	if body := rr.Body.String(); strings.Contains(body, "bogus") {
		t.Errorf("unvalidated color leaked into document: %q", body)
	}
}

// extractDownloadPayload pulls the text inside the hidden asciiText textarea of
// the rendered index page.
func extractDownloadPayload(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, `name="asciiText"`)
	if i < 0 {
		t.Fatal("download textarea not found")
	}
	gt := strings.Index(body[i:], ">")
	if gt < 0 {
		t.Fatal("malformed textarea tag")
	}
	start := i + gt + 1
	end := strings.Index(body[start:], "</textarea>")
	if end < 0 {
		t.Fatal("textarea not closed")
	}
	return body[start : start+end]
}

// TestAsciiArtHandler_DownloadPayloadIsPlainForColor checks that the hidden
// download payload on the full-page (no-JS) render is always the plain text,
// even when the visible art is colorized. Otherwise "Download .txt" would save
// raw <span> markup.
func TestAsciiArtHandler_DownloadPayloadIsPlainForColor(t *testing.T) {
	form := url.Values{}
	form.Set("text", "Hi")
	form.Set("banner", "standard")
	form.Set("color", "#ff0000")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}
	payload := extractDownloadPayload(t, rr.Body.String())
	if strings.Contains(payload, "span") {
		t.Errorf("download payload contains span markup, want plain text: %q", payload)
	}
}

// TestDownloadHandler_HtmlColorHasContrastingPlate checks that a colored HTML
// export sits on a plate derived from the color (mirroring the result card),
// so a dark color stays readable on its backing.
func TestDownloadHandler_HtmlColorHasContrastingPlate(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "AB")
	form.Set("format", "html")
	form.Set("color", "#ff0000")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if body := rr.Body.String(); !strings.Contains(body, "background-color: oklch(from #ff0000") {
		t.Errorf("colored export missing contrasting plate: %q", body)
	}
}

// TestDownloadHandler_HtmlPlainHasNoPlate checks that an uncolored HTML export
// has no derived plate — plain light-on-dark, like the card with no color.
func TestDownloadHandler_HtmlPlainHasNoPlate(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "AB")
	form.Set("format", "html")
	rr := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(rr, postForm(t, "/download", form))

	if body := rr.Body.String(); strings.Contains(body, "oklch") {
		t.Errorf("plain export should not have a plate: %q", body)
	}
}

// ===================  Output escaping and input limits  ===================

// extractRenderedArt pulls the contents of the <pre class="result"> block from
// the rendered index page — the art exactly as the browser would receive it.
func extractRenderedArt(t *testing.T, body string) string {
	t.Helper()
	open := strings.Index(body, `<pre class="result">`)
	if open < 0 {
		t.Fatal("result <pre> not found")
	}
	start := open + len(`<pre class="result">`)
	end := strings.Index(body[start:], "</pre>")
	if end < 0 {
		t.Fatal("result <pre> not closed")
	}
	return body[start : start+end]
}

// TestAsciiArtHandler_PlainResultIsEscaped checks the full-page (no-JavaScript)
// render escapes plain art. Several standard-banner glyphs — B, K, X, k, x, <
// and & — contain a literal '<'. Result is declared template.HTML, so nothing
// escapes it downstream; if the handler stops escaping, the browser parses those
// glyphs as tags and silently swallows part of the art.
func TestAsciiArtHandler_PlainResultIsEscaped(t *testing.T) {
	form := url.Values{}
	form.Set("text", "X")
	form.Set("banner", "standard")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}

	art := extractRenderedArt(t, rr.Body.String())
	if !strings.Contains(art, "&lt;") {
		t.Errorf("expected escaped '<' in plain art, got: %q", art)
	}
	if strings.Contains(art, "<") {
		t.Errorf("raw '<' reached the page and will be parsed as a tag: %q", art)
	}
}

// TestAsciiArtHandler_ColoredResultUsesClass checks that colored art carries the
// color through the .art-line class rather than an inline style attribute. An
// inline style would force the Content-Security-Policy back to 'unsafe-inline'.
func TestAsciiArtHandler_ColoredResultUsesClass(t *testing.T) {
	form := url.Values{}
	form.Set("text", "X")
	form.Set("banner", "standard")
	form.Set("color", "#ff0000")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `<span class="art-line">`) {
		t.Errorf("colored art missing .art-line spans: %q", extractRenderedArt(t, body))
	}
	if strings.Contains(body, `style="color:`) {
		t.Errorf("inline color style present, CSP would need 'unsafe-inline': %q", body)
	}
}

// TestAsciiArtHandler_DownloadPayloadKeepsRawGlyphs guards the escaping fix from
// being "simplified" into ascii.Render itself. The same render feeds the hidden
// download field, and the downloaded .txt must contain the real '<' glyph rather
// than the entity &lt;. The template escapes the textarea body on the way out, so
// unescape before asserting — that is what the browser submits back.
func TestAsciiArtHandler_DownloadPayloadKeepsRawGlyphs(t *testing.T) {
	form := url.Values{}
	form.Set("text", "X")
	form.Set("banner", "standard")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", rr.Code)
	}

	payload := html.UnescapeString(extractDownloadPayload(t, rr.Body.String()))
	if !strings.Contains(payload, "<") {
		t.Errorf("download payload lost its literal '<' glyph — art must not be escaped: %q", payload)
	}
}

// TestAsciiArtHandler_RejectsOversizedText checks that text beyond maxTextLength
// is refused. The renderer expands every character into eight rows of glyph, so
// an unbounded field is a memory amplifier.
func TestAsciiArtHandler_RejectsOversizedText(t *testing.T) {
	form := url.Values{}
	form.Set("text", strings.Repeat("A", maxTextLength+1))
	form.Set("banner", "standard")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized text: got %v, want %v", rr.Code, http.StatusRequestEntityTooLarge)
	}
}

// TestAsciiArtHandler_AcceptsMaxLengthText checks the boundary from the other
// side, so the limit cannot silently drift down and start rejecting valid input.
func TestAsciiArtHandler_AcceptsMaxLengthText(t *testing.T) {
	form := url.Values{}
	form.Set("text", strings.Repeat("A", maxTextLength))
	form.Set("banner", "standard")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code != http.StatusOK {
		t.Errorf("text at the limit: got %v, want 200", rr.Code)
	}
}

// TestAsciiArtHandler_RejectsOversizedBody checks that a body larger than
// maxRequestBodyBytes is cut off by MaxBytesReader before any form parsing, so
// the request fails rather than being read into memory in full.
func TestAsciiArtHandler_RejectsOversizedBody(t *testing.T) {
	form := url.Values{}
	form.Set("text", strings.Repeat("A", maxRequestBodyBytes+1))
	form.Set("banner", "standard")
	rr := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(rr, postForm(t, "/ascii-art", form))

	if rr.Code == http.StatusOK {
		t.Errorf("oversized body was accepted: got %v, want a failure status", rr.Code)
	}
}

// ===================  Live preview (the fetch path)  ===================

// post_preview_form builds the same form POST as postForm but marks it as coming
// from the frontend's live-preview fetch(), which is what makes the handler
// answer with a bare art fragment instead of a full HTML page.
func post_preview_form(t *testing.T, path string, form url.Values) *http.Request {
	t.Helper()

	preview_request := postForm(t, path, form)
	preview_request.Header.Set("X-Requested-With", "fetch")

	return preview_request
}

// TestAsciiArtHandler_PreviewEmptyTextIsNotAnError checks that an empty textarea
// on the live-preview path means "clear the output", not "bad request". A 400
// here would make the preview flash an error every time the user erases the box.
func TestAsciiArtHandler_PreviewEmptyTextIsNotAnError(t *testing.T) {
	form := url.Values{}
	form.Set("text", "")
	form.Set("banner", "standard")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, post_preview_form(t, "/ascii-art", form))

	if recorder.Code != http.StatusOK {
		t.Errorf("status: got %v, want 200", recorder.Code)
	}

	// The frontend removes the result area when the body comes back empty.
	if recorder.Body.Len() != 0 {
		t.Errorf("body: got %q, want empty", recorder.Body.String())
	}
}

// TestAsciiArtHandler_PreviewPlainReturnsTextFragment checks that a preview
// without a color returns the bare rendered art as text/plain — no page chrome,
// and no HTML escaping, because the frontend reads it with response.text().
func TestAsciiArtHandler_PreviewPlainReturnsTextFragment(t *testing.T) {
	form := url.Values{}
	form.Set("text", "X")
	form.Set("banner", "standard")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, post_preview_form(t, "/ascii-art", form))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", recorder.Code)
	}

	if content_type := recorder.Header().Get("Content-Type"); !strings.Contains(content_type, "text/plain") {
		t.Errorf("Content-Type: got %q, want text/plain", content_type)
	}

	response_body := recorder.Body.String()
	if strings.Contains(response_body, "<!DOCTYPE") {
		t.Errorf("preview returned a full page instead of a fragment: %q", response_body)
	}

	// The 'X' glyph of the standard banner is drawn with literal '<' characters.
	// On this path they must stay literal — the frontend inserts them as text.
	if !strings.Contains(response_body, "<") {
		t.Errorf("plain preview lost its literal '<' glyph: %q", response_body)
	}

	if strings.Contains(response_body, "&lt;") {
		t.Errorf("plain preview was HTML-escaped: %q", response_body)
	}
}

// TestAsciiArtHandler_PreviewColoredReturnsHtmlFragment checks that a preview
// with a color returns the span-wrapped markup as text/html, again with no page
// chrome around it.
func TestAsciiArtHandler_PreviewColoredReturnsHtmlFragment(t *testing.T) {
	form := url.Values{}
	form.Set("text", "X")
	form.Set("banner", "standard")
	form.Set("color", "#ff0000")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, post_preview_form(t, "/ascii-art", form))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status: got %v, want 200", recorder.Code)
	}

	if content_type := recorder.Header().Get("Content-Type"); !strings.Contains(content_type, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", content_type)
	}

	response_body := recorder.Body.String()
	if !strings.Contains(response_body, `<span class="art-line">`) {
		t.Errorf("colored preview is missing its .art-line spans: %q", response_body)
	}

	if strings.Contains(response_body, "<!DOCTYPE") {
		t.Errorf("preview returned a full page instead of a fragment: %q", response_body)
	}
}

// TestAsciiArtHandler_PreviewStillValidatesInput checks that the preview path
// does not become a way around the input checks: the shorter fetch response must
// not mean weaker validation.
func TestAsciiArtHandler_PreviewStillValidatesInput(t *testing.T) {
	invalid_submissions := []struct {
		case_name       string
		form_values     url.Values
		expected_status int
	}{
		{
			case_name:       "unknown banner",
			form_values:     url.Values{"text": {"Hi"}, "banner": {"nonexistent"}},
			expected_status: http.StatusBadRequest,
		},
		{
			case_name:       "invalid color",
			form_values:     url.Values{"text": {"Hi"}, "banner": {"standard"}, "color": {"notacolor"}},
			expected_status: http.StatusBadRequest,
		},
		{
			case_name:       "text beyond the length limit",
			form_values:     url.Values{"text": {strings.Repeat("A", maxTextLength+1)}, "banner": {"standard"}},
			expected_status: http.StatusRequestEntityTooLarge,
		},
	}

	for _, submission := range invalid_submissions {
		recorder := httptest.NewRecorder()
		http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, post_preview_form(t, "/ascii-art", submission.form_values))

		if recorder.Code != submission.expected_status {
			t.Errorf("%s: got %v, want %v", submission.case_name, recorder.Code, submission.expected_status)
		}
	}
}

// ===================  Download: rejected requests  ===================

// TestDownloadHandler_RejectsGET checks that GET /download is a bad request —
// the endpoint only exists to receive a posted payload.
func TestDownloadHandler_RejectsGET(t *testing.T) {
	recorder := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(recorder, httptest.NewRequest("GET", "/download", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status: got %v, want 400", recorder.Code)
	}
}

// TestDownloadHandler_RejectsEmptyContent checks that a POST with nothing to
// download is refused, rather than producing an empty attachment.
func TestDownloadHandler_RejectsEmptyContent(t *testing.T) {
	form := url.Values{}
	form.Set("asciiText", "")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(DownloadHandler).ServeHTTP(recorder, postForm(t, "/download", form))

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status: got %v, want 400", recorder.Code)
	}

	if content_disposition := recorder.Header().Get("Content-Disposition"); content_disposition != "" {
		t.Errorf("an attachment was offered for empty content: %q", content_disposition)
	}
}

// ===================  Missing files on disk  ===================
//
// The handlers resolve "templates/" and "banners/" relative to the working
// directory, which TestMain points at the project root. The tests below use
// t.Chdir to move into an empty temporary directory so those lookups fail on
// purpose; t.Chdir restores the previous directory when the test ends and
// refuses to run in a parallel test, so the switch cannot leak into the others.

// TestHomeHandler_MissingTemplateFallsBack checks the two-stage failure: the
// index template cannot be parsed, and the error page used to report that is
// missing as well, so RenderError falls back to a plain http.Error. The status
// must still be the one the handler chose.
func TestHomeHandler_MissingTemplateFallsBack(t *testing.T) {
	t.Chdir(t.TempDir())

	recorder := httptest.NewRecorder()
	http.HandlerFunc(HomeHandler).ServeHTTP(recorder, httptest.NewRequest("GET", "/", nil))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Template not found") {
		t.Errorf("body: got %q, want the fallback message", recorder.Body.String())
	}
}

// TestRenderError_FallsBackWithoutErrorTemplate checks the fallback in isolation,
// including that the status code is written exactly once. RenderError parses the
// template before writing any status precisely so this path does not end up
// sending the header twice.
func TestRenderError_FallsBackWithoutErrorTemplate(t *testing.T) {
	t.Chdir(t.TempDir())

	recorder := httptest.NewRecorder()
	RenderError(recorder, http.StatusInternalServerError, "Internal Server Error")

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status: got %v, want 500", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Internal Server Error") {
		t.Errorf("body: got %q, want the message", recorder.Body.String())
	}
}

// TestRenderError_UsesErrorTemplateWhenPresent checks the ordinary path, so the
// fallback test above cannot pass simply because the template is never used.
func TestRenderError_UsesErrorTemplateWhenPresent(t *testing.T) {
	recorder := httptest.NewRecorder()
	RenderError(recorder, http.StatusNotFound, "Page not found")

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	response_body := recorder.Body.String()
	if !strings.Contains(response_body, "404") {
		t.Errorf("body does not carry the status code: %q", response_body)
	}

	if !strings.Contains(response_body, "Page not found") {
		t.Errorf("body does not carry the message: %q", response_body)
	}
}

// TestAsciiArtHandler_MissingBannerFileIsNotFound checks that a banner name that
// passes validation but has no file behind it is reported as a 404 rather than a
// 500 — the handler distinguishes the two by inspecting the read error.
func TestAsciiArtHandler_MissingBannerFileIsNotFound(t *testing.T) {
	t.Chdir(t.TempDir())

	form := url.Values{}
	form.Set("text", "Hi")
	form.Set("banner", "standard")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, postForm(t, "/ascii-art", form))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Banner not found") {
		t.Errorf("body: got %q, want the banner-not-found message", recorder.Body.String())
	}
}

// TestAsciiArtHandler_UnreadableBannerIsServerError checks the other side of that
// split: a banner path that exists but cannot be read is a server-side fault, so
// it must be a 500 and not a 404. A directory standing where the file belongs
// produces exactly that kind of read error.
func TestAsciiArtHandler_UnreadableBannerIsServerError(t *testing.T) {
	temporary_root := t.TempDir()

	// banners/standard.txt exists, but as a directory — opening it succeeds while
	// reading it fails with an error that is not "does not exist".
	unreadable_banner_path := filepath.Join(temporary_root, "banners", "standard.txt")
	if mkdir_error := os.MkdirAll(unreadable_banner_path, 0o755); mkdir_error != nil {
		t.Fatalf("building the unreadable banner: %v", mkdir_error)
	}

	t.Chdir(temporary_root)

	form := url.Values{}
	form.Set("text", "Hi")
	form.Set("banner", "standard")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, postForm(t, "/ascii-art", form))

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status: got %v, want 500", recorder.Code)
	}
}

// TestAsciiArtHandler_MissingTemplateIsReported checks the full-page path when
// the banner loads but templates/index.html does not exist. The banner is copied
// into the temporary directory so the render gets past the earlier failure and
// actually reaches the template parse.
func TestAsciiArtHandler_MissingTemplateIsReported(t *testing.T) {
	// Read the real banner before changing directory, while the relative path
	// still resolves against the project root.
	banner_file_contents, read_error := os.ReadFile("banners/standard.txt")
	if read_error != nil {
		t.Fatalf("reading the standard banner: %v", read_error)
	}

	temporary_root := t.TempDir()
	banner_directory := filepath.Join(temporary_root, "banners")
	if mkdir_error := os.MkdirAll(banner_directory, 0o755); mkdir_error != nil {
		t.Fatalf("creating the banner directory: %v", mkdir_error)
	}

	banner_copy_path := filepath.Join(banner_directory, "standard.txt")
	if write_error := os.WriteFile(banner_copy_path, banner_file_contents, 0o644); write_error != nil {
		t.Fatalf("copying the standard banner: %v", write_error)
	}

	// The banners are in place but templates/ is not, so ParseFiles is what fails.
	t.Chdir(temporary_root)

	form := url.Values{}
	form.Set("text", "Hi")
	form.Set("banner", "standard")
	recorder := httptest.NewRecorder()
	http.HandlerFunc(AsciiArtHandler).ServeHTTP(recorder, postForm(t, "/ascii-art", form))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Template not found") {
		t.Errorf("body: got %q, want the template-not-found message", recorder.Body.String())
	}
}

// ===================  Failures while writing the response  ===================

// failing_response_writer is an http.ResponseWriter whose every write fails, the
// way a client that hung up mid-response behaves. It records the status codes it
// is handed so a test can see which one the handler settled on.
type failing_response_writer struct {
	response_headers      http.Header
	recorded_status_codes []int
}

// new_failing_response_writer returns a writer with its header map initialised.
func new_failing_response_writer() *failing_response_writer {
	return &failing_response_writer{response_headers: make(http.Header)}
}

func (writer *failing_response_writer) Header() http.Header {
	return writer.response_headers
}

// Write always fails, which is what drives template execution into its error
// return.
func (writer *failing_response_writer) Write(data []byte) (int, error) {
	return 0, errors.New("client went away")
}

func (writer *failing_response_writer) WriteHeader(status_code int) {
	writer.recorded_status_codes = append(writer.recorded_status_codes, status_code)
}

// last_recorded_status returns the final status the handler wrote, or zero when
// it never wrote one.
func (writer *failing_response_writer) last_recorded_status() int {
	if len(writer.recorded_status_codes) == 0 {
		return 0
	}

	return writer.recorded_status_codes[len(writer.recorded_status_codes)-1]
}

// TestHomeHandler_TemplateExecutionFailure checks that a write failure part-way
// through rendering the home page is turned into a 500 rather than being
// swallowed.
func TestHomeHandler_TemplateExecutionFailure(t *testing.T) {
	failing_writer := new_failing_response_writer()

	HomeHandler(failing_writer, httptest.NewRequest("GET", "/", nil))

	if failing_writer.last_recorded_status() != http.StatusInternalServerError {
		t.Errorf("status: got %v, want 500", failing_writer.last_recorded_status())
	}
}

// TestAsciiArtHandler_TemplateExecutionFailure checks the same failure on the
// full-page render of a valid submission.
func TestAsciiArtHandler_TemplateExecutionFailure(t *testing.T) {
	form := url.Values{}
	form.Set("text", "Hi")
	form.Set("banner", "standard")

	failing_writer := new_failing_response_writer()

	AsciiArtHandler(failing_writer, postForm(t, "/ascii-art", form))

	if failing_writer.last_recorded_status() != http.StatusInternalServerError {
		t.Errorf("status: got %v, want 500", failing_writer.last_recorded_status())
	}
}
