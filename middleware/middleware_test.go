package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serveThroughSecureHeaders runs a trivial handler behind SecureHeaders and
// returns both the recorded response and the nonce the handler saw on its
// request context.
func serveThroughSecureHeaders(t *testing.T) (*httptest.ResponseRecorder, string) {
	t.Helper()

	var observed_nonce string
	inner_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed_nonce = NonceFromContext(r.Context())
	})

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	SecureHeaders(inner_handler).ServeHTTP(recorder, request)

	return recorder, observed_nonce
}

// TestSecureHeaders_CspHasNonceNotUnsafeInline checks that style-src is carried
// by a nonce rather than 'unsafe-inline'. This is the whole point of moving the
// art color out of inline style attributes, so it is worth pinning down.
func TestSecureHeaders_CspHasNonceNotUnsafeInline(t *testing.T) {
	recorder, _ := serveThroughSecureHeaders(t)

	policy := recorder.Header().Get("Content-Security-Policy")
	if policy == "" {
		t.Fatal("no Content-Security-Policy header was set")
	}
	if strings.Contains(policy, "unsafe-inline") {
		t.Errorf("policy still allows inline styles: %q", policy)
	}
	if !strings.Contains(policy, "style-src 'self' 'nonce-") {
		t.Errorf("policy is missing a style-src nonce: %q", policy)
	}
}

// TestSecureHeaders_NonceReachesHandler checks that the nonce advertised in the
// policy is the same one handlers read from the request context. If these ever
// diverge the inline <style> block is silently blocked by the browser.
func TestSecureHeaders_NonceReachesHandler(t *testing.T) {
	recorder, observed_nonce := serveThroughSecureHeaders(t)

	if observed_nonce == "" {
		t.Fatal("handler saw no nonce on the request context")
	}

	policy := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(policy, "'nonce-"+observed_nonce+"'") {
		t.Errorf("nonce %q from the context is not the one in the policy: %q", observed_nonce, policy)
	}
}

// TestSecureHeaders_NonceIsPerRequest checks that a fresh nonce is minted for
// every request. A reused nonce would be worthless — injected markup could just
// quote it.
func TestSecureHeaders_NonceIsPerRequest(t *testing.T) {
	_, first_nonce := serveThroughSecureHeaders(t)
	_, second_nonce := serveThroughSecureHeaders(t)

	if first_nonce == second_nonce {
		t.Errorf("nonce was reused across requests: %q", first_nonce)
	}
}

// TestSecureHeaders_HstsIsOptIn checks that Strict-Transport-Security stays off
// unless it is explicitly enabled. Sending it from a plain-HTTP development
// origin would pin that origin to HTTPS in the developer's browser.
func TestSecureHeaders_HstsIsOptIn(t *testing.T) {
	recorder, _ := serveThroughSecureHeaders(t)

	if header := recorder.Header().Get("Strict-Transport-Security"); header != "" {
		t.Errorf("HSTS was sent without ENABLE_HSTS being set: %q", header)
	}
}

// TestSecureHeaders_HstsWhenEnabled checks the opposite side of that switch.
func TestSecureHeaders_HstsWhenEnabled(t *testing.T) {
	t.Setenv("ENABLE_HSTS", "true")

	recorder, _ := serveThroughSecureHeaders(t)

	header := recorder.Header().Get("Strict-Transport-Security")
	if !strings.Contains(header, "max-age=31536000") {
		t.Errorf("HSTS not set with ENABLE_HSTS=true: %q", header)
	}
}

// TestSecureHeaders_SetsHardeningHeaders checks the fixed set of hardening
// headers that do not depend on any environment switch.
func TestSecureHeaders_SetsHardeningHeaders(t *testing.T) {
	recorder, _ := serveThroughSecureHeaders(t)

	expected_headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	for header_name, expected_value := range expected_headers {
		if actual_value := recorder.Header().Get(header_name); actual_value != expected_value {
			t.Errorf("%s: got %q, want %q", header_name, actual_value, expected_value)
		}
	}
}

// TestNonceFromContext_EmptyWithoutMiddleware checks the fallback path: a request
// that never passed through SecureHeaders has no nonce on its context, and the
// lookup returns an empty string rather than panicking on the type assertion.
func TestNonceFromContext_EmptyWithoutMiddleware(t *testing.T) {
	bare_request := httptest.NewRequest("GET", "/", nil)

	if nonce := NonceFromContext(bare_request.Context()); nonce != "" {
		t.Errorf("nonce without the middleware: got %q, want empty", nonce)
	}
}

