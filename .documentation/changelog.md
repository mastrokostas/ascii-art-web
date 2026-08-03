# Changelog

## 2026-08-03

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
