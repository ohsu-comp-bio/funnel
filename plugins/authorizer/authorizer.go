package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"example.com/auth"
	"example.com/plugin"
	"example.com/tes"
	"github.com/golang/gddo/log"
	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
)

type ExampleAuthorizer struct{}

func (ExampleAuthorizer) Hooks() []string {
	return []string{"contents"}
}

func (ExampleAuthorizer) Authorize(authHeader http.Header, task tes.TesTask) (auth.Auth, error) {
	// TOOD: Currently we're just using the first Authorization header in the request
	// How might we support multiple Authorization headers?
	if authHeader == nil {
		return auth.Auth{}, fmt.Errorf("No Authorization header found")
	}

	user := authHeader.Get("Authorization")
	if user == "" {
		return auth.Auth{}, fmt.Errorf("No user found in Authorization header %s", user)
	}

	if !strings.HasPrefix(user, "Bearer ") {
		return auth.Auth{}, fmt.Errorf("Invalid Authorization header: expected 'Bearer <token>', got %s", user)
	}

	user = strings.TrimPrefix(user, "Bearer ")

	creds, err := ExampleAuthorizer{}.getUser(user)
	if err != nil {
		log.Info(context.Background(), "401: User Unauthorized", "user", user)
		return auth.Auth{}, err
	}

	log.Info(context.Background(), "200: User Authorized", "user", creds.User)

	return creds, nil
}

func (ExampleAuthorizer) getUser(user string) (auth.Auth, error) {
	// Check if the user is authorized
	// Read the "internal" User Database
	// Here we're just using a CSV file to represent the list of authorized users
	// A real-world example would use a database or an external service (e.g. OAuth)
	userFile := os.Getenv("EXAMPLE_USERS")
	if userFile == "" {
		log.Info(context.Background(), "EXAMPLE_USERS not set, using default example-users.csv")
		userFile = "authorizer/example-users.csv"
	}

	file, err := os.Open(userFile)
	if err != nil {
		return auth.Auth{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if strings.Contains(line, user) {
			result := strings.Split(line, ",")

			auth := auth.Auth{
				User:  result[0],
				Token: result[1],
			}

			return auth, nil
		}
	}

	return auth.Auth{}, fmt.Errorf("User %s not found", user)
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level: hclog.Warn,
	})

	plugins := map[string]goplugin.Plugin{
		"authorize": &plugin.AuthorizePlugin{
			Impl: ExampleAuthorizer{},
		},
	}

	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins:         plugins,
		Logger:          logger,
	})
}
