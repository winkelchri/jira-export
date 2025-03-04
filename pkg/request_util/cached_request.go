package request_util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"jira-export/pkg/logger"
	"jira-export/pkg/output"
	"net/http"
	"os"
)

var (
	DEFAULT_CACHE_DIR = "cache"
)

type CachedRequest struct {
	*http.Request
	hash string
}

type CacheConfig struct {
	OutputDir string
	Debug     bool
}

// NewCachedRequest creates a new CachedRequest object
func NewCachedRequest(req *http.Request) *CachedRequest {
	return &CachedRequest{Request: req}
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

// SendRequest sends the HTTP request and returns the response
func (req *CachedRequest) SendRequest() (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req.Request)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	return resp, nil
}

// GetCacheID returns the cache ID for the request
func (req *CachedRequest) GetCacheID() string {
	if req.hash == "" {
		// Create the cache filename from the request URL and query parameters
		// and encode it as sha256
		hash := sha1.New()
		hash.Write([]byte(req.URL.String()))
		req.hash = hex.EncodeToString(hash.Sum(nil))
	}

	return req.hash
}

// GetCacheFile returns the cache file path for the request
func (req *CachedRequest) GetCacheFile(config *CacheConfig) string {
	return fmt.Sprintf("%s/%x.json", config.OutputDir, req.GetCacheID())
}

// ClearCacheFile deletes the cache file
func (req *CachedRequest) ClearCacheFile(config *CacheConfig) error {
	logger.Logger.Info("Clearing cache file", "cacheFile", req.GetCacheFile(config))

	cacheFile := req.GetCacheFile(config)
	err := os.Remove(cacheFile)
	if err != nil {
		return fmt.Errorf("error clearing cache file: %v", err)
	}
	return nil
}

// Cache sends the request and stores the response body in a file.
func (req *CachedRequest) Cache(config ...*CacheConfig) (*http.Response, error) {
	// Use the default cache output directory if the config object is not provided
	outputDir := DEFAULT_CACHE_DIR
	if len(config) > 0 && config[0].OutputDir != "" {
		outputDir = config[0].OutputDir
	}

	debug := false
	if len(config) > 0 {
		debug = config[0].Debug
	}

	cacheFile := fmt.Sprintf("%s/%x.json", outputDir, req.GetCacheID())

	// Check if there is a cache file and load body from it
	// otherwise send the request
	if _, err := os.Stat(cacheFile); err == nil {
		if debug {
			logger.Logger.Info("Loading response body from cache file", "cacheFile", cacheFile)
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
	resp, err := req.SendRequest()
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	// Store the response body in a file
	if debug {
		logger.Logger.Info("Storing response body into cache file", "cacheFile", cacheFile)
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
