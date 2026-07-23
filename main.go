package main

import (
	"ascii-art/internal/handlers"
	"log"
	"net/http"
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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
