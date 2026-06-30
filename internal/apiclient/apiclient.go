// Package apiclient is a thin outbound client to the SaltyBytes API's admin
// endpoints. It lets the dashboard drive the live AI-model switch + registry
// without ever exposing the admin token to the browser — the dashboard backend
// holds the token and proxies operator actions.
package apiclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client talks to the API admin API. A zero/unconfigured client reports
// Enabled() == false and is never used to make a request.
type Client struct {
	baseURL    string
	idHeader   string
	adminToken string
	http       *http.Client
}

// New builds a client. baseURL/idHeader/adminToken come from the dashboard's
// optional API_BASE_URL / API_ID_HEADER / ADMIN_TOKEN env vars.
func New(baseURL, idHeader, adminToken string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		idHeader:   idHeader,
		adminToken: adminToken,
		// Activation runs a live validation probe on the API (a real extraction),
		// so allow generous time.
		http: &http.Client{Timeout: 90 * time.Second},
	}
}

// Enabled reports whether live model management is configured (base URL + token).
func (c *Client) Enabled() bool {
	return c != nil && c.baseURL != "" && c.adminToken != ""
}

// do issues a request to an admin path and returns the API's status + body
// verbatim, so the dashboard can relay validation errors (400 + {"error":...})
// straight to the operator.
func (c *Client) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.idHeader != "" {
		req.Header.Set("X-SaltyBytes-Identifier", c.idHeader)
	}
	req.Header.Set("X-Admin-Token", c.adminToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, respBody, nil
}

// CreateModel POSTs a new registry entry (the API validation-probes it).
func (c *Client) CreateModel(ctx context.Context, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/v1/admin/ai/models", body)
}

// UpdateModel PUTs edits to an existing entry.
func (c *Client) UpdateModel(ctx context.Context, id uint, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPut, fmt.Sprintf("/v1/admin/ai/models/%d", id), body)
}

// DeleteModel removes an entry.
func (c *Client) DeleteModel(ctx context.Context, id uint) (int, []byte, error) {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/admin/ai/models/%d", id), nil)
}

// SetActive switches the live light tier (the API probes first, fail-closed).
func (c *Client) SetActive(ctx context.Context, body []byte) (int, []byte, error) {
	return c.do(ctx, http.MethodPut, "/v1/admin/ai/active", body)
}
