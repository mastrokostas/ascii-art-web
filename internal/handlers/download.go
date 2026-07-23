package handlers

import (
	"ascii-art/internal/ascii"
	"fmt"
	"html"
	"net/http"
	"strconv"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Only POST is accepted — any other method is a bad request
	if r.Method != http.MethodPost {
		renderError(w, http.StatusBadRequest, "Bad Request")
		return
	}
	asciiArt := r.FormValue("asciiText")
	if asciiArt == "" {
		http.Error(w, "No content to download", http.StatusBadRequest)
		return
	}

	// format selects the export type. Anything other than "html" is the original
	// plain-text download, so forms without a format field keep working.
	if r.FormValue("format") == "html" {
		writeHTMLDownload(w, asciiArt, r.FormValue("color"))
		return
	}

	asciiBytes := []byte(asciiArt)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(asciiBytes)))
	w.Header().Set("Content-Disposition", "attachment; filename=\"ascii-art.txt\"")
	w.Write(asciiBytes)
}

// writeHTMLDownload writes the ASCII art as a self-contained HTML document.
// The art is a single uniform color, so it goes into one <pre>; when a valid
// #rrggbb color is supplied it is applied as an inline style. The art is
// HTML-escaped and the color is validated, so nothing untrusted reaches the
// output — an invalid color is simply dropped and the art renders plain.
func writeHTMLDownload(w http.ResponseWriter, art, color string) {
	style := ""
	if color != "" {
		if validated, err := ascii.ValidateHexColor(color); err == nil {
			// Mirror the result card: the art keeps its color and rests on a single
			// contrasting plate derived from that color — a light plate for dark art,
			// a dark plate for light art, flipping at OKLCH lightness 0.62.
			style = fmt.Sprintf(
				" style=\"color: %s; background-color: oklch(from %s clamp(0, (0.62 - l) * 1000, 1) 0 0 / 0.85)\"",
				validated, validated,
			)
		}
	}

	doc := fmt.Sprintf(htmlDocument, style, html.EscapeString(art))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(doc)))
	w.Header().Set("Content-Disposition", "attachment; filename=\"ascii-art.html\"")
	w.Write([]byte(doc))
}

// htmlDocument is the standalone export page. Two %s slots: the <pre> inline
// style (or empty) and the escaped art. Near-black background so the color pops;
// self-contained with no external fonts, so it opens anywhere offline.
const htmlDocument = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>ASCII Art</title>
<style>
  html, body { margin: 0; height: 100%%; background: #050506; }
  body { display: flex; align-items: center; justify-content: center; padding: 2rem; }
  pre {
    margin: 0;
    padding: 1.1rem 1.3rem;
    border-radius: 16px;
    font-family: 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 14px;
    line-height: 1.15;
    white-space: pre;
    color: #e8e8e8;
  }
  footer {
    position: fixed;
    bottom: 12px;
    left: 0;
    right: 0;
    text-align: center;
    font-family: ui-monospace, monospace;
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.28);
  }
</style>
</head>
<body>
<pre%s>%s</pre>
<footer>made with ASCII Art Web</footer>
</body>
</html>
`
