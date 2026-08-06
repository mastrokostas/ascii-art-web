# Changelog

## 2026-08-06

### Added
- `internal/ascii/ascii_test.go`, the first test file for the rendering
  package (17 tests). It also settles a misleading coverage reading: `go test
  -cover ./...` counts only the statements a package's *own* test files
  execute, so `internal/ascii` reported **0.0%** even though the handler tests
  drove `LoadBanner`, `Render`, `RenderWithColor` and `ValidateHexColor` on
  every POST. Run as `go test -coverpkg=./... ./...` the same suite already
  reported those four functions at 81–90%. The new file makes the figure honest
  rather than incidental:
  - `ValidateHexColor`: the valid case, lowercase normalisation, whitespace
    trimming, a missing `#`, wrong digit counts (`#fff`, `#ff88000`, `#`), and
    the per-character hex loop (`#gg0000`, `#12345z`, `#ff 800`, `#-f0000`).
    That last group was previously unreachable — the handler tests use
    `"notacolor"` and `"bogus"`, which fail on the `#` prefix check and never
    enter the digit loop.
  - `LoadBanner`: the block-to-rune mapping starting at ASCII 32, asserted
    against a synthetic `fstest.MapFS` banner so a parser bug cannot hide behind
    the real files; the read-error path, checked with `errors.Is(err,
    fs.ErrNotExist)` because `AsciiArtHandler` splits 404 from 500 on exactly
    that; and all three shipped banner files, each asserted to yield 95
    characters with at least eight rows apiece.
  - `Render` and `RenderWithColor`: exact output for a known input, pinning both
    the row order and the left-to-right character order; the blank-line branch in
    both functions; unmapped runes being skipped rather than panicking on the
    row lookup; the `.art-line` span wrapping; `<`, `&` and `>` escaping; and
    that neither an inline `style=` attribute nor the raw color ever reaches the
    markup, which is what keeps `'unsafe-inline'` out of the CSP (2026-08-03).
- 14 tests in `middleware/middleware_test.go`, covering the three features that
  had none. `NoListFileSystem.Open` and `InterceptNotFound` were at 0%, which
  meant the directory-listing suppression — a security control — was entirely
  unexercised:
  - `NoListFileSystem.Open`: a directory without `index.html` refused as
    `os.ErrNotExist`, one with `index.html` served, regular files served,
    missing paths propagated, and the stat-failure branch reached through a fake
    `http.File`, including the assertion that the handle is closed before the
    error returns. An end-to-end check through `http.FileServer` confirms the
    resulting 404 body does not leak the name of a file inside the hidden
    directory.
  - `InterceptNotFound`: the 404 body replaced by only the custom page, the
    stale `text/plain` header cleared, 200 and 500 both forwarded untouched with
    the fallback never invoked, and the two middlewares composed the way
    `main.go` composes them.
  - `not_found_interceptor.Write`: the discarded write is reported back as fully
    successful, so `http.FileServer` does not log a short write on every miss.
  - Plus the `NonceFromContext` path for a request that never passed through
    `SecureHeaders`, and the fixed hardening headers.
- 14 tests in `internal/handlers/handlers_test.go`:
  - The live-preview path (`X-Requested-With: fetch`) had no coverage at all.
    Empty text now asserts 200 with an empty body — "clear the output", not the
    400 a classic submission gets; a plain preview asserts a `text/plain`
    fragment whose `<` glyphs stay literal; a colored preview asserts a
    `text/html` fragment carrying `.art-line` spans; both assert a fragment
    rather than a full page. A table check confirms the preview path still
    enforces the banner, color and length validation.
  - `/download` rejections: `GET` → 400, and an empty `asciiText` → 400 with no
    `Content-Disposition` offered.
  - The on-disk failure paths, reached with `t.Chdir` into an empty temporary
    directory (it restores the previous directory and refuses to run in a
    parallel test, so the switch cannot leak into the rest of the package):
    `RenderError`'s `http.Error` fallback when `templates/error.html` is
    missing, tested both in isolation and through `HomeHandler`; a missing
    banner file → 404; an *unreadable* banner → 500, produced by standing a
    directory where `standard.txt` belongs so the read fails with something
    other than `ErrNotExist`; and a template parse failure, with the banner
    deliberately copied in so the render gets far enough to reach it.
  - Template execution failures, driven by a `failing_response_writer` whose
    every write errors the way a client that hung up mid-response would. Both
    `HomeHandler` and `AsciiArtHandler` are asserted to settle on a 500.

  All three packages now report 100% statement coverage (from 0.0%, 36.2% and
  76.5%); across the module `-coverpkg=./...` rises from 43.8% to 91.4%. The
  remainder is `main.go`, which stays at 0% because `main()` is never invoked
  by a test — reaching it needs the mux and server construction extracted into
  a separate function, a production-code change that was left alone. 73 tests
  pass under `-race` and under repeated `-shuffle=on` runs, so the `t.Chdir`
  tests are order-independent.

