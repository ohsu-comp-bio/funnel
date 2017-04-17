package task

import (
	"bytes"
	"funnel/proto/tes"
	"github.com/golang/protobuf/jsonpb"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(mux http.Handler) *httptest.Server {
	// Start test server
	lis, err := net.Listen("tcp", ":20001")
	if err != nil {
		panic(err)
	}
	ts := httptest.NewUnstartedServer(mux)
	ts.Listener = lis
	ts.Start()
	return ts
}

func TestAddressTrailingSlash(t *testing.T) {
	c := NewClient("http://funnel.com:8000/")
	if c.address != "http://funnel.com:8000" {
		t.Error("Expected trailing slash to be stripped")
	}
}

func TestGetTask(t *testing.T) {
	var err error

	// Set up test server response
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/jobs/1", func(w http.ResponseWriter, r *http.Request) {
		m := jsonpb.Marshaler{}
		ta := tes.Job{JobID: "test-id"}
		m.Marshal(w, &ta)
	})

	ts := testServer(mux)
	defer ts.Close()

	// Make test client call
	c := NewClient("http://localhost:20001")
	body, err := c.GetTask("1")
	if err != nil {
		t.Fatal(err)
	}
	task := tes.Job{}
	jsonpb.Unmarshal(bytes.NewReader(body), &task)

	if task.JobID != "test-id" {
		log.Debug("RESPONSE", task)
		t.Error("Unexpected response")
	}
}

func TestGetTaskTrailingSlash(t *testing.T) {
	var err error

	// Set up test server response
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/jobs/1", func(w http.ResponseWriter, r *http.Request) {
		m := jsonpb.Marshaler{}
		ta := tes.Job{JobID: "test-id"}
		m.Marshal(w, &ta)
	})

	ts := testServer(mux)
	defer ts.Close()

	// Make test client call
	c := NewClient("http://localhost:20001")
	body, err := c.GetTask("1")
	if err != nil {
		t.Fatal(err)
	}
	task := tes.Job{}
	jsonpb.Unmarshal(bytes.NewReader(body), &task)

	if task.JobID != "test-id" {
		log.Debug("RESPONSE", task)
		t.Error("Unexpected response")
	}
}
