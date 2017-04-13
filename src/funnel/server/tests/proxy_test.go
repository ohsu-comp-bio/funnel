package tests

import (
	"funnel/logger"
	server_mocks "funnel/server/mocks"
	"net/http"
	"testing"
)

func TestNoCacheHeader(t *testing.T) {
	conf := server_mocks.NewMockServerConfig()
	srv := server_mocks.MockServerFromConfig(conf)
	defer srv.Close()

	resp, err := http.Get(conf.HTTPAddress() + "/v1/jobs")

	if err != nil {
		panic(err)
	}

	logger.Debug("HEADERS", resp.Header)

	cch := resp.Header["Cache-Control"]
	if len(cch) < 1 || cch[0] != "no-store" {
		t.Error("Expected cache-control: no-store from List endpoint")
	}
}
