package e2e

import (
	"net/http"
	"testing"
)

func TestListNoCacheHeader(t *testing.T) {
	resp, err := http.Get(fun.Conf.Server.HTTPAddress() + "/v1/tasks")

	if err != nil {
		panic(err)
	}

	log.Debug("HEADERS", resp.Header)

	cch := resp.Header["Cache-Control"]
	if len(cch) < 1 || cch[0] != "no-store" {
		t.Error("Expected cache-control: no-store from List endpoint")
	}
}

func TestGetNoCacheHeader(t *testing.T) {
	resp, err := http.Get(fun.Conf.Server.HTTPAddress() + "/v1/tasks/1")

	if err != nil {
		panic(err)
	}

	log.Debug("HEADERS", resp.Header)

	cch := resp.Header["Cache-Control"]
	if len(cch) < 1 || cch[0] != "no-store" {
		t.Error("Expected cache-control: no-store from Get endpoint")
	}
}
