package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"funnel/proto/tes"
	"io/ioutil"
	"net/http"
	"strings"
)

// NewClient returns a new HTTP client for accessing
// Create/List/Get/Cancel Task endpoints. "address" is the address
// of the TES server.
func NewClient(address string) *Client {
	// Strip trailing slash. A quick and dirty fix.
	address = strings.TrimSuffix(address, "/")
	return &Client{address, &http.Client{}}
}

// Client represents the HTTP Task client.
type Client struct {
	address string
	client  *http.Client
}

// GetTask returns the raw bytes from GET /v1/jobs/{id}
func (c *Client) GetTask(id string) ([]byte, error) {
	return check(c.client.Get(c.address + "/v1/jobs/" + id))
}

// ListTasks returns the result of GET /v1/jobs
// TODO returning interface{} is weird
func (c *Client) ListTasks() (interface{}, error) {
	body, err := check(c.client.Get(c.address + "/v1/jobs"))
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	return res["jobs"], nil
}

// CreateTask POSTs a Task message to /v1/jobs
func (c *Client) CreateTask(msg []byte) (string, error) {
	if !isTask(msg) {
		return "", fmt.Errorf("Not a valid Job message")
	}
	var err error

	// Send request
	r := bytes.NewReader(msg)
	u := c.address + "/v1/jobs"
	body, err := check(c.client.Post(u, "application/json", r))

	// Parse response
	jobID := tes.JobID{}
	err = json.Unmarshal(body, &jobID)
	if err != nil {
		return "", err
	}
	return jobID.Value, nil
}

// CancelTask DELETEs to /v1/jobs/{id}
func (c *Client) CancelTask(id string) ([]byte, error) {
	u := c.address + "/v1/jobs/" + id
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	return check(resp, err)
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

func isTask(b []byte) bool {
	var js tes.Job
	return json.Unmarshal(b, &js) == nil
}
