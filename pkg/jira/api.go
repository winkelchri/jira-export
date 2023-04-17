package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	rutil "jira-export/pkg/request_util"
	"jira-export/pkg/secrets"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	config = &rutil.CacheConfig{
		OutputDir: "cache",
		Debug:     true,
	}
)

// HandleJSONDecodeError handles the error when decoding the JSON response
// and stores the response body in a file for debugging
func HandleJSONDecodeError(decodeErr error, resp *http.Response) error {
	// Store the response body in a file
	file, err := os.Create("error.txt")
	if err != nil {
		return fmt.Errorf("error creating error file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	_, err = file.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("error writing to error file: %v", err)
	}
	return fmt.Errorf("error decoding JSON: %v. Response body stored in error.txt", decodeErr)
}

func NewJiraAPI(secrets secrets.Secrets, maxResults int) JiraAPI {
	return JiraAPI{secrets: secrets, MaxResults: maxResults}
}

func makeRequest(url string, secrets secrets.Secrets) (*http.Request, error) {
	// Make HTTP request to Jira API to get filter data
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %v", err)
	}
	req.SetBasicAuth(secrets.Username, secrets.Token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

type JiraAPI struct {
	secrets    secrets.Secrets
	MaxResults int
}

// GetFilterResult returns the Jira Issues for a given filter
func (j JiraAPI) GetFilterResults(jql string) (results JiraSearchResults, err error) {
	// Build the search URL
	url := fmt.Sprintf("%s/rest/api/3/search", j.secrets.URL)

	// Prepare the cache directory
	if err := config.PrepareCacheDir(); err != nil {
		return results, fmt.Errorf("error preparing cache directory: %v", err)
	}

	// Build the request object
	req, err := buildSearchRequest(url, j.secrets, jql, j.MaxResults)
	if err != nil {
		return results, fmt.Errorf("error building search request: %v", err)
	}

	// Send the request with incremental backoff using the CachedRequest function
	resp, err := sendRequestWithBackoff(req, config)
	if err != nil {
		return results, fmt.Errorf("error sending search request: %v", err)
	}
	defer resp.Body.Close()

	// Decode the response body
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return results, HandleJSONDecodeError(err, resp)
	}

	// Fetch additional pages of results if necessary
	if results.Total > results.MaxResults {
		additionalData, err := j.fetchAdditionalResults(req, config, results.MaxResults, results.Total)
		if err != nil {
			return results, fmt.Errorf("error fetching additional results: %v", err)
		}
		results.Issues = append(results.Issues, additionalData...)
	}

	return results, nil
}

// sendRequestWithBackoff sends an HTTP request with incremental backoff using the CachedRequest function
func sendRequestWithBackoff(req *http.Request, config *rutil.CacheConfig) (*http.Response, error) {
	backoff := time.Second
	cr := rutil.CachedRequest{req}
	for {
		resp, err := cr.Cache(config)
		if err != nil {
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Minute {
				backoff = time.Minute
			}
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			time.Sleep(backoff)
			backoff *= 2
			if backoff > time.Minute {
				backoff = time.Minute
			}
			continue
		}
		backoff = time.Second
		return resp, nil
	}
}

func responseIsTooManyRequests(reader *bytes.Reader) bool {
	// Read the first 100 bytes of the response body
	data, err := io.ReadAll(io.LimitReader(reader, 100))
	if err != nil {
		return false
	}

	// Check if the response body starts with "<!DOCTYPE html>"
	// if true, then the response is an HTML page and not JSON
	// and the request was rate limited
	return string(data) == "<!DOCTYPE html>"
}

// fetchAdditionalResults fetches additional pages of Jira search results
func (j JiraAPI) fetchAdditionalResults(req *http.Request, config *rutil.CacheConfig, startAt int, total int) ([]interface{}, error) {
	additionalData := []interface{}{}

	// Build the search queries
	rs := buildSearchRequests(req, startAt, total)

	results := make(chan []interface{})
	defer close(results)

	// Send the requests in batches of 10
	for i := 0; i < len(rs); i += 10 {
		end := i + 10
		if end > len(rs) {
			end = len(rs)
		}

		// Send the requests in parallel
		for _, r := range rs[i:end] {
			go func(r *http.Request) {

				cr := rutil.CachedRequest{r}

				backoff := 1.0

				// Send the request with incremental backoff
				for {
					resp, err := cr.Cache(config)
					if err != nil {
						fmt.Println("Error sending request:", err)
						results <- nil
						return
					}

					respBodyBytes, err := io.ReadAll(resp.Body)
					if err != nil {
						fmt.Println("Error reading response body:", err)
						results <- nil
						return
					}

					// Check if the response is a rate limit error
					r1 := bytes.NewReader(respBodyBytes)
					r2 := bytes.NewReader(respBodyBytes)

					if responseIsTooManyRequests(r1) {
						sleepTime := time.Duration(float64(rand.Intn(500)+500) * backoff)
						time.Sleep(sleepTime * time.Millisecond)
						backoff *= 2
						continue
					}

					var data JiraSearchResults

					if err := json.NewDecoder(r2).Decode(&data); err != nil {
						fmt.Println("Error decoding JSON:", err)
						results <- nil
						return
					}

					defer resp.Body.Close()

					results <- data.Issues
					break
				}
			}(r)
		}

		for j := 0; j < end-i; j++ {
			r := <-results
			if r == nil {
				return nil, fmt.Errorf("parallel send failed with error")
			}
			additionalData = append(additionalData, r...)
		}
	}
	return additionalData, nil
}

// buildSearchRequest builds a GET request object for a Jira search query
func buildSearchRequest(url string, secrets secrets.Secrets, jql string, maxResults int) (*http.Request, error) {
	req, err := makeRequest(url, secrets)
	if err != nil {
		return nil, fmt.Errorf("error preparing GET request: %v", err)
	}

	// // Remove tailing and leading single quotes
	jql = strings.Trim(jql, "'")
	// // Remove tailing and leading double quotes
	jql = strings.Trim(jql, "\"")
	fmt.Println("JQL:", jql)

	q := req.URL.Query()
	q.Set("jql", jql)
	q.Set("maxResults", strconv.Itoa(maxResults))
	req.URL.RawQuery = q.Encode()

	return req, nil
}

// buildSearchRequests builds a slice of search query strings for fetching additional pages of Jira search results
func buildSearchRequests(req *http.Request, startAt int, total int) (r []*http.Request) {
	// maxResults := total - startAt
	maxResults := 100
	for i := startAt; i < total; i += maxResults {
		q := req.URL.Query()
		q.Set("startAt", strconv.Itoa(i))
		q.Set("maxResults", strconv.Itoa(maxResults))
		req.URL.RawQuery = q.Encode()

		r = append(r, req.Clone(context.Background()))
	}
	return r
}
