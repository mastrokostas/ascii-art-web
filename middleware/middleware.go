package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
)

// nonce_context_key is the private key type used to store the per-request CSP
// nonce on the request context. It is its own named type (rather than a plain
// string) so no other package can collide with the key by accident.
type nonce_context_key struct{}

// NonceFromContext returns the per-request Content-Security-Policy nonce that
// SecureHeaders placed on the request context, or an empty string when the
// request did not pass through that middleware. Handlers hand this value to
// their templates so an inline <style> block can be marked as trusted.
func NonceFromContext(request_context context.Context) string {
	nonce, is_string := request_context.Value(nonce_context_key{}).(string)
	if !is_string {
		return ""
	}

	return nonce
}

// generate_nonce returns 16 cryptographically random bytes, base64 encoded, for
// use as a single request's CSP nonce. A fresh value per request is what makes
// the nonce meaningful: a reused nonce would let injected markup claim it.
func generate_nonce() string {
	nonce_bytes := make([]byte, 16)

	// crypto/rand.Read is documented never to return an error and to always
	// fill the slice completely, so there is no failure case to handle here.
	rand.Read(nonce_bytes)

	return base64.StdEncoding.EncodeToString(nonce_bytes)
}

// NoListFileSystem wraps an http.FileSystem to disable automatic directory
// listings. When a directory is requested it is only served if that directory
// contains an index.html file; otherwise the request is treated as "not found"
// so visitors can never browse the raw file tree.
type NoListFileSystem struct {
	// Fs is the underlying file system that files are actually served from.
	Fs http.FileSystem
}

func (no_list_file_system NoListFileSystem) Open(requested_path string) (http.File, error) {
	// Try to open whatever was requested (a file or a directory).
	opened_file, open_error := no_list_file_system.Fs.Open(requested_path)
	if open_error != nil {
		return nil, open_error
	}

	// Inspect the entry so we can tell files and directories apart.
	file_info, stat_error := opened_file.Stat()
	if stat_error != nil {
		opened_file.Close()
		return nil, stat_error
	}

	// For a directory, only allow access when it holds an index.html file;
	// otherwise report "not found" to suppress the directory listing.
	if file_info.IsDir() {
		index_file_path := requested_path + "index.html"
		index_file, index_open_error := no_list_file_system.Fs.Open(index_file_path)
		if index_open_error != nil {
			opened_file.Close()
			return nil, os.ErrNotExist
		}

		// The index file was only opened to confirm it exists; close it
		// immediately so its handle is not leaked on every directory request.
		index_file.Close()
	}

	return opened_file, nil
}

// InterceptNotFound wraps next_handler and, whenever the wrapped handler
// responds with a 404, suppresses that default response body and delegates to
// not_found_handler instead. This lets the static file server render a custom
// 404 page (templates/error.html) rather than http.FileServer's built-in plain
// text "404 page not found" response.
func InterceptNotFound(next_handler http.Handler, not_found_handler http.HandlerFunc) http.Handler {
	return intercept_not_found_handler{
		next_handler:      next_handler,
		not_found_handler: not_found_handler,
	}
}

// intercept_not_found_handler holds the wrapped handler and the fallback handler
// that renders the custom 404 page.
type intercept_not_found_handler struct {
	// next_handler is the handler whose responses are being watched (the static
	// file server).
	next_handler http.Handler
	// not_found_handler is invoked to write the response body whenever
	// next_handler produces a 404.
	not_found_handler http.HandlerFunc
}

// ServeHTTP runs next_handler through a response interceptor. If next_handler
// wrote a 404, the interceptor will have swallowed that default response, so we
// clean up the stale headers and let not_found_handler render the custom page.
func (handler intercept_not_found_handler) ServeHTTP(response_writer http.ResponseWriter, request *http.Request) {
	// Wrap the real response writer so we can watch the status code and, on a
	// 404, discard http.FileServer's default plain-text body.
	interceptor := &not_found_interceptor{ResponseWriter: response_writer}

	handler.next_handler.ServeHTTP(interceptor, request)

	// If the wrapped handler did not report a 404, its response has already been
	// written straight through and there is nothing more to do.
	if !interceptor.is_not_found {
		return
	}

	// http.FileServer sets "Content-Type: text/plain" before writing its 404.
	// That header is still sitting on the response writer, and the custom error
	// page does not set a content type of its own, so remove the stale value to
	// let the HTML template be served (and sniffed) as HTML.
	response_writer.Header().Del("Content-Type")

	// Render the custom 404 page onto the real response writer.
	handler.not_found_handler(response_writer, request)
}

