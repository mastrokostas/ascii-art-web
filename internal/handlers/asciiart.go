package handlers

import (
	"ascii-art-web/internal/ascii"
	"ascii-art-web/middleware"
	"errors"
	"html"
	"html/template"
	"net/http"
	"os"
	"strings"
)

// maxRequestBodyBytes caps how much of a POST body the handlers will read. Go's
// own limit for urlencoded forms is 10 MB, which is far more than this form ever
// needs and far more than is safe here: the renderer expands every input
// character into eight rows of banner glyph, so the response is roughly sixty
// times the size of the input. Bounding the input bounds that amplification.
const maxRequestBodyBytes = 64 * 1024

// maxTextLength caps the number of bytes accepted in the text field itself, so
// an oversized value is rejected with a clear message rather than being read and
// only then found to be unreasonable.
const maxTextLength = 5000

// maxDownloadBodyBytes caps the /download body. It has to be much larger than
// maxRequestBodyBytes because what gets posted there is the *rendered* art, not
// the source text: a full maxTextLength input on a single line expands to a bit
// under 300 KB, so this leaves comfortable headroom above the largest art this
// server will ever produce while still bounding the request.
const maxDownloadBodyBytes = 512 * 1024

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

	// Refuse to read an oversized body at all, before any form parsing happens.
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	// Requests from the live-preview fetch() carry this header. They receive just
	// the rendered ASCII fragment; classic form submissions (JavaScript disabled)
	// still get a full HTML page, so the form keeps working without JavaScript.
	isLivePreview := r.Header.Get("X-Requested-With") == "fetch"

	text := r.FormValue("text")
	banner := r.FormValue("banner")
	color := r.FormValue("color")

	// Reject text that would expand into an unreasonably large render.
	if len(text) > maxTextLength {
		RenderError(w, http.StatusRequestEntityTooLarge, "Text is too long")
		return
	}

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
	// validatedColor is the normalised #rrggbb value, or empty when no color was
	// submitted. Everything downstream echoes this rather than the raw input, so
	// only a value that passed validation is ever written back into the page.
	var validatedColor string
	var display string
	if color != "" {
		normalisedColor, err := ascii.ValidateHexColor(color)
		if err != nil {
			RenderError(w, http.StatusBadRequest, "Invalid color value")
			return
		}
		validatedColor = normalisedColor
		display = ascii.RenderWithColor(lines, bannerMap, validatedColor)
	} else {
		display = ascii.Render(lines, bannerMap)
	}

	// Live preview: return only the rendered fragment for the frontend to inject.
	// The colored render really is HTML (a span per row); the plain render is
	// literal text whose glyphs contain real '<' characters, so it is typed as
	// text/plain to describe honestly what is being sent. The frontend reads the
	// body with response.text() and is not affected by the distinction.
	if isLivePreview {
		if color != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
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

	// display is already-escaped markup on the colored path, but literal text on
	// the plain one — and several banner glyphs (B, K, X, k, x, < and &) contain
	// a real '<'. Since Result is declared template.HTML the template will not
	// escape it, so the plain case has to be escaped here or the browser parses
	// those glyphs as tags and swallows the art. rawResult is deliberately left
	// alone: it is the download payload and must stay plain text.
	displayHTML := template.HTML(display)
	if color == "" {
		displayHTML = template.HTML(html.EscapeString(display))
	}

	pageData := PageData{
		Text:      text,
		Banner:    banner,
		Color:     validatedColor,
		ColorCSS:  template.CSS(validatedColor),
		Result:    displayHTML,
		RawResult: rawResult,
		Nonce:     middleware.NonceFromContext(r.Context()),
	}
	if err := tmpl.Execute(w, pageData); err != nil {
		RenderError(w, http.StatusInternalServerError, "Internal Server Error")
	}
}
