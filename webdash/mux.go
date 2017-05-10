package webdash

import "net/http"

func Handler() *http.ServeMux {
	// Static files are bundled into web-dashboard
	fs := FileServer()
	// Set up URL path handlers
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	return mux
}
