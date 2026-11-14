package main

import (
	"io/fs"
	"net/http"
	"strings"

	"kubelearn"
)

// interfaceHandler serves the web interface compiled into the binary. A build
// without it (a development build) explains how to get one instead.
func interfaceHandler() http.Handler {
	files, ok := kubelearn.Frontend()
	if !ok {
		return http.HandlerFunc(noInterface)
	}
	return spaHandler(files)
}

// spaHandler serves static files, answering any unknown path with the
// application page so a reload on a client-side route still works.
func spaHandler(files fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(files, name); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		page := r.Clone(r.Context())
		page.URL.Path = "/"
		// The page changes with every release; its assets have hashed names.
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, page)
	})
}

func noInterface(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("The KubeLearn backend is running, but this build does not include the web interface.\n\n" +
		"Start the development interface with \"make frontend-start\" and open http://localhost:3000,\n" +
		"or build the complete binary with \"make build\".\n"))
}
