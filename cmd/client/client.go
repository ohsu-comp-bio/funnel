package client

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var log = logger.New("tes http client")

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
		Marshaler: &jsonpb.Marshaler{
			EnumsAsInts:  false,
			EmitDefaults: false,
			Indent:       "\t",
		},
	}
}

// Client represents the HTTP Task client.
type Client struct {
	address   string
	client    *http.Client
	Marshaler *jsonpb.Marshaler
}

// GetTask returns the raw bytes from GET /v1/tasks/{id}
func (c *Client) GetTask(id string, view string) (*tes.Task, error) {
	if !validateView(view) {
		return nil, fmt.Errorf("Invalid view. Must be one of MINIMAL, BASIC, FULL")
	}
	// Send request
	body, err := util.CheckHTTPResponse(c.client.Get(c.address + "/v1/tasks/" + id + "?view=" + view))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &tes.Task{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListTasks returns the result of GET /v1/tasks
func (c *Client) ListTasks(req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	// Build url query parameters
	v := url.Values{}
	addString(v, "project", req.GetProject())
	addString(v, "name_prefix", req.GetNamePrefix())
	addUInt32(v, "page_size", req.GetPageSize())
	addString(v, "page_token", req.GetPageToken())
	addString(v, "view", req.GetView().String())

	// Send request
	u := c.address + "/v1/tasks?" + v.Encode()
	body, err := util.CheckHTTPResponse(c.client.Get(u))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &tes.ListTasksResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateTask POSTs a Task message to /v1/tasks
func (c *Client) CreateTask(msg []byte) (*tes.CreateTaskResponse, error) {
	var err error
	err = isValidTask(msg)
	if err != nil {
		return nil, fmt.Errorf("Not a valid Task message: %s", err)
	}

	// Send request
	r := bytes.NewReader(msg)
	u := c.address + "/v1/tasks"
	body, err := util.CheckHTTPResponse(c.client.Post(u, "application/json", r))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &tes.CreateTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CancelTask POSTs to /v1/tasks/{id}:cancel
func (c *Client) CancelTask(id string) (*tes.CancelTaskResponse, error) {
	u := c.address + "/v1/tasks/" + id + ":cancel"
	body, err := util.CheckHTTPResponse(c.client.Post(u, "application/json", nil))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &tes.CancelTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetServiceInfo returns result of GET /v1/tasks/service-info
func (c *Client) GetServiceInfo() (*tes.ServiceInfo, error) {
	u := c.address + "/v1/tasks/service-info"
	body, err := util.CheckHTTPResponse(c.client.Get(u))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &tes.ServiceInfo{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// WaitForTask polls /v1/tasks/{id} for each Id provided and returns
// once all tasks are in a terminal state.
func (c *Client) WaitForTask(taskIDs ...string) error {
	for range time.NewTicker(time.Second * 2).C {
		done := false
		for _, id := range taskIDs {
			r, err := c.GetTask(id, "MINIMAL")
			if err != nil {
				return err
			}
			switch r.State {
			case tes.State_COMPLETE:
				done = true
			case tes.State_ERROR, tes.State_SYSTEM_ERROR, tes.State_CANCELED:
				errMsg := fmt.Sprintf("Task %s exited with state %s", id, r.State.String())
				return errors.New(errMsg)
			default:
				done = false
			}
		}
		if done {
			return nil
		}
	}
	return nil
}

func isValidTask(b []byte) error {
	var js tes.Task
	err := jsonpb.UnmarshalString(string(b), &js)
	if err != nil {
		return err
	}
	verr := tes.Validate(&js)
	if verr != nil {
		return verr
	}
	return nil
}

func validateView(s string) bool {
	return s == "MINIMAL" || s == "BASIC" || s == "FULL"
}