// ===================  NoListFileSystem  ===================

// buildFileTree creates a temporary directory tree for the file system tests and
// returns its root:
//
//	root/visible.txt          a plain file that must always be servable
//	root/listable/note.txt    a directory with no index.html — must be hidden
//	root/indexed/index.html   a directory with an index.html — must be servable
func buildFileTree(t *testing.T) string {
	t.Helper()

	root_directory := t.TempDir()

	write_file := func(relative_path string, contents string) {
		full_path := filepath.Join(root_directory, relative_path)

		if mkdir_error := os.MkdirAll(filepath.Dir(full_path), 0o755); mkdir_error != nil {
			t.Fatalf("creating directory for %q: %v", relative_path, mkdir_error)
		}

		if write_error := os.WriteFile(full_path, []byte(contents), 0o644); write_error != nil {
			t.Fatalf("writing %q: %v", relative_path, write_error)
		}
	}

	write_file("visible.txt", "visible file contents")
	write_file("listable/note.txt", "a file nobody should be able to discover by browsing")
	write_file("indexed/index.html", "<h1>index</h1>")

	return root_directory
}

// TestNoListFileSystem_HidesDirectoryWithoutIndex checks the whole point of the
// wrapper: a directory that has no index.html is reported as missing, so
// http.FileServer renders a 404 instead of listing the raw file tree.
func TestNoListFileSystem_HidesDirectoryWithoutIndex(t *testing.T) {
	no_list_file_system := NoListFileSystem{Fs: http.Dir(buildFileTree(t))}

	opened_file, open_error := no_list_file_system.Open("/listable/")
	if open_error == nil {
		opened_file.Close()
		t.Fatal("a directory without index.html was opened, the listing would be served")
	}

	if !errors.Is(open_error, os.ErrNotExist) {
		t.Errorf("error is not os.ErrNotExist, so it will not become a 404: %v", open_error)
	}
}

// TestNoListFileSystem_ServesDirectoryWithIndex checks the other half of the
// rule: a directory holding an index.html is still served normally.
func TestNoListFileSystem_ServesDirectoryWithIndex(t *testing.T) {
	no_list_file_system := NoListFileSystem{Fs: http.Dir(buildFileTree(t))}

	opened_file, open_error := no_list_file_system.Open("/indexed/")
	if open_error != nil {
		t.Fatalf("a directory with index.html was refused: %v", open_error)
	}
	defer opened_file.Close()

	file_info, stat_error := opened_file.Stat()
	if stat_error != nil {
		t.Fatalf("stat on the opened directory failed: %v", stat_error)
	}

	if !file_info.IsDir() {
		t.Error("expected the opened entry to be a directory")
	}
}

// TestNoListFileSystem_ServesRegularFile checks that ordinary files pass through
// untouched — the wrapper must only affect directories.
func TestNoListFileSystem_ServesRegularFile(t *testing.T) {
	no_list_file_system := NoListFileSystem{Fs: http.Dir(buildFileTree(t))}

	opened_file, open_error := no_list_file_system.Open("/visible.txt")
	if open_error != nil {
		t.Fatalf("a regular file was refused: %v", open_error)
	}
	defer opened_file.Close()

	file_contents := make([]byte, len("visible file contents"))
	if _, read_error := opened_file.Read(file_contents); read_error != nil {
		t.Fatalf("reading the opened file failed: %v", read_error)
	}

	if string(file_contents) != "visible file contents" {
		t.Errorf("file contents: got %q, want %q", file_contents, "visible file contents")
	}
}

// TestNoListFileSystem_PropagatesMissingFileError checks that a request for
// something that simply is not there still fails, rather than being masked.
func TestNoListFileSystem_PropagatesMissingFileError(t *testing.T) {
	no_list_file_system := NoListFileSystem{Fs: http.Dir(buildFileTree(t))}

	opened_file, open_error := no_list_file_system.Open("/no-such-file.txt")
	if open_error == nil {
		opened_file.Close()
		t.Fatal("opening a missing file returned no error")
	}

	if !errors.Is(open_error, os.ErrNotExist) {
		t.Errorf("error is not os.ErrNotExist: %v", open_error)
	}
}

// unstattable_file is an http.File that opens successfully but cannot be
// inspected. A real disk file rarely behaves this way, so it is faked here to
// reach the stat-failure branch of NoListFileSystem.Open.
type unstattable_file struct {
	http.File
	// was_closed records whether Open released the handle before returning the
	// error, which is the leak this test is really watching for.
	was_closed *bool
}

