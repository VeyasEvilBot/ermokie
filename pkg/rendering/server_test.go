package rendering

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOverlayHandlerServesHTMLCSSAndHealth(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "output.html"), []byte("<main>ok</main>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "output.css"), []byte("body{color:red}"), 0o600); err != nil {
		t.Fatal(err)
	}

	h := NewOverlayServer(0, dir).Handler()
	cases := []struct {
		path        string
		contentType string
		body        string
	}{
		{path: "/overlay", contentType: "text/html", body: "<main>ok</main>"},
		{path: "/output.css", contentType: "text/css", body: "body{color:red}"},
		{path: "/healthz", contentType: "application/json", body: `{"status":"ok"}`},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
			}
			if got := res.Header().Get("Content-Type"); !strings.HasPrefix(got, tc.contentType) {
				t.Fatalf("Content-Type = %q", got)
			}
			if got := strings.TrimSpace(res.Body.String()); got != tc.body {
				t.Fatalf("body = %q, want %q", got, tc.body)
			}
		})
	}
}

func TestOverlayHandlerReportsMissingRender(t *testing.T) {
	h := NewOverlayServer(0, t.TempDir()).Handler()
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/overlay", nil))
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}
}
