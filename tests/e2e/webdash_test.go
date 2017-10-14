package e2e

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWebdash(t *testing.T) {
	setLogOutput(t)
	// Get the webdash health check endpoint
	address := fun.Conf.Server.HTTPAddress()
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
		t.Fatal("Webdash health check fail", string(b))
	}
}
