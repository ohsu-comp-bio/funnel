package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
)

var host = "http://localhost:8080/token?user="

func TestAuthorizedUser(t *testing.T) {
	auth := Authorize{}
	raw, err := auth.Get("example", host, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parse actual JSON response
	var actual shared.Response
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Define expected response
	c := config.Config{}
	c.AmazonS3.AWSConfig.Key = "key1"
	c.AmazonS3.AWSConfig.Secret = "secret1"

	expected := shared.Response{
		Code:   200,
		Config: &c,
	}

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}

func TestUnauthorizedUser(t *testing.T) {
	auth := Authorize{}
	raw, err := auth.Get("error", host, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Parse actual JSON response
	var actual shared.Response
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	// Define expected response
	expected := shared.Response{
		Code:    401,
		Message: "User not authorized",
	}

	// Compare using reflect.DeepEqual
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}
