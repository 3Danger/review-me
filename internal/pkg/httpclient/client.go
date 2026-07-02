package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"review-info/internal/domain"
)

func classifyHTTPError(code int, body []byte) error {
	switch {
	case code == http.StatusUnauthorized:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrUnauthorized, code, string(body))
	case code == http.StatusForbidden:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrForbidden, code, string(body))
	case code == http.StatusNotFound:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrNotFound, code, string(body))
	case code == http.StatusBadRequest:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrBadRequest, code, string(body))
	case code >= 500:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrServerError, code, string(body))
	default:
		return fmt.Errorf("%w: status=%d, body=%s", domain.ErrUnknown, code, string(body))
	}
}

// Client wraps a domain.HTTPClient with a base URL and provides
// convenience methods that eliminate HTTP lifecycle boilerplate.
type Client struct {
	raw     domain.HTTPClient
	baseURL string
}

// New returns a Client that prepends baseURL to every request path.
func New(client domain.HTTPClient, baseURL string) *Client {
	return &Client{
		raw:     client,
		baseURL: baseURL,
	}
}

// Get performs an HTTP GET, checks status 200, and JSON-decodes into target.
// On non-200 status, returns a descriptive error with the response body.
func (c *Client) Get(ctx context.Context, path string, headers map[string]string, target interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.raw.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrNetwork, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return classifyHTTPError(resp.StatusCode, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding json: %w", err)
	}
	return nil
}
