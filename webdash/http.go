package webdash

import (
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/ohsu-comp-bio/funnel/logger"
)

var fs = &assetfs.AssetFS{
	Asset:     Asset,
	AssetDir:  AssetDir,
	AssetInfo: AssetInfo,
}

var index = MustAsset("index.html")

// FileServer provides access to the bundled web assets (HTML, CSS, etc)
// via an http.Handler
func FileServer() http.Handler {
	return http.FileServer(fs)
}

// RootHandler returns an http handler which always responds with the /index.html file.
func RootHandler(log *logger.Logger) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		_, err := resp.Write(index)
		if err != nil {
			log.Error("HTTP handler error", "error", err, "url", req.URL)
		}
	})
}
