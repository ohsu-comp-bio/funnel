package server

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/ohsu-comp-bio/funnel/tests"
)

func TestWebdash(t *testing.T) {
	tests.SetLogOutput(log, t)
	// Get the webdash health check endpoint
	address := fun.Conf.Server.HTTPAddress()
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(address + "/static/health.html")
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
		t.Fatal("Webdash health check fail", string(b))
	}
}
