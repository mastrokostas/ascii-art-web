package handlers

import (
	"html/template"
	"net/http"
)

// PageData is the data structure passed to the index template.
// Result holds the ASCII art output after a POST request.
// It is declared as template.HTML so the template engine does not escape
// the <span> tags injected by RenderWithColor.
// It is empty on the initial GET request, so the result area is not rendered.
//
// Text, Banner and Color echo the submitted inputs back into the form on the
// classic (JavaScript-disabled) POST path, so the form keeps its state after a
// full-page reload. They are empty on the initial GET request.
type PageData struct {
	Text      string
	Banner    string
	Color     string
	Result    template.HTML
	RawResult string
}

// RenderError is a shared helper used by all handlers to write an error response.
// It sets the status code, then attempts to render error.html with the code and message.
// If the error template itself is missing, it falls back to a plain http.Error.
// It is exported so callers outside this package (e.g. the static file server's
// custom 404 handler in main) can reuse the same error page.
func RenderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, message, statusCode)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"Code":    statusCode,
		"Message": message,
	})
}

// HomeHandler handles GET / requests.
// It rejects any path other than "/" with a 404 to prevent the default mux catch-all behaviour.
// On success it renders index.html with an empty PageData (no result yet).
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// The default mux routes everything that doesn't match another handler to "/".
	// We explicitly reject anything that isn't the exact root path.
	if r.URL.Path != "/" {
		RenderError(w, http.StatusNotFound, "Page not found")
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		RenderError(w, http.StatusNotFound, "Template not found")
		return
	}

	// Execute the template with empty PageData — result area will not render
	if err := tmpl.Execute(w, PageData{}); err != nil {
		RenderError(w, http.StatusInternalServerError, "Internal Server Error")
	}
}