// not_found_interceptor wraps an http.ResponseWriter to detect a 404 status and,
// when one is seen, prevent the wrapped handler's default status and body from
// reaching the client so a custom 404 page can be written instead.
type not_found_interceptor struct {
	// ResponseWriter is the real response writer being wrapped; embedding it
	// means every method we do not override is forwarded to it unchanged.
	http.ResponseWriter
	// is_not_found records whether the wrapped handler responded with a 404.
	is_not_found bool
}

// WriteHeader records a 404 status and suppresses it (so the real response
// writer never sends it); any other status is forwarded through unchanged.
func (interceptor *not_found_interceptor) WriteHeader(status_code int) {
	if status_code == http.StatusNotFound {
		// Remember the 404 and do not forward it; the custom 404 handler will
		// write its own status later.
		interceptor.is_not_found = true
		return
	}

	interceptor.ResponseWriter.WriteHeader(status_code)
}

// Write discards the response body while a 404 is being intercepted (so the
// default "404 page not found" text is dropped); otherwise it forwards the
// bytes to the real response writer.
func (interceptor *not_found_interceptor) Write(data []byte) (int, error) {
	if interceptor.is_not_found {
		// Pretend the write succeeded so the wrapped handler does not error,
		// but drop the bytes so the default 404 body is never sent.
		return len(data), nil
	}

	return interceptor.ResponseWriter.Write(data)
}

// SecureHeaders wraps next_handler and attaches a set of security-related HTTP
// response headers to every response before delegating to the wrapped handler.
func SecureHeaders(next_handler http.Handler) http.Handler {
	return secure_headers_handler{next_handler: next_handler}
}

// secure_headers_handler holds the handler that requests are passed on to
// after the security headers have been set.
type secure_headers_handler struct {
	next_handler http.Handler
}

// ServeHTTP sets the security-related response headers, then delegates to the
// wrapped handler.
func (handler secure_headers_handler) ServeHTTP(response_writer http.ResponseWriter, request *http.Request) {
	// Stop browsers from MIME-sniffing responses away from their declared type.
	response_writer.Header().Set("X-Content-Type-Options", "nosniff")
	// Forbid the site from being embedded in a frame (clickjacking defense).
	response_writer.Header().Set("X-Frame-Options", "DENY")
	// Send only the origin as the referrer when navigating to another origin.
	response_writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	// Disable powerful browser features the site does not use.
	response_writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// HTTP Strict-Transport-Security tells browsers to reach this origin over
	// HTTPS only. It is opt-in through an environment variable because sending
	// it from a plain-HTTP origin (local development on localhost) pins that
	// origin to HTTPS in the developer's browser and makes it unreachable.
	// Enable it only where TLS actually terminates in front of the app.
	if os.Getenv("ENABLE_HSTS") == "true" {
		response_writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// A fresh nonce for this request, used to authorise the one inline <style>
	// block the JavaScript-disabled page needs for the user's chosen art color.
	nonce := generate_nonce()

	// Content-Security-Policy: keep every resource first-party ('self') so no
	// injected or third-party script can run. The site loads all of its
	// JavaScript from /static/js and has no inline <script>, so 'self' alone
	// covers scripts. style-src carries a per-request nonce rather than
	// 'unsafe-inline': the only inline style left is the single <style> block
	// that carries the picked color into the page on the no-JavaScript path,
	// and it is tagged with this nonce. Every other style lives in style.css.
	response_writer.Header().Set("Content-Security-Policy",
		"default-src 'self'; "+
			"script-src 'self'; "+
			"style-src 'self' 'nonce-"+nonce+"'; "+
			"font-src 'self'; "+
			"img-src 'self'; "+
			"connect-src 'self'; "+
			"form-action 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"object-src 'none'")

	// Publish the nonce on the request context so the handlers downstream can
	// stamp it onto the inline <style> block they render.
	request_with_nonce := request.WithContext(
		context.WithValue(request.Context(), nonce_context_key{}, nonce),
	)

	handler.next_handler.ServeHTTP(response_writer, request_with_nonce)
}