## 2026-08-05

### Added
- `nginx/nginx.conf`, the reverse-proxy configuration that the `nginx/`
  container image copies over the stock server block. It is what makes the
  image added on 2026-08-04 do anything useful:
  - A `listen 80` server that answers every hostname (`server_name _`) with a
    `301` to `https://$host$request_uri`, so a plain-HTTP request is redirected
    rather than proxied and never reaches the app.
  - A `listen 443 ssl` server terminating TLS with a Cloudflare Origin
    certificate, read from `/etc/nginx/certs/cloudflare-origin.pem` and
    `/etc/nginx/certs/cloudflare-origin.key`. Those files are deliberately not
    in the repository; they are supplied at run time through a bind mount.
  - A single `location /` block. The prefix match on `/` catches every path and
    proxies to `http://app:8080`, because the Go mux in `main.go` does its own
    routing and nginx needs no per-route rules. `app` is the Docker network
    alias of the application container and `8080` is the port `main.go`
    defaults to.
  - Four forwarding headers on that proxied request: `Host`, so the app sees
    the hostname the browser actually asked for instead of `app:8080`;
    `X-Real-IP` with the client address; `X-Forwarded-For`, appending the
    client to any existing chain; and `X-Forwarded-Proto`, which reports
    `https` even though the inward hop to the app is plain HTTP.
  With TLS now terminating in front of the application, the condition for
  `ENABLE_HSTS=true` (added 2026-08-03) is met, so the Go container can send
  `Strict-Transport-Security` in this deployment.

## 2026-08-04

### Added
- `nginx/` directory holding the reverse proxy that fronts the Go server. Its
  `Dockerfile` builds on `nginx:1.27` — pinned to a minor rather than
  `nginx:latest` — and copies `nginx.conf` to
  `/etc/nginx/conf.d/default.conf`, the path the stock image already includes,
  which replaces the built-in "Welcome to nginx" server block. `nginx.conf`
  was committed empty here and written the following day.
- A favicon: `static/favicon/favicon.svg`, linked from both
  `templates/index.html` and `templates/error.html` with
  `<link rel="icon" type="image/svg+xml" href="/static/favicon/favicon.svg">`.
  The icon is drawn as inline SVG rather than shipped as a raster set — a
  rounded `#0D0D0D` square behind a bold monospace `A` and an underscore bar,
  both `#00FF41` under an SVG `drop-shadow` filter for the terminal glow — so
  one file stays sharp at every size. It is served from `/static`, which the
  existing `img-src 'self'` directive already covers, so the
  Content-Security-Policy needed no change.

## 2026-08-03

### Fixed
- Plain (uncolored) ASCII art is now HTML-escaped before it reaches the page.
  `ascii.RenderWithColor` escaped its output but `ascii.Render` did not, and the
  result was assigned to `PageData.Result`, which is declared `template.HTML` and
  therefore not escaped by the template either. Seven `standard` banner glyphs —
  `B`, `K`, `X`, `k`, `x`, `<` and `&` — contain a literal `<`, so the browser
  parsed those as tags and swallowed part of the art. Only the full-page
  (JavaScript-disabled) path was affected; the live preview already routed plain
  output through `textContent`.
  The escaping is applied in `AsciiArtHandler`, **not** inside `ascii.Render`:
  the same render feeds the hidden download field, where the art must stay plain
  text. `TestAsciiArtHandler_DownloadPayloadKeepsRawGlyphs` pins that down.
- `RenderError` no longer writes the response status before parsing the error
  template. On a parse failure it fell through to `http.Error`, which writes a
  status of its own, producing a "superfluous response.WriteHeader call" log.

### Security
- `style-src` in the Content-Security-Policy no longer allows `'unsafe-inline'`.
  It now carries a per-request nonce (16 bytes from `crypto/rand`, base64), put
  on the request context by `SecureHeaders` and read back by handlers through the
  new `middleware.NonceFromContext`. Removing `'unsafe-inline'` required moving
  every inline style out of parsed markup:
  - `ascii.RenderWithColor` emits `<span class="art-line">` instead of
    `<span style="color:…">`; the color arrives through the `--art-color` custom
    property that `style.css` already used.
  - The hidden download textarea uses a new `.hidden-field` class instead of
    `style="display:none"`, in both `templates/index.html` and the copy that
    `live-preview.js` builds at runtime.
  - `cursor-glow.js` no longer constructs a `<style>` element for the `?pose`
    screenshot aid; that rule moved verbatim into `style.css` and the script only
    adds the `.glow-posed` class.
  - Custom properties set through the CSSOM (`element.style.setProperty`) are
    untouched — CSP does not govern the CSSOM, only parsed markup.
  The no-JavaScript path still needs to apply a user-chosen color, so
  `index.html` carries one nonced `<style>` block, emitted only when a color was
  submitted. Its value is passed as `template.CSS` so `html/template`'s CSS
  filter does not rewrite the validated color to `ZgotmplZ`.
