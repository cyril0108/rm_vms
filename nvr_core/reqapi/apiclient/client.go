package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a generic HTTP client wrapper (like an Axios instance)
type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// NewClient creates a new API Client pointing to a specific root URL
func NewClient(baseURL string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}

	return &Client{
		BaseURL: parsedURL,
		// ALWAYS set a timeout in Go to prevent goroutine leaks if the external API hangs
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Do handles the actual HTTP request, JSON marshaling, and unmarshaling
func (c *Client) Do(ctx context.Context, method, path string, bodyPayload any, responseTarget any) error {
	// Resolve the path against the BaseURL
	rel, err := url.Parse(path)
	if err != nil {
		return err
	}
	u := c.BaseURL.ResolveReference(rel)

	// Marshal body if provided
	var reqBody io.Reader
	if bodyPayload != nil {
		jsonData, err := json.Marshal(bodyPayload)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// Create the request with Context (allows cancellation/timeouts)
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute Request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle HTTP Errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Unmarshal response if a target struct was provided
	if responseTarget != nil {
		if err := json.NewDecoder(resp.Body).Decode(responseTarget); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}