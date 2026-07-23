package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
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
