package rendering

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type OverlayServer struct {
	port      int
	dirPath   string
	htmlBytes []byte
	cssBytes  []byte
	mu        sync.RWMutex
	watcher   *fsnotify.Watcher
}

func (o *OverlayServer) Serve() {
	o.loadFiles()
	go o.watchFiles()

	r := http.NewServeMux()
	r.HandleFunc("/overlay", o.handleOverlay)
	http.ListenAndServe(fmt.Sprintf("localhost:%d", o.port), r)
}

func (o *OverlayServer) loadFiles() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	var err error
	o.htmlBytes, err = os.ReadFile(filepath.Join(o.dirPath, "output.html"))
	if err != nil {
		return err
	}

	o.cssBytes, err = os.ReadFile(filepath.Join(o.dirPath, "output.css"))
	if err != nil {
		return err
	}

	return nil
}

func (o *OverlayServer) watchFiles() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Watcher error:", err)
		return
	}
	defer watcher.Close()

	o.watcher = watcher
	watcher.Add(o.dirPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			name := filepath.Base(event.Name)
			if (name == "output.html" || name == "output.css") && event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("File changed:", name)
				o.loadFiles()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}
}

func (o *OverlayServer) handleOverlay(w http.ResponseWriter, r *http.Request) {
	o.mu.RLock()
	html := o.htmlBytes
	css := o.cssBytes
	o.mu.RUnlock()

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
	fmt.Fprintf(w, "<style>%s</style>", css)
}

func NewOverlayServer(port int, dirPath string) *OverlayServer {
	return &OverlayServer{
		port:    port,
		dirPath: dirPath,
	}
}