- Request bodies are now capped with `http.MaxBytesReader` — 64 KB on
  `/ascii-art`, 512 KB on `/download`, which carries rendered art rather than
  source text — and the `text` field is capped at 5,000 bytes. Go's own limit for
  urlencoded forms is 10 MB, which was far too generous here: the renderer
  expands each input character into eight rows of glyph, so a 10 MB body produced
  roughly 640 MB of string in memory.
- `Strict-Transport-Security` is sent when `ENABLE_HSTS=true`. It is opt-in so
  that local plain-HTTP development is unaffected — HSTS from a localhost origin
  pins it to HTTPS in the developer's browser.

### Changed
- Relicensed the project from the MIT License to a proprietary **All Rights
  Reserved** notice in `LICENSE`, ahead of making the repository public. MIT
  granted everyone the right to use, sell, and sublicense the code, which is not
  the intent for this project. The MIT-licensed code was never publicly
  distributed — the repository was private for its whole history — so no third
  party ever received the MIT grant and the relicense is clean.
- The new `LICENSE` grants no rights at all, states explicitly that public
  visibility on GitHub is not a license, excludes the third-party banner files
  and web fonts from the copyright claim, and keeps the MIT warranty disclaimer
  verbatim as clause 5.

### Added
- `## 📜 License` section at the end of `README.md`, so the restriction is
  visible to anyone who reads only the README.
- SIL Open Font License texts for the three self-hosted font families:
  `static/fonts/OFL-Geist.txt`, `static/fonts/OFL-InstrumentSerif.txt`, and
  `static/fonts/OFL-JetBrainsMono.txt`. The fonts were previously shipped as
  bare `.woff2` files with no license text, which the OFL does not permit —
  it requires the license to be distributed with the Font Software. Each file
  is the upstream text copied verbatim from the `google/fonts` repository,
  retaining its own copyright line. `LICENSE` clause 3 now names all three.
- `middleware/middleware_test.go`: covers the CSP nonce (present in the policy,
  absent of `'unsafe-inline'`, matching the value handlers read from the context,
  and freshly generated per request) and the HSTS opt-in switch.
- Handler tests for the escaping fix and the new input limits, including
  `TestAsciiArtHandler_PlainResultIsEscaped` and the boundary pair
  `TestAsciiArtHandler_RejectsOversizedText` /
  `TestAsciiArtHandler_AcceptsMaxLengthText`.

## 2026-07-23

### Changed
- Static file server 404s now render the styled `templates/error.html` page
  instead of `http.FileServer`'s built-in plain-text "404 page not found"
  response.
- Exported the shared error renderer `renderError` → `RenderError` in
  `internal/handlers/home.go` so it can be reused outside the `handlers`
  package. Updated all call sites in `home.go`, `asciiart.go`, and
  `download.go`.
- Self-hosted the web fonts (Geist, Instrument Serif, JetBrains Mono). The
  latin and latin-ext `.woff2` files now live in `static/fonts/` and are
  declared in the new `static/css/fonts.css`, replacing the Google Fonts
  `<link>` tags in `templates/index.html` and `templates/error.html`. This
  keeps every resource first-party so the Content-Security-Policy needs no
  external hosts.
- Renamed the Go module from `ascii-art` to `ascii-art-web` in `go.mod`, so the
  module path matches the project. The import paths in `main.go`,
  `internal/handlers/asciiart.go` and `internal/handlers/download.go` were
  updated to match; no other code changed.
- Dropped the leftover task names from the build artifacts. The Docker image's
  `org.opencontainers.image.title` label is now "Ascii Art Web" (was "Ascii Art
  Web Dockerize"), and the compiled binary is `ascii-art-web` (was
  `ascii-art-web-dockerize`) in both the `go build -o` flag and the `CMD` that
  runs it.
- `README.md` follows the same rename: the Docker section builds and runs
  `ascii-art-web` instead of `ascii-art-web-export-file`, and the ASCII banner
  at the top of the file — which spelled out the old task name — was
  regenerated from the shorter one.

### Added
- `middleware.InterceptNotFound` in `middleware/middleware.go`: wraps a handler
  and, when it responds with a 404, suppresses the default response body,
  clears the stale `Content-Type` header, and delegates to a custom not-found
  handler. Wired around the static file server in `main.go` to render
  `error.html` via `handlers.RenderError`.
- `Content-Security-Policy` response header in `middleware.SecureHeaders`. Every
  resource is locked to `'self'` (`script-src`, `font-src`, `connect-src`,
  `form-action`, plus `frame-ancestors 'none'`, `base-uri 'self'`,
  `object-src 'none'`); `style-src` also allows `'unsafe-inline'` so the colored
  ASCII output's inline `style="color:…"` attributes keep rendering.
