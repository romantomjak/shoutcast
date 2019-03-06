package shoutcast

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertStrings(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRequiredHTTPHeadersArePresent(t *testing.T) {
	var headers http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
	}))
	defer ts.Close()

	Open(ts.URL)

	assertStrings(t, headers.Get("icy-metadata"), "1")
	assertStrings(t, headers.Get("user-agent")[:6], "iTunes")
}
