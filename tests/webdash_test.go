package tests

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWebdash(t *testing.T) {
	srv := NewFunnel(NewConfig())
	srv.Start()
	defer srv.Stop()

	// Get the webdash health check endpoint
	address := "http://localhost:" + srv.Conf.HTTPPort
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(address + "/health.html")
	if err != nil {
		t.Fatal(err)
	}

	// Get the response body
	defer resp.Body.Close()
	b, berr := ioutil.ReadAll(resp.Body)
	if berr != nil {
		t.Fatal(berr)
	}

	if string(b) != "OK\n" {
		logger.Error("webdash check", string(b))
		t.Fatal("Webdash health check fail")
	}
}
