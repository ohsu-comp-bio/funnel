package tests

import (
	"funnel/logger"
	server_mocks "funnel/server/mocks"
	"net/http"
	"testing"
)

func TestNoCacheHeader(t *testing.T) {
	srv := server_mocks.NewServer(server_mocks.NewConfig())
	srv.Start()
	defer srv.Stop()

	resp, err := http.Get(srv.Conf.HTTPAddress() + "/v1/jobs")

	if err != nil {
		panic(err)
	}

	logger.Debug("HEADERS", resp.Header)

	cch := resp.Header["Cache-Control"]
	if len(cch) < 1 || cch[0] != "no-store" {
		t.Error("Expected cache-control: no-store from List endpoint")
	}
}
