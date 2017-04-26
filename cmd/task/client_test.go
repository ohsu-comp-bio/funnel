package task

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
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
	mux.HandleFunc("/v1/tasks/1", func(w http.ResponseWriter, r *http.Request) {
		m := jsonpb.Marshaler{}
		ta := tes.Task{Id: "test-id"}
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
	task := tes.Task{}
	jsonpb.Unmarshal(bytes.NewReader(body), &task)

	if task.Id != "test-id" {
		log.Debug("RESPONSE", task)
		t.Error("Unexpected response")
	}
}

func TestGetTaskTrailingSlash(t *testing.T) {
	var err error

	// Set up test server response
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks/1", func(w http.ResponseWriter, r *http.Request) {
		m := jsonpb.Marshaler{}
		ta := tes.Task{Id: "test-id"}
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
	task := tes.Task{}
	jsonpb.Unmarshal(bytes.NewReader(body), &task)

	if task.Id != "test-id" {
		log.Debug("RESPONSE", task)
		t.Error("Unexpected response")
	}
}

func TestClientTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestClientTimeout.")
	}

	// Set up test server response
	mux := http.NewServeMux()
	ch := make(chan struct{})
	mux.HandleFunc("/v1/tasks/1", func(w http.ResponseWriter, r *http.Request) {
		<-ch
	})

	ts := testServer(mux)
	defer ts.Close()

	c := NewClient("http://localhost:20001")
	_, err := c.GetTask("1")
	close(ch)
	if err == nil {
		t.Fatal("Request did not timeout.")
	}
}
