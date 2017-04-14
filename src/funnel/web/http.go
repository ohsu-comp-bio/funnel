package web

import (
	"github.com/elazarl/go-bindata-assetfs"
	"net/http"
)

// FileServer provides access to the bundled web assets (HTML, CSS, etc)
// via an http.Handler
func FileServer() http.Handler {
	fs := &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    "web",
	}
	return http.FileServer(fs)
}
