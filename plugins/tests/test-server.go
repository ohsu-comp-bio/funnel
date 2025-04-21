package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
)

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/token", tokenHandler)

	// Currently hardcoding the endpoint of the token service
	// TODO: This should be made configurable similar to the plugin (see plugin/auth_impl.go)
	fmt.Println("Server is running on http://0.0.0.0:8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// Handler for root endpoint
func indexHandler(w http.ResponseWriter, r *http.Request) {
	resp := shared.Response{
		Code:    http.StatusOK,
		Message: "Hello, world! To get a token, send a GET request to /token?user=[USER]",
	}
	json.NewEncoder(w).Encode(resp)
}

// Handler for retrieving user tokens
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received token request:", r)

	// Load users from the CSV file
	userDB, err := loadUsers("example-users.csv")
	if err != nil {
		fmt.Println("Error loading users:", err)
		return
	}

	user := r.URL.Query().Get("user")

	// No user provided in the query (Bad Request: 400)
	if user == "" {
		resp := shared.Response{
			Code:    http.StatusBadRequest,
			Message: "User is required",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	token, found := userDB[user]

	if found {
		// User found (OK: 200)
		c := config.Config{}
		c.AmazonS3.AWSConfig.Key = token.AmazonS3.Key
		c.AmazonS3.AWSConfig.Secret = token.AmazonS3.Secret

		resp := shared.Response{
			Code:   http.StatusOK,
			Config: &c,
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		// User not found (Unauthorized: 401)
		resp := shared.Response{
			Code:    http.StatusUnauthorized,
			Message: "User not authorized",
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// Load user tokens from the CSV file
func loadUsers(filename string) (map[string]config.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	userDB := make(map[string]config.Config)
	mutex := sync.RWMutex{}
	for i, row := range records {
		if i == 0 {
			continue // Skip header
		}
		mutex.Lock()
		userDB[row[0]] = config.Config{
			AmazonS3: config.AmazonS3Storage{
				AWSConfig: config.AWSConfig{
					Key:    row[1],
					Secret: row[2],
				},
			},
		}
		mutex.Unlock()
	}
	return userDB, nil
}
