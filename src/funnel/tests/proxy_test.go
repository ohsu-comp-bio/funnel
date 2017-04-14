package tests

import (
	"funnel/logger"
	"net/http"
	"testing"
)

func TestNoCacheHeader(t *testing.T) {
	srv := NewFunnel(NewConfig())
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