func (file unstattable_file) Stat() (os.FileInfo, error) {
	return nil, errors.New("stat is unavailable")
}

func (file unstattable_file) Close() error {
	*file.was_closed = true
	return nil
}

// unstattable_file_system hands out unstattable_file for every path.
type unstattable_file_system struct {
	was_closed *bool
}

func (file_system unstattable_file_system) Open(requested_path string) (http.File, error) {
	return unstattable_file{was_closed: file_system.was_closed}, nil
}

// TestNoListFileSystem_ReportsStatFailure checks that an entry which cannot be
// inspected is refused rather than served blind, and that its handle is closed
// on the way out — this path runs on every request, so a leak here would
// accumulate.
func TestNoListFileSystem_ReportsStatFailure(t *testing.T) {
	handle_was_closed := false
	no_list_file_system := NoListFileSystem{Fs: unstattable_file_system{was_closed: &handle_was_closed}}

	opened_file, open_error := no_list_file_system.Open("/anything")
	if open_error == nil {
		opened_file.Close()
		t.Fatal("an entry that could not be inspected was served")
	}

	if !handle_was_closed {
		t.Error("the file handle was not closed before returning the error")
	}
}

// TestNoListFileSystem_ServedOverFileServer checks the wrapper in the shape it is
// actually used in main: behind http.FileServer. A browsable directory must come
// back as a 404 whose body does not leak the names of the files inside it.
func TestNoListFileSystem_ServedOverFileServer(t *testing.T) {
	file_server := http.FileServer(NoListFileSystem{Fs: http.Dir(buildFileTree(t))})

	recorder := httptest.NewRecorder()
	file_server.ServeHTTP(recorder, httptest.NewRequest("GET", "/listable/", nil))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status for a listable directory: got %v, want 404", recorder.Code)
	}

	if strings.Contains(recorder.Body.String(), "note.txt") {
		t.Errorf("the directory listing leaked a filename: %q", recorder.Body.String())
	}
}

// ===================  InterceptNotFound  ===================

// TestInterceptNotFound_ReplacesNotFoundBody checks that a 404 from the wrapped
// handler is swallowed and the custom page is written in its place — the default
// "404 page not found" text must not survive anywhere in the response.
func TestInterceptNotFound_ReplacesNotFoundBody(t *testing.T) {
	wrapped_handler := http.HandlerFunc(func(response_writer http.ResponseWriter, request *http.Request) {
		http.Error(response_writer, "404 page not found", http.StatusNotFound)
	})

	custom_not_found_handler := func(response_writer http.ResponseWriter, request *http.Request) {
		response_writer.WriteHeader(http.StatusNotFound)
		response_writer.Write([]byte("<html>custom error page</html>"))
	}

	recorder := httptest.NewRecorder()
	InterceptNotFound(wrapped_handler, custom_not_found_handler).
		ServeHTTP(recorder, httptest.NewRequest("GET", "/missing", nil))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	response_body := recorder.Body.String()
	if response_body != "<html>custom error page</html>" {
		t.Errorf("body: got %q, want only the custom page", response_body)
	}

	if strings.Contains(response_body, "404 page not found") {
		t.Errorf("the default 404 body was not discarded: %q", response_body)
	}
}

// TestInterceptNotFound_ClearsStaleContentType checks that the "text/plain"
// header http.FileServer sets before writing its own 404 is removed. Left in
// place it would make the browser display the HTML error page as source text.
func TestInterceptNotFound_ClearsStaleContentType(t *testing.T) {
	wrapped_handler := http.HandlerFunc(func(response_writer http.ResponseWriter, request *http.Request) {
		http.Error(response_writer, "404 page not found", http.StatusNotFound)
	})

	// The custom handler deliberately sets no content type of its own, matching
	// the real error page, so whatever is observed came from the wrapped handler.
	custom_not_found_handler := func(response_writer http.ResponseWriter, request *http.Request) {
		response_writer.WriteHeader(http.StatusNotFound)
		response_writer.Write([]byte("<html>custom error page</html>"))
	}

	recorder := httptest.NewRecorder()
	InterceptNotFound(wrapped_handler, custom_not_found_handler).
		ServeHTTP(recorder, httptest.NewRequest("GET", "/missing", nil))

	if content_type := recorder.Header().Get("Content-Type"); strings.Contains(content_type, "text/plain") {
		t.Errorf("stale Content-Type survived: %q", content_type)
	}
}

