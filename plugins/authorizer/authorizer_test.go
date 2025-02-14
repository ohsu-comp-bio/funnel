package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"example.com/auth"
	"example.com/tes"
)

func readExampleTask(filename string) tes.TesTask {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var task tes.TesTask
	err = json.Unmarshal(data, &task)
	if err != nil {
		return tes.TesTask{}
	}

	return task
}

func TestValidUser(t *testing.T) {
	// Setup test file
	os.Setenv("EXAMPLE_USERS", "example-users.csv")

	// Read example task
	task := readExampleTask("../example-tasks/hello-world.json")

	// Create an Authorization Header to pass to the authorizer
	authHeader := http.Header{
		"Authorization": []string{"Bearer Alyssa P. Hacker"},
	}

	authenticator := ExampleAuthorizer{}
	resp, err := authenticator.Authorize(authHeader, task)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := auth.Auth{
		User:  "Alyssa P. Hacker",
		Token: "<Alyssa's Secret>",
	}

	if resp.User != expected.User || resp.Token != expected.Token {
		t.Errorf("Expected (%s, %s), got (%s, %s)",
			expected.User, expected.Token,
			resp.User, resp.Token)
	}
}

func TestInvalidUser(t *testing.T) {
	// Setup test file
	os.Setenv("EXAMPLE_USERS", "example-users.csv")

	// Read example task
	task := readExampleTask("../example-tasks/hello-world.json")

	// Create an Authorization Header to pass to the authorizer
	authHeader := http.Header{
		"Authorization": []string{"Bearer Foo"},
	}

	authenticator := ExampleAuthorizer{}
	_, err := authenticator.Authorize(authHeader, task)

	if err == nil {
		t.Fatalf("Expected 401 Unauthorized error, got %v", err)
	}
}

func TestInvalidAuthHeader(t *testing.T) {
	// Setup test file
	os.Setenv("EXAMPLE_USERS", "example-users.csv")

	// Read example task
	task := readExampleTask("../example-tasks/hello-world.json")

	// Create an Authorization Header to pass to the authorizer
	authHeader := http.Header{
		"Authorization": []string{"Basic Alyssa P. Hacker"},
	}

	authenticator := ExampleAuthorizer{}
	_, err := authenticator.Authorize(authHeader, task)

	if err == nil {
		t.Fatalf("Expected 401 Unauthorized error, got %v", err)
	}
}

func TestMissingAuthHeader(t *testing.T) {
	// Setup test file
	os.Setenv("EXAMPLE_USERS", "example-users.csv")

	// Read example task
	task := readExampleTask("../example-tasks/hello-world.json")

	// Create an Authorization Header to pass to the authorizer
	authHeader := http.Header{}

	authenticator := ExampleAuthorizer{}
	_, err := authenticator.Authorize(authHeader, task)

	if err == nil {
		t.Fatalf("Expected 401 Unauthorized error, got %v", err)
	}
}
