package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ReleaseClient fetches release metadata for a GitHub repo. It's an
// interface so callers (internal/neovim) can be unit tested against a fake
// instead of hitting the real API.
type ReleaseClient interface {
	LatestRelease(ctx context.Context, owner, repo string) (Release, error)
}

// HTTPReleaseClient is the real ReleaseClient, backed by net/http against
// the public GitHub REST API.
type HTTPReleaseClient struct {
	HTTPClient *http.Client
	// BaseURL defaults to https://api.github.com; overridable in tests to
	// point at an httptest.Server.
	BaseURL string
}

// NewHTTPReleaseClient returns an HTTPReleaseClient with sane defaults.
func NewHTTPReleaseClient() *HTTPReleaseClient {
	return &HTTPReleaseClient{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    "https://api.github.com",
	}
}

func (c *HTTPReleaseClient) LatestRelease(ctx context.Context, owner, repo string) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.BaseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("GET %s: unexpected status %s", url, resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, fmt.Errorf("decoding response from %s: %w", url, err)
	}
	return release, nil
}
