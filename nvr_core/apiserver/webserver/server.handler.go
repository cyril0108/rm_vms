package webserver

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"nvr_core/utils"
)

func ServeWeb(mux *http.ServeMux, config *utils.Config) {

	webfolder := http.Dir("./web")
	fs := http.FileServer(webfolder)

	// Serve config.js
	mux.HandleFunc("/web/config.js", HandleGenerateConfig(GenerateConfig(config)))

	// mux.Handle("/web/", fs)

	// SPA Route & Static Assets
	mux.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {

		reqPath := strings.TrimPrefix(r.URL.Path, "/web")

		// Ensure the path always starts with a slash for http.Dir relative resolution
		if !strings.HasPrefix(reqPath, "/") {
			reqPath = "/" + reqPath
		}

		log.Printf("Serving : %s -> %s", r.URL.Path, reqPath)

		// ALWAYS use the raw r.URL.Path. Stripping the slash [1:] can break 
		// http.Dir's internal absolute path resolution on some operating systems.
		f, err := webfolder.Open(reqPath)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()

			// If it's a real file (not a folder like /assets/), serve it
			if err == nil && !stat.IsDir() {
				// StripPrefix safely rewrites r.URL.Path for the underlying FileServer
				http.StripPrefix("/web", fs).ServeHTTP(w, r)
				return
			}
		}

		// --- Assets Query Check ---
		// If the browser is asking for a specific asset file (.js, .css, .png, etc.) 
		// and it was NOT found on disk, we MUST return a 404. 
		ext := filepath.Ext(reqPath)
		if ext != "" && ext != ".html" {
			http.NotFound(w, r)
			return
		}

		// --- SERVE INDEX.HTML ---
		log.Printf("Serving Vue Router fallback for: %s", r.URL.Path)
		index, err := webfolder.Open("/index.html")
		if err != nil {
			// If index is nil, calling index.Close() or index.(io.ReadSeeker) will PANIC and crash Go.
			// We catch the error safely here.
			http.Error(w, "index.html missing from ./web folder", http.StatusNotFound)
			return
		}
		defer index.Close()

		http.ServeContent(w, r, "index.html", time.Now(), index.(io.ReadSeeker))
	})

}