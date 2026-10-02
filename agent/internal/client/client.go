// Package client talks to the control-plane agent API over HTTP.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"agent/internal/protocol"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNoJob        = errors.New("no job")
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Heartbeat(ctx context.Context, req protocol.HeartbeatRequest) (protocol.HeartbeatResponse, error) {
	var out protocol.HeartbeatResponse
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/agent/heartbeat", req, &out)
	return out, err
}

// NextJob returns (nil, nil) when the control plane has nothing queued.
func (c *Client) NextJob(ctx context.Context) (*protocol.Job, error) {
	var out protocol.Job
	err := c.doJSON(ctx, http.MethodGet, "/api/v1/agent/jobs/next", nil, &out)
	if errors.Is(err, ErrNoJob) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, nil
	}
	return &out, nil
}

func (c *Client) ReportJob(ctx context.Context, jobID string, result protocol.JobResultRequest) error {
	path := "/api/v1/agent/jobs/" + jobID + "/result"
	return c.doJSON(ctx, http.MethodPost, path, result, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Agent-Protocol", fmt.Sprintf("%d", protocol.Version))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	switch {
	case resp.StatusCode == http.StatusNoContent:
		return ErrNoJob
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return ErrUnauthorized
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return fmt.Errorf("%s %s: %s: %s", method, path, resp.Status, truncate(respBody, 200))
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
