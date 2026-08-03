package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveThroughSecureHeaders runs a trivial handler behind SecureHeaders and
// returns both the recorded response and the nonce the handler saw on its
// request context.
func serveThroughSecureHeaders(t *testing.T) (*httptest.ResponseRecorder, string) {
	t.Helper()

	var observed_nonce string
	inner_handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed_nonce = NonceFromContext(r.Context())
	})

	request := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	SecureHeaders(inner_handler).ServeHTTP(recorder, request)

	return recorder, observed_nonce
}

// TestSecureHeaders_CspHasNonceNotUnsafeInline checks that style-src is carried
// by a nonce rather than 'unsafe-inline'. This is the whole point of moving the
// art color out of inline style attributes, so it is worth pinning down.
func TestSecureHeaders_CspHasNonceNotUnsafeInline(t *testing.T) {
	recorder, _ := serveThroughSecureHeaders(t)

	policy := recorder.Header().Get("Content-Security-Policy")
	if policy == "" {
		t.Fatal("no Content-Security-Policy header was set")
	}
	if strings.Contains(policy, "unsafe-inline") {
		t.Errorf("policy still allows inline styles: %q", policy)
	}
	if !strings.Contains(policy, "style-src 'self' 'nonce-") {
		t.Errorf("policy is missing a style-src nonce: %q", policy)
	}
}

// TestSecureHeaders_NonceReachesHandler checks that the nonce advertised in the
// policy is the same one handlers read from the request context. If these ever
// diverge the inline <style> block is silently blocked by the browser.
func TestSecureHeaders_NonceReachesHandler(t *testing.T) {
	recorder, observed_nonce := serveThroughSecureHeaders(t)

	if observed_nonce == "" {
		t.Fatal("handler saw no nonce on the request context")
	}

	policy := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(policy, "'nonce-"+observed_nonce+"'") {
		t.Errorf("nonce %q from the context is not the one in the policy: %q", observed_nonce, policy)
	}
}

// TestSecureHeaders_NonceIsPerRequest checks that a fresh nonce is minted for
// every request. A reused nonce would be worthless — injected markup could just
// quote it.
func TestSecureHeaders_NonceIsPerRequest(t *testing.T) {
	_, first_nonce := serveThroughSecureHeaders(t)
	_, second_nonce := serveThroughSecureHeaders(t)

	if first_nonce == second_nonce {
		t.Errorf("nonce was reused across requests: %q", first_nonce)
	}
}

// TestSecureHeaders_HstsIsOptIn checks that Strict-Transport-Security stays off
// unless it is explicitly enabled. Sending it from a plain-HTTP development
// origin would pin that origin to HTTPS in the developer's browser.
func TestSecureHeaders_HstsIsOptIn(t *testing.T) {
	recorder, _ := serveThroughSecureHeaders(t)

	if header := recorder.Header().Get("Strict-Transport-Security"); header != "" {
		t.Errorf("HSTS was sent without ENABLE_HSTS being set: %q", header)
	}
}

// TestSecureHeaders_HstsWhenEnabled checks the opposite side of that switch.
func TestSecureHeaders_HstsWhenEnabled(t *testing.T) {
	t.Setenv("ENABLE_HSTS", "true")

	recorder, _ := serveThroughSecureHeaders(t)

	header := recorder.Header().Get("Strict-Transport-Security")
	if !strings.Contains(header, "max-age=31536000") {
		t.Errorf("HSTS not set with ENABLE_HSTS=true: %q", header)
	}
}
