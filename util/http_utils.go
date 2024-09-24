package util

import (
	"fmt"
	"io"
	"net/http"
)

// CheckHTTPResponse does some basic error handling
// and reads the response body into a byte array
func CheckHTTPResponse(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if (resp.StatusCode / 100) != 2 {
		return nil, fmt.Errorf("[STATUS CODE - %d]\t%s", resp.StatusCode, body)
	}
	return body, nil
}
