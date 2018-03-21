package tes

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
	"github.com/ohsu-comp-bio/funnel/util"
	"golang.org/x/net/context"
)

// NewClient returns a new HTTP client for accessing
// Create/List/Get/Cancel Task endpoints. "address" is the address
// of the TES server.
func NewClient(address string) (*Client, error) {
	user := os.Getenv("FUNNEL_SERVER_USER")
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
		Marshaler: &Marshaler,
		User:      user,
		Password:  password,
	}, nil
}

// Client represents the HTTP Task client.
type Client struct {
	address   string
	client    *http.Client
	Marshaler *jsonpb.Marshaler
	User      string
	Password  string
}

// GetTask returns the raw bytes from GET /v1/tasks/{id}
func (c *Client) GetTask(ctx context.Context, req *GetTaskRequest) (*Task, error) {
	// Send request
	u := c.address + "/v1/tasks/" + req.Id + "?view=" + req.View.String()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth(c.User, c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &Task{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ListTasks returns the result of GET /v1/tasks
func (c *Client) ListTasks(ctx context.Context, req *ListTasksRequest) (*ListTasksResponse, error) {
	// Build url query parameters
	v := url.Values{}
	addUInt32(v, "page_size", req.GetPageSize())
	addString(v, "page_token", req.GetPageToken())
	addString(v, "view", req.GetView().String())

	if req.GetState() != Unknown {
		addString(v, "state", req.State.String())
	}

	for key, val := range req.Tags {
		v.Add(fmt.Sprintf("tags[%s]", key), val)
	}

	// Send request
	u := c.address + "/v1/tasks?" + v.Encode()
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth(c.User, c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}
	// Parse response
	resp := &ListTasksResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateTask POSTs a Task message to /v1/tasks
func (c *Client) CreateTask(ctx context.Context, task *Task) (*CreateTaskResponse, error) {
	verr := Validate(task)
	if verr != nil {
		return nil, fmt.Errorf("invalid task message: %v", verr)
	}

	var b bytes.Buffer
	err := Marshaler.Marshal(&b, task)
	if err != nil {
		return nil, fmt.Errorf("error marshaling task message: %v", err)
	}

	// Send request
	u := c.address + "/v1/tasks"
	hreq, _ := http.NewRequest("POST", u, &b)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	hreq.SetBasicAuth(c.User, c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &CreateTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CancelTask POSTs to /v1/tasks/{id}:cancel
func (c *Client) CancelTask(ctx context.Context, req *CancelTaskRequest) (*CancelTaskResponse, error) {
	u := c.address + "/v1/tasks/" + req.Id + ":cancel"
	hreq, _ := http.NewRequest("POST", u, nil)
	hreq.WithContext(ctx)
	hreq.Header.Add("Content-Type", "application/json")
	hreq.SetBasicAuth(c.User, c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &CancelTaskResponse{}
	err = jsonpb.UnmarshalString(string(body), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetServiceInfo returns result of GET /v1/tasks/service-info
func (c *Client) GetServiceInfo(ctx context.Context, req *ServiceInfoRequest) (*ServiceInfo, error) {
	u := c.address + "/v1/tasks/service-info"
	hreq, _ := http.NewRequest("GET", u, nil)
	hreq.WithContext(ctx)
	hreq.SetBasicAuth(c.User, c.Password)
	body, err := util.CheckHTTPResponse(c.client.Do(hreq))
	if err != nil {
		return nil, err
	}

	// Parse response
	resp := &ServiceInfo{}
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
			r, err := c.GetTask(ctx, &GetTaskRequest{
				Id:   id,
				View: TaskView_MINIMAL,
			})
			if err != nil {
				return err
			}
			switch r.State {
			case State_COMPLETE:
				done = true
			case State_EXECUTOR_ERROR, State_SYSTEM_ERROR, State_CANCELED:
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
