package webdash

import (
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
)

var fs = &assetfs.AssetFS{
	Asset:     Asset,
	AssetDir:  AssetDir,
	AssetInfo: AssetInfo,
	Prefix:    "webdash",
}
var index = MustAsset("webdash/index.html")

// FileServer provides access to the bundled web assets (HTML, CSS, etc)
// via an http.Handler
func FileServer() http.Handler {
	return http.FileServer(fs)
}

// RootHandler returns an http handler which always responds with the /index.html file.
func RootHandler() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write(index)
	})
}
