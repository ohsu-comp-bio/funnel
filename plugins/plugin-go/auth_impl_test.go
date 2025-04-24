package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
)

func TestAuthorizedUser(t *testing.T) {
	// Start test HTTP server
	// Here we simply return a hardcoded response similar to that of the test server used in the plugin tests
	// ref: https://github.com/ohsu-comp-bio/funnel-plugins/blob/main/tests/test-server.go
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "example" {
			resp := shared.Response{
				Code: 200,
				Config: &config.Config{
					AmazonS3: &config.AmazonS3Storage{
						AWSConfig: &config.AWSConfig{
							Key:    "key1",
							Secret: "secret1",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := shared.Response{
				Code:    401,
				Message: "User not authorized",
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer ts.Close()

	auth := Authorize{}
	raw, err := auth.Get("example", ts.URL+"/token?user=")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var actual shared.Response
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	expected := shared.Response{
		Code: 200,
		Config: &config.Config{
			AmazonS3: &config.AmazonS3Storage{
				AWSConfig: &config.AWSConfig{
					Key:    "key1",
					Secret: "secret1",
				},
			},
		},
	}

	if actual.Code != expected.Code || actual.Config.AmazonS3.AWSConfig.Key != expected.Config.AmazonS3.AWSConfig.Key {
		t.Errorf("expected %+v, got %+v", expected, actual)
	}
}
