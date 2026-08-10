package main

import (
	"ascii-art-web/internal/handlers"
	"ascii-art-web/middleware"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Route GET / to the home handler — serves the form page
	mux.HandleFunc("/", handlers.HomeHandler)

	// Route POST /ascii-art to the ascii art handler — processes the form and returns the result
	mux.HandleFunc("/ascii-art", handlers.AsciiArtHandler)

	// Route POST /download
	mux.HandleFunc("/download", handlers.DownloadHandler)

	// Serve static files (css, js, etc.) from the static/ directory
	// Middleware prevents from listing the file system
	static_file_server := http.StripPrefix("/static/", http.FileServer(middleware.NoListFileSystem{Fs: http.Dir("static")}))

	// Wrap the static file server so its 404s render templates/error.html via the
	// shared handlers.RenderError helper, instead of http.FileServer's built-in
	// plain-text "404 page not found" response.
	static_file_handler := middleware.InterceptNotFound(static_file_server, func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderError(w, http.StatusNotFound, "Page not found")
	})

	mux.Handle("/static/", static_file_handler)

	// TCP port the server listens on. Defaults to 8080 but can be overridden by
	// setting the PORT environment variable.
	port := "8080"
	port_override := os.Getenv("PORT")
	if port_override != "" {
		port = port_override
	}

	// Custom server as opposed to http generic server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middleware.SecureHeaders(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("Listening on http://localhost:" + port)

	server_error := server.ListenAndServe()
	if server_error != nil {
		slog.Error("server stopped", "error", server_error)
		os.Exit(1)
	}
}
