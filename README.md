<div style="display:flex; justify-content:center;">
<pre style="overflow-x:auto; overflow-y:hidden; white-space:pre; max-width:100%;"><font color="#007D9C">                   _  _                       _                            _     
                  (_)(_)                     | |                          | |    
  __ _  ___   ___  _  _           __ _  _ __ | |_         __      __  ___ | |__  
 / _` |/ __| / __|| || | ______  / _` || '__|| __| ______ \ \ /\ / / / _ \| '_ \ 
| (_| |\__ \| (__ | || ||______|| (_| || |   | |_ |______| \ V  V / |  __/| |_) |
 \__,_||___/ \___||_||_|         \__,_||_|    \__|          \_/\_/   \___||_.__/ </font></pre>
</div>

<div align="center">A web-based application written in Go that allows users to generate large ASCII block letters through a graphical interface in the browser, with color, a live preview that renders as you type, and a one-click export of the result to a file.</div>

## 🌐 Overview

This project is a web server implementation of the ASCII Art generator. Instead of using the command line, users can access a web page, type their text, select a banner font, optionally pick a color, see the generated ASCII art directly in their browser, and download it as a file.

By default the art renders live as the user types — each change sends a background `fetch` to the server and the returned art is dropped into the page without reloading. A switch turns this off in favour of rendering on a button press instead. The server still serves a full page on a plain form submission, so the form works even with JavaScript disabled.

The server handles specific HTTP endpoints and responds with appropriate HTTP status codes (200 OK, 400 Bad Request, 404 Not Found, 500 Internal Server Error).

## ✨ Features

- **Export to a file** — a download button under the result saves the generated art. Pick `.txt` for the plain art or `.html` for a self-contained page that keeps the color you chose.
- **Live preview** — the art renders as you type, debounced, with no page reload, so the animated background and your text are never reset.
- **Live / manual switch** — a toggle beside the banner selection chooses between rendering live or only on the Generate button.
- **Color** — an interactive HSL palette and a hex input, kept in sync both ways; colored output is rendered with inline `<span>` styles.
- **Auto-growing input** — the textarea grows to fit its content instead of scrolling, with no manual resize handle.
- **Progressive enhancement** — without JavaScript the form posts normally and the server returns a full page with the result and the submitted inputs preserved.

## 🚀 Usage

To start the server, open your terminal in the project's root directory and run:

```bash
go run .
```

Once the server is running, open your preferred web browser and navigate to:
http://localhost:8080

### 🐳 Docker

Build the image from the project's root directory:

```bash
docker build -t ascii-art-web .
```

Run the container, mapping the container's port `8080` to the host:

```bash
docker run -p 8080:8080 ascii-art-web
```

Then open http://localhost:8080 in your browser.

Check the image metadata (the `org.opencontainers.image` labels set in the `Dockerfile`):

```bash
docker inspect --format '{{json .Config.Labels}}' <container_name>
```

Open a shell inside the running container and list its contents:

```bash
docker exec -it <container_name> /bin/bash
ls -l
```

### 🔒 Security Note:
For security reasons, the container application does not run as root. It runs under a dedicated non-root user (ascii-art-web with UID 1001).

To prevent unauthorized modifications, the application files copied into the container are owned by root and are read-only for the execution user.

If you need to perform troubleshooting or administrative tasks inside the running container with root privileges, you can override the default user by passing the -u 0 flag:

```bash
docker exec -u 0 -it <container_name> /bin/bash
```

### Response headers and the Content-Security-Policy

`middleware.SecureHeaders` attaches `X-Content-Type-Options`, `X-Frame-Options`,
`Referrer-Policy`, `Permissions-Policy` and a Content-Security-Policy to every response.

The policy keeps every resource first-party — `default-src 'self'`, with `frame-ancestors 'none'`,
`base-uri 'self'` and `object-src 'none'`. `style-src` carries a **per-request nonce** rather than
`'unsafe-inline'`: 16 random bytes from `crypto/rand`, generated fresh for each request, published
on the request context and read back by the handlers through `middleware.NonceFromContext`.

That means no inline `style` attribute can appear in markup the browser parses. The art color is
applied through the `--art-color` custom property instead — `live-preview.js` sets it via the
CSSOM (which CSP does not police), and the JavaScript-disabled path uses the single nonced
`<style>` block in `index.html`. Colored rows carry the `.art-line` class rather than
`style="color:…"`.

`Strict-Transport-Security` is opt-in through the `ENABLE_HSTS=true` environment variable. It is
off by default on purpose: sending HSTS from a plain-HTTP development origin pins that origin to
HTTPS in your browser and makes it unreachable. Enable it only where TLS terminates in front of
the app.

### Input handling

The inputs are validated, not sanitized — stripping characters would corrupt legitimate ASCII art.
`banner` must match a fixed allow-list, `color` must pass `ascii.ValidateHexColor` (`#rrggbb`), and
`text` is capped at 5,000 bytes with the request body capped at 64 KB (512 KB on `/download`, which
carries rendered art rather than source text). The caps matter because the renderer expands every
input character into eight rows of banner glyph, roughly a sixtyfold amplification.

Output is escaped at the point of display. Several `standard` banner glyphs — `B`, `K`, `X`, `k`,
`x`, `<` and `&` — contain a literal `<`, so the rendered art is HTML-escaped before it reaches the
page. The download payload is deliberately *not* escaped: it must stay plain text, and a regression
test pins that distinction in place.

## 🔌 Endpoints

- **GET /** — serves the main HTML page: the text input, banner radios, color picker, and the live/manual switch.
- **POST /ascii-art** — accepts `text`, `banner`, and an optional `color`. The response depends on who is asking:
  - A request from the live preview carries an `X-Requested-With: fetch` header and gets back **only** the rendered ASCII fragment. Empty `text` returns an empty `200` (the frontend clears the output).
  - A plain form submission (no header) gets back a **full HTML page** with the result, and the submitted inputs echoed so the form keeps its state. Empty `text` here is a `400`.
- **POST /download** — exports the art already on the page as a file. Accepts `asciiText` (the plain, uncolored art), an optional `format` (`txt` or `html`, defaulting to `txt`), and an optional `color`. It responds with a `Content-Disposition: attachment` header so the browser saves it as `ascii-art.txt` or `ascii-art.html`. Empty `asciiText` or a non-POST method is a `400`.
- **GET /static/** — serves CSS and JavaScript from the `static/` directory.

## 🔠 Banner Fonts
Three fonts are available and selectable from the web interface:

    - standard
    - shadow
    - thinkertoy

The application uses Go's io/fs package via os.DirFS("banners") to safely load the banner files without modifying them.

## 📂 Project Structure

```text
ascii-art-web/
├── main.go                        # Server initialization, port configuration, and routing
├── go.mod                         # Go module definition
├── Dockerfile                     # Container build
├── README.md                      # Project documentation and usage guide
├── LICENSE
├── internal/
│   ├── ascii/
│   │   ├── banner.go              # Banner file loader using io/fs
│   │   ├── render.go              # ASCII art rendering engine (plain + colored)
│   │   └── color.go               # Hex color validation
│   └── handlers/
│       ├── home.go                # HTTP handler for the index page (GET)
│       ├── asciiart.go            # HTTP handler for /ascii-art (POST): page or fragment
│       ├── download.go            # HTTP handler for /download (POST): .txt or .html export
│       └── handlers_test.go       # HTTP tests for the endpoints
├── templates/
│   ├── index.html                 # Front-end interface: form, color picker, result card, download form
│   └── error.html                 # Error page template
├── static/
│   ├── css/
│   │   └── style.css              # Stylesheet
│   └── js/
│       ├── live-preview.js        # Debounced fetch rendering + keeps the download form in sync
│       ├── color-picker.js        # HSL palette and hex input
│       ├── format-dropdown.js     # Custom .txt / .html export-format dropdown
│       ├── matrix.js              # Animated matrix-rain background
│       └── cursor-glow.js         # Cursor-following glow on the glass cards
├── banners/
│   ├── shadow.txt
│   ├── standard.txt
│   └── thinkertoy.txt
└── .documentation/                # Design notes, workflow, and changelog
```

## 🧪 Running Tests

The project includes HTTP testing using Go's built-in net/http/httptest package. To run the tests and verify the server's behavior, execute:

```bash
go test ./...
```

### Manually triggering the error responses

The tests in `handlers_test.go` already cover these, but a reviewer can also exercise each status code by hand against a running server (`go run .`).

**A note on the empty form:** in the browser with JavaScript on, an empty submission will *not* produce a `400`. The live-preview script intercepts the form submit and sends the request via `fetch`, and on that path empty text returns an empty `200` (it simply clears the output). To exercise the raw server-side validation, bypass the JavaScript — either use `curl` (which sends no `X-Requested-With: fetch` header, so the handler takes the full-page path) or disable JavaScript in the browser and submit the form.

**400 Bad Request**

```bash
# empty text on a plain form post
curl -i -X POST localhost:8080/ascii-art --data 'text=&banner=standard'
# unknown banner name
curl -i -X POST localhost:8080/ascii-art --data 'text=hi&banner=bogus'
# malformed color
curl -i -X POST localhost:8080/ascii-art --data 'text=hi&banner=standard&color=zzz'
# wrong method (GET instead of POST)
curl -i localhost:8080/ascii-art
# download with nothing to export
curl -i -X POST localhost:8080/download --data 'asciiText=&format=txt'
```

The export itself can also be exercised from the command line:

```bash
# plain text export
curl -i -X POST localhost:8080/download --data 'asciiText=hello&format=txt'
# self-contained HTML export, colored
curl -i -X POST localhost:8080/download --data 'asciiText=hello&format=html&color=%23007d9c'
```

**404 Not Found**

```bash
# any path other than "/"
curl -i localhost:8080/does-not-exist
```

The handler also returns `404` ("Banner not found") when a selected banner's `.txt` file is missing from `banners/`. Since the three selectable names are validated and all three files ship with the project, you only reach this by removing one:

```bash
mv banners/thinkertoy.txt banners/thinkertoy.bak
curl -i -X POST localhost:8080/ascii-art --data 'text=hi&banner=thinkertoy'
mv banners/thinkertoy.bak banners/thinkertoy.txt
```

**500 Internal Server Error**

A `500` only happens on an unexpected internal failure — the expected cases (bad input, missing file) are handled as `400`/`404`. To force one, make a banner file present but unreadable so the read fails with something other than "not found"; replacing it with a directory does this:

```bash
mv banners/shadow.txt banners/shadow.bak && mkdir banners/shadow.txt
curl -i -X POST localhost:8080/ascii-art --data 'text=hi&banner=shadow'
rmdir banners/shadow.txt && mv banners/shadow.bak banners/shadow.txt
```

(Any of these is undone in one step with `git restore banners/<file>` if the repo is committed. As a non-root user, `chmod 000 banners/shadow.txt` also forces the 500; root bypasses file permissions, so the directory swap is the portable trigger.)

## 🔧 Implementation Details

### Banner loading

Each banner file contains 95 character glyphs covering ASCII 32 (space) through 126 (~). The file starts with a leading newline which is trimmed. The remaining content is split on double newlines (`\n\n`) to produce one block per character. Each block is split on single newlines to produce the 8 lines that form the glyph. The result is stored in a `map[rune][]string` where the key is the character and the value is its 8-line representation.

### Rendering

The input text is normalised by replacing `\r\n` (submitted by HTML forms) with `\n`, then split into lines. Each non-empty line is rendered row by row: for each of the 8 art rows, every character in the line contributes one row slice from the banner map, concatenated into a single output line. Empty input lines produce a blank line in the output. The full result is returned as a string — written straight to the response as a fragment for live-preview requests, or injected into the HTML template via `{{.Result}}` for a full-page render.

### Color rendering

If a `color` is submitted it is validated by `ValidateHexColor`, which accepts only `#rrggbb` and returns the normalised lowercase value (any other form is a `400`). `RenderWithColor` then wraps each output row in a `<span style="color: …">`. Each row is passed through `html.EscapeString` first, because some banner glyphs are drawn with literal `<` and `>` characters that would otherwise corrupt the markup. `Result` is typed `template.HTML` so the template does not escape these intentional spans.

### Exporting the art to a file

The result card carries a small `POST /download` form. A hidden textarea holds `asciiText` — always the **plain** art, never the colored `<span>` markup — and a hidden field carries the chosen `color`; `live-preview.js` refreshes both every time the art re-renders, so the export always matches what is on screen. The format `<select>` (`.txt` / `.html`) is submitted as `format`, and `format-dropdown.js` replaces the unstylable native popup with a matching glass listbox while leaving the real `<select>` in the DOM, so the export still works with JavaScript disabled.

`DownloadHandler` writes the file with a `Content-Disposition: attachment` header:

- **`.txt`** — the art written out as `text/plain`, exactly as rendered.
- **`.html`** — a self-contained `text/html` document (no external fonts or stylesheets, so it opens offline) with the art in a single `<pre>`. If a valid `#rrggbb` color was submitted it becomes an inline `color` style, on a plate whose lightness flips against the art color so the text always stays legible. The art is passed through `html.EscapeString` and the color through `ValidateHexColor` — an invalid color is dropped and the art simply renders plain, so nothing untrusted reaches the output.

Anything other than `html` in `format` falls through to the plain-text download, which keeps a form that omits the field working.

### Live preview and the fragment response

`AsciiArtHandler` distinguishes the two callers by the `X-Requested-With: fetch` header that the live-preview script adds. With the header it writes only the rendered ASCII; without it (a plain form submit) it renders the full page and echoes `Text`, `Banner`, and `Color` back through `PageData` so the form survives the reload. The frontend debounces typing, uses an `AbortController` so responses cannot arrive out of order, and injects the fragment as `innerHTML` for colored output or `textContent` for plain output (so literal angle brackets are not parsed as tags). See `.documentation/workflow.md` for the full request flow.

## ✍️ Authors

- Stergios Fourlataras
- Konstantinos Koletsis

## 📜 License

**All Rights Reserved.** This project is published for demonstration and portfolio
purposes only. It is **not** open source, and no permission is granted to use, copy,
modify, or distribute it. See [LICENSE](LICENSE) for the full terms and
[28702553+mastrokostas@users.noreply.github.com](mailto:28702553+mastrokostas@users.noreply.github.com)
for licensing enquiries.
