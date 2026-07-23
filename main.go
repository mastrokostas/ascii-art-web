package main

import (
	"ascii-art/internal/handlers"
	"ascii-art/middleware"
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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(middleware.NoListFileSystem{Fs: http.Dir("static")})))

	// TCP port the server listens on. Defaults to 8080 but can be overridden by
	// setting the PORT environment variable.
	port := "8080"
	port_override := os.Getenv("PORT")
	if port_override != "" {
		port = port_override
	}

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
