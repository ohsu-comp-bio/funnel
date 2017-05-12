package webdash

import "net/http"

// Handler handles static webdash files
func Handler() *http.ServeMux {
	// Static files are bundled into webdash
	fs := FileServer()
	// Set up URL path handlers
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	return mux
}
