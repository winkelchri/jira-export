package request_util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"jira-export/pkg/output"
	t "jira-export/pkg/terminal"
	"net/http"
	"os"
)

var (
	DEFAULT_CACHE_DIR = "cache"
)

type CachedRequest struct {
	*http.Request
}

type CacheConfig struct {
	OutputDir string
	Debug     bool
}

// PrepareCacheDir creates the cache directory if it does not exist
func (config *CacheConfig) PrepareCacheDir() error {
	if _, err := os.Stat(config.OutputDir); os.IsNotExist(err) {
		if err := os.Mkdir(config.OutputDir, 0755); err != nil {
			return fmt.Errorf("error creating cache directory: %v", err)
		}
	}
	return nil
}

// Cache sends the request and stores the response body in a file.
func (req *CachedRequest) Cache(config ...*CacheConfig) (*http.Response, error) {
	// Create the cache filename from the request URL and query parameters
	// and encode it as sha256
	hash := sha1.New()
	hash.Write([]byte(req.URL.String()))
	cacheID := hex.EncodeToString(hash.Sum(nil))

	// Use the default cache output directory if the config object is not provided
	outputDir := DEFAULT_CACHE_DIR
	if len(config) > 0 && config[0].OutputDir != "" {
		outputDir = config[0].OutputDir
	}

	debug := false
	if len(config) > 0 {
		debug = config[0].Debug
	}

	cacheFile := fmt.Sprintf("%s/%x.json", outputDir, cacheID)

	// Check if there is a cache file and load body from it
	// otherwise send the request
	if _, err := os.Stat(cacheFile); err == nil {
		if debug {
			fmt.Println(t.Cyan+"loading"+t.Reset, "response body from cache file:", t.Underline+cacheFile+t.Reset)
		}

		file, err := os.Open(cacheFile)
		if err != nil {
			return nil, fmt.Errorf("error opening file: %v", err)
		}
		// defer file.Close()

		resp := &http.Response{
			StatusCode: 200,
			Body:       file,
		}
		return resp, nil
	}

	// Send the request
	client := &http.Client{}
	// Debug the request
	if debug {
		fmt.Println(t.Green+"sending"+t.Reset, "request:", t.Underline+req.URL.String()+t.Reset)
	}

	resp, err := client.Do(req.Request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	// Store the response body in a file
	if debug {
		fmt.Println(t.Purple+"storing"+t.Reset, "response body into cache file:", t.Underline+cacheFile+t.Reset)
	}
	err = output.StoreJSON(resp.Body, cacheFile)
	if err != nil {
		return nil, fmt.Errorf("error storing response body: %v", err)
	}

	// Reopen the file and set it as the response body
	file, err := os.Open(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	// defer file.Close()
	resp.Body = file

	return resp, nil
}
