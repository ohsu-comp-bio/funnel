package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"funnel/proto/tes"
	"github.com/golang/protobuf/jsonpb"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// NewClient returns a new HTTP client for accessing
// Create/List/Get/Cancel Task endpoints. "address" is the address
// of the TES server.
func NewClient(address string) *Client {
	// Strip trailing slash. A quick and dirty fix.
	address = strings.TrimSuffix(address, "/")
	return &Client{
		address: address,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Client represents the HTTP Task client.
type Client struct {
	address string
	client  *http.Client
}

// GetTask returns the raw bytes from GET /v1/tasks/{id}
func (c *Client) GetTask(id string) ([]byte, error) {
	return check(c.client.Get(c.address + "/v1/tasks/" + id))
}

// ListTasks returns the result of GET /v1/tasks
// TODO returning interface{} is weird
func (c *Client) ListTasks() (interface{}, error) {
	body, err := check(c.client.Get(c.address + "/v1/tasks"))
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res["tasks"], nil
}

// CreateTask POSTs a Task message to /v1/tasks
func (c *Client) CreateTask(msg []byte) (string, error) {
	var err error
	err = isTask(msg)
	if err != nil {
		return "", fmt.Errorf("Not a valid Task message: %v", err)
	}

	// Send request
	r := bytes.NewReader(msg)
	u := c.address + "/v1/tasks"
	body, err := check(c.client.Post(u, "application/json", r))

	// Parse response
	resp := tes.CreateTaskResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

// CancelTask POSTs to /v1/tasks/{id}:cancel
func (c *Client) CancelTask(id string) ([]byte, error) {
	u := c.address + "/v1/tasks/" + id + ":cancel"
	return check(c.client.Post(u, "application/json", nil))
}

// check does some basic error handling
// and reads the response body into a byte array
func check(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if (resp.StatusCode / 100) != 2 {
		return nil, fmt.Errorf("[STATUS CODE - %d]\t%s", resp.StatusCode, body)
	}
	return body, nil
}

// TODO replace with proper message validation
func isTask(b []byte) error {
	var js tes.Task
	return jsonpb.UnmarshalString(string(b), &js)
}
