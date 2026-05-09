package handler

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type AssetHandler struct {
	BasePath string
}

func NewAssetHandler(basePath string) *AssetHandler {
	return &AssetHandler{BasePath: basePath}
}

func (h *AssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if !strings.HasPrefix(path, "/assets/") {
		http.NotFound(w, r)
		return
	}

	relPath := strings.TrimPrefix(path, "/assets/")
	if relPath == "" || strings.Contains(relPath, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(h.BasePath, "assets", filepath.FromSlash(relPath))
	fullPath = filepath.Clean(fullPath)
	assetsDir := filepath.Join(h.BasePath, "assets")
	if !strings.HasPrefix(fullPath, assetsDir) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	ext := filepath.Ext(fullPath)
	if ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
	}

	http.ServeFile(w, r, fullPath)
}
