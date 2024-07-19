package server

import (
	"embed"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed web/*
var EmbedFiles embed.FS

func handleStaticFiles(w http.ResponseWriter, r *http.Request) {
	fname := r.URL.Path[1:] // remove first slash
	if strings.HasPrefix(fname, "templates/") {
		http.NotFound(w, r)
		return
	}

	if fname == "" {
		fname = "index.html"
	}

	file, err := EmbedFiles.ReadFile("web/" + fname)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	fext := filepath.Ext(fname)[1:]
	switch fext {
	case "js":
		w.Header().Set("Content-Type", "text/javascript")
	default:
		w.Header().Set("Content-Type", "text/"+fext)
	}

	w.Write(file)
}
