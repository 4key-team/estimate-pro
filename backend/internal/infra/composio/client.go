package composio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://backend.composio.dev/api/v3"

// Client communicates with the Composio REST API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// WithBaseURL sets a custom base URL (for testing).
func (c *Client) WithBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// ExecuteAction runs a Composio action with the given parameters.
func (c *Client) ExecuteAction(ctx context.Context, action, connectedAccountID string, params map[string]any) error {
	body := map[string]any{
		"connectedAccountId": connectedAccountID,
		"input":              params,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("composio.ExecuteAction: marshal: %w", err)
	}

	url := fmt.Sprintf("%s/actions/%s/execute", c.baseURL, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("composio.ExecuteAction: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("composio.ExecuteAction: do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("composio.ExecuteAction: status %d for action %s: %s", resp.StatusCode, action, string(respBody))
	}
	return nil
}