// TestInterceptNotFound_PassesThroughSuccess checks that a normal response is
// forwarded untouched and the fallback handler is never invoked.
func TestInterceptNotFound_PassesThroughSuccess(t *testing.T) {
	wrapped_handler := http.HandlerFunc(func(response_writer http.ResponseWriter, request *http.Request) {
		response_writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		response_writer.WriteHeader(http.StatusOK)
		response_writer.Write([]byte("body { color: red }"))
	})

	fallback_was_called := false
	custom_not_found_handler := func(response_writer http.ResponseWriter, request *http.Request) {
		fallback_was_called = true
	}

	recorder := httptest.NewRecorder()
	InterceptNotFound(wrapped_handler, custom_not_found_handler).
		ServeHTTP(recorder, httptest.NewRequest("GET", "/static/css/style.css", nil))

	if fallback_was_called {
		t.Error("the 404 handler ran for a successful response")
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("status: got %v, want 200", recorder.Code)
	}

	if recorder.Body.String() != "body { color: red }" {
		t.Errorf("body: got %q, want the wrapped handler's body", recorder.Body.String())
	}

	if content_type := recorder.Header().Get("Content-Type"); content_type != "text/css; charset=utf-8" {
		t.Errorf("Content-Type: got %q, want it forwarded unchanged", content_type)
	}
}

// TestInterceptNotFound_PassesThroughOtherErrors checks that a non-404 failure is
// forwarded rather than being turned into the custom 404 page. Only 404 is
// intercepted.
func TestInterceptNotFound_PassesThroughOtherErrors(t *testing.T) {
	wrapped_handler := http.HandlerFunc(func(response_writer http.ResponseWriter, request *http.Request) {
		http.Error(response_writer, "boom", http.StatusInternalServerError)
	})

	fallback_was_called := false
	custom_not_found_handler := func(response_writer http.ResponseWriter, request *http.Request) {
		fallback_was_called = true
	}

	recorder := httptest.NewRecorder()
	InterceptNotFound(wrapped_handler, custom_not_found_handler).
		ServeHTTP(recorder, httptest.NewRequest("GET", "/broken", nil))

	if fallback_was_called {
		t.Error("the 404 handler ran for a 500 response")
	}

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status: got %v, want 500", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "boom") {
		t.Errorf("body: got %q, want the wrapped handler's body", recorder.Body.String())
	}
}

// TestInterceptNotFound_ReportsFullWriteWhileIntercepting checks the contract the
// interceptor offers the wrapped handler while it is discarding a 404 body: the
// write is reported as fully successful. Returning a short write instead would
// make http.FileServer log a spurious error on every missing file.
func TestInterceptNotFound_ReportsFullWriteWhileIntercepting(t *testing.T) {
	discarded_body := []byte("404 page not found\n")

	var reported_byte_count int
	var reported_write_error error

	wrapped_handler := http.HandlerFunc(func(response_writer http.ResponseWriter, request *http.Request) {
		response_writer.WriteHeader(http.StatusNotFound)
		reported_byte_count, reported_write_error = response_writer.Write(discarded_body)
	})

	recorder := httptest.NewRecorder()
	InterceptNotFound(wrapped_handler, func(response_writer http.ResponseWriter, request *http.Request) {}).
		ServeHTTP(recorder, httptest.NewRequest("GET", "/missing", nil))

	if reported_write_error != nil {
		t.Errorf("discarded write reported an error: %v", reported_write_error)
	}

	if reported_byte_count != len(discarded_body) {
		t.Errorf("discarded write reported %d bytes, want %d", reported_byte_count, len(discarded_body))
	}

	if recorder.Body.Len() != 0 {
		t.Errorf("discarded bytes reached the client: %q", recorder.Body.String())
	}
}

// TestInterceptNotFound_OverFileServer checks the two middlewares in the
// composition main actually builds: the no-listing file system behind
// http.FileServer, wrapped by the 404 interceptor.
func TestInterceptNotFound_OverFileServer(t *testing.T) {
	file_server := http.FileServer(NoListFileSystem{Fs: http.Dir(buildFileTree(t))})

	static_file_handler := InterceptNotFound(file_server, func(response_writer http.ResponseWriter, request *http.Request) {
		response_writer.WriteHeader(http.StatusNotFound)
		response_writer.Write([]byte("<html>custom error page</html>"))
	})

	recorder := httptest.NewRecorder()
	static_file_handler.ServeHTTP(recorder, httptest.NewRequest("GET", "/listable/", nil))

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status: got %v, want 404", recorder.Code)
	}

	if recorder.Body.String() != "<html>custom error page</html>" {
		t.Errorf("body: got %q, want the custom error page", recorder.Body.String())
	}
}
