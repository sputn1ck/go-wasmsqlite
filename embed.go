package wasmsqlite

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
)

// Embedded SQLite WASM files for browser usage
var (
	//go:embed assets/sqlite3.wasm
	SQLite3WASM []byte

	//go:embed assets/sqlite3.js
	SQLite3JS string

	//go:embed assets/sqlite3-worker1.js
	SQLite3Worker1JS string

	//go:embed assets/sqlite3-worker1-promiser.js
	SQLite3Worker1PromiserJS string

	//go:embed assets/sqlite3-opfs-async-proxy.js
	SQLite3OPFSAsyncProxyJS string
)

// AssetHandler serves embedded assets over HTTP
type AssetHandler struct {
	// CustomHeaders allows adding custom headers like CORS
	CustomHeaders map[string]string
}

// ServeHTTP implements http.Handler for serving embedded assets
func (h *AssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add custom headers if provided
	for key, value := range h.CustomHeaders {
		w.Header().Set(key, value)
	}

	// Determine content type and serve appropriate asset
	switch r.URL.Path {
	case "/sqlite3.wasm":
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(SQLite3WASM)
	case "/sqlite3.js":
		w.Header().Set("Content-Type", "application/javascript")
		io.WriteString(w, SQLite3JS)
	case "/sqlite3-worker1.js":
		w.Header().Set("Content-Type", "application/javascript")
		io.WriteString(w, SQLite3Worker1JS)
	case "/sqlite3-worker1-promiser.js":
		w.Header().Set("Content-Type", "application/javascript")
		io.WriteString(w, SQLite3Worker1PromiserJS)
	case "/sqlite3-opfs-async-proxy.js":
		w.Header().Set("Content-Type", "application/javascript")
		io.WriteString(w, SQLite3OPFSAsyncProxyJS)
	default:
		http.NotFound(w, r)
	}
}

// NewAssetHandler creates a new asset handler with CORS headers for development
func NewAssetHandler() *AssetHandler {
	return &AssetHandler{
		CustomHeaders: map[string]string{
			"Cross-Origin-Embedder-Policy": "require-corp",
			"Cross-Origin-Opener-Policy":   "same-origin",
		},
	}
}

// ExtractAssets writes all embedded assets to a directory
// This is useful for users who want to serve assets separately
func ExtractAssets(dir string) error {
	// Implementation for extracting assets to filesystem
	// This would write each embedded file to the specified directory
	return fmt.Errorf("not implemented yet")
}