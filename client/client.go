package client

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
)

// NewClient returns a new HTTP client for accessing
// Create/List/Get/Cancel Task endpoints. "address" is the address
// of the TES server.
func NewClient(address string) (*Client, error) {
	password := os.Getenv("FUNNEL_SERVER_PASSWORD")

	re := regexp.MustCompile("^(.+://)?(.[^/]+)(.+)?$")
	endpoint := re.ReplaceAllString(address, "$1$2")

	reScheme := regexp.MustCompile("^.+://")
	if reScheme.MatchString(endpoint) {
		if !strings.HasPrefix(endpoint, "http") {
			return nil, fmt.Errorf("invalid protocol: '%s'; expected: 'http://' or 'https://'", reScheme.FindString(endpoint))
		}
	} else {
		endpoint = "http://" + endpoint
	}

	return &Client{
		address: endpoint,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		Marshaler: &tes.Marshaler,
		Password:  password,
	}, nil
}

// Client represents the HTTP Task client.
type Client struct {
	address   string
	client    *http.Client
	Marshaler *jsonpb.Marshaler
	Password  string
}

// GetTask returns the raw bytes from GET /v1/tasks/{id}
func (c *Client) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	// Send request
	u := c.address + "/v1/tasks/" + req.Id + "?view=" + req.View.String()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth("funnel", c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
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
func (c *Client) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	// Build url query parameters
	v := url.Values{}
	addUInt32(v, "page_size", req.GetPageSize())
	addString(v, "page_token", req.GetPageToken())
	addString(v, "view", req.GetView().String())

	if req.GetState() != tes.Unknown {
		addString(v, "state", req.State.String())
	}

	// Send request
	u := c.address + "/v1/tasks?" + v.Encode()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth("funnel", c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
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
func (c *Client) CreateTask(ctx context.Context, task *tes.Task) (*tes.CreateTaskResponse, error) {
	verr := tes.Validate(task)
	if verr != nil {
		return nil, fmt.Errorf("invalid task message: %v", verr)
	}

	var b bytes.Buffer
	err := tes.Marshaler.Marshal(&b, task)
	if err != nil {
		return nil, fmt.Errorf("error marshaling task message: %v", err)
	}

	// Send request
	u := c.address + "/v1/tasks"
	hreq, _ := http.NewRequest("POST", u, &b)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	hreq.SetBasicAuth("funnel", c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
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
func (c *Client) CancelTask(ctx context.Context, req *tes.CancelTaskRequest) (*tes.CancelTaskResponse, error) {
	u := c.address + "/v1/tasks/" + req.Id + ":cancel"
	hreq, _ := http.NewRequest("POST", u, nil)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	hreq.SetBasicAuth("funnel", c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
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
func (c *Client) GetServiceInfo(ctx context.Context, req *tes.ServiceInfoRequest) (*tes.ServiceInfo, error) {
	u := c.address + "/v1/tasks/service-info"
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth("funnel", c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
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
func (c *Client) WaitForTask(ctx context.Context, taskIDs ...string) error {
	for range time.NewTicker(time.Second * 2).C {
		done := false
		for _, id := range taskIDs {
			r, err := c.GetTask(ctx, &tes.GetTaskRequest{
				Id:   id,
				View: tes.TaskView_MINIMAL,
			})
			if err != nil {
				return err
			}
			switch r.State {
			case tes.State_COMPLETE:
				done = true
			case tes.State_EXECUTOR_ERROR, tes.State_SYSTEM_ERROR, tes.State_CANCELED:
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
