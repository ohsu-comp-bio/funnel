package plugins

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Define a struct that matches the expected JSON response
type TokenResponse struct {
	User  string `json:"user,omitempty"`
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func TestAuthorizedUser(t *testing.T) {
	auth := AuthorizeImpl{}
	raw, err := auth.Get("example")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parse actual JSON response
	var actual TokenResponse
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Define expected response
	expected := TokenResponse{
		User:  "example",
		Token: "example's secret",
	}

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}

func TestUnauthorizedUser(t *testing.T) {
	auth := AuthorizeImpl{}
	raw, err := auth.Get("error")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parse actual JSON response
	var actual TokenResponse
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Define expected response
	expected := TokenResponse{
		Error: "user 'error' not found",
	}

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}
