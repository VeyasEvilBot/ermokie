package rendering

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type OverlayServer struct {
	port    int
	dirPath string
}

// Handler returns the overlay HTTP handler. Files are read for every request,
// so an atomic render update is visible immediately without watcher races.
func (o *OverlayServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/overlay", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/overlay", o.serveFile("output.html", "text/html; charset=utf-8"))
	mux.HandleFunc("/output.css", o.serveFile("output.css", "text/css; charset=utf-8"))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	return mux
}

func (o *OverlayServer) serveFile(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// name is supplied only by the fixed /overlay and /output.css routes.
		content, err := os.ReadFile(filepath.Join(o.dirPath, name)) // #nosec G304
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "overlay has not been rendered yet", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "could not read overlay", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodGet {
			_, _ = w.Write(content)
		}
	}
}

func (o *OverlayServer) Serve() error {
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", o.port),
		Handler:           o.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}

func NewOverlayServer(port int, dirPath string) *OverlayServer {
	return &OverlayServer{port: port, dirPath: dirPath}
}
