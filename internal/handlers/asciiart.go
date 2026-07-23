package handlers

import (
	"ascii-art-web/internal/ascii"
	"errors"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// validBanners is the set of accepted banner names.
// Any banner value not in this map is rejected with a 400.
var validBanners = map[string]bool{
	"standard":   true,
	"shadow":     true,
	"thinkertoy": true,
}

// AsciiArtHandler handles POST /ascii-art requests.
// It validates the form inputs, loads the banner file, runs the renderer,
// and returns the result by re-rendering index.html with the populated PageData.
func AsciiArtHandler(w http.ResponseWriter, r *http.Request) {
	// Only POST is accepted — any other method is a bad request
	if r.Method != http.MethodPost {
		RenderError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Requests from the live-preview fetch() carry this header. They receive just
	// the rendered ASCII fragment; classic form submissions (JavaScript disabled)
	// still get a full HTML page, so the form keeps working without JavaScript.
	isLivePreview := r.Header.Get("X-Requested-With") == "fetch"

	text := r.FormValue("text")
	banner := r.FormValue("banner")
	color := r.FormValue("color")

	// For the live preview an empty textarea simply means "clear the output" —
	// it is not the error condition that an empty classic submission is.
	if text == "" {
		if isLivePreview {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			return // empty body: the frontend removes the result area
		}
		RenderError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// banner must be one of the three known values
	if !validBanners[banner] {
		RenderError(w, http.StatusBadRequest, "Bad Request")
		return
	}

	// Load the banner file — missing file is a 404, any other read error is a 500
	bannerMap, err := ascii.LoadBanner(os.DirFS("banners"), banner+".txt")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			RenderError(w, http.StatusNotFound, "Banner not found")
		} else {
			RenderError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	// Normalise line endings — HTML forms submit \r\n, the renderer expects \n
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	// If a color was submitted, validate it and render with color.
	// If absent or invalid, fall back to plain render.
	var display string
	if color != "" {
		validatedColor, err := ascii.ValidateHexColor(color)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid color value")
			return
		}
		display = ascii.RenderWithColor(lines, bannerMap, validatedColor)
	} else {
		display = ascii.Render(lines, bannerMap)
	}

	// Live preview: return only the rendered fragment for the frontend to inject.
	if isLivePreview {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(display))
		return
	}

	// Classic full-page render (graceful fallback when JavaScript is disabled).
	// The submitted inputs are echoed back via PageData so the form keeps state.
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		RenderError(w, http.StatusNotFound, "Template not found")
		return
	}
	// The download payload (RawResult) must always be plain text — the visible
	// Result may be colorized <span> HTML. When a color was applied, recompute
	// the plain render for the hidden download field; otherwise display is
	// already plain.
	rawResult := display
	if color != "" {
		rawResult = ascii.Render(lines, bannerMap)
	}
	pageData := PageData{
		Text:      text,
		Banner:    banner,
		Color:     color,
		Result:    template.HTML(display),
		RawResult: rawResult,
	}
	if err := tmpl.Execute(w, pageData); err != nil {
		RenderError(w, http.StatusInternalServerError, "Internal Server Error")
	}
}
