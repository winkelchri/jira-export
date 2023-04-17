package jira

// Testing basic functionality like making a request and decoding the response
// is not necessary because the Go standard library already has tests for
// these functions. Testing the error handling is more important because
// this is where the program will fail if the Jira API changes.

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockHTTPClient is a mock HTTP client that returns a canned response
type MockHTTPClient struct {
	// Response is the canned response that the client will return
	Response *http.Response
}

// Do is the mock implementation of the HTTP client's Do method
func (c *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Response, nil
}

// TestHandleJSONDecodeError tests the HandleJSONDecodeError function
func TestHandleJSONDecodeError(t *testing.T) {
	// Create a temporary file to store the response body
	file, err := ioutil.TempFile("", "jira-export-")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	// Create a mock response with a body
	resp := &http.Response{
		Body: ioutil.NopCloser(file),
	}

	// Call HandleJSONDecodeError
	err = HandleJSONDecodeError(fmt.Errorf("test error"), resp)
	assert.Error(t, err)

	// remove the error.txt file
	err = os.Remove("error.txt")
	assert.NoError(t, err)
}
