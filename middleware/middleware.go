package middleware

import (
	"net/http"
	"os"
)

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

	handler.next_handler.ServeHTTP(response_writer, request)
}
