# Changelog

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
