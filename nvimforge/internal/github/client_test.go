package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLatestRelease_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/repos/neovim/neovim/releases/latest"
		if r.URL.Path != wantPath {
			t.Errorf("request path = %q, want %q", r.URL.Path, wantPath)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"tag_name": "v0.10.2",
			"name": "v0.10.2",
			"assets": [
				{"name": "nvim-macos-arm64.tar.gz", "browser_download_url": "https://example.com/nvim-macos-arm64.tar.gz", "size": 12345}
			]
		}`)
	}))
	defer srv.Close()

	c := &HTTPReleaseClient{HTTPClient: srv.Client(), BaseURL: srv.URL}
	release, err := c.LatestRelease(context.Background(), "neovim", "neovim")
	if err != nil {
		t.Fatalf("LatestRelease: %v", err)
	}

	if release.TagName != "v0.10.2" {
		t.Errorf("TagName = %q, want %q", release.TagName, "v0.10.2")
	}
	if len(release.Assets) != 1 {
		t.Fatalf("got %d assets, want 1", len(release.Assets))
	}
	if release.Assets[0].Name != "nvim-macos-arm64.tar.gz" {
		t.Errorf("asset name = %q", release.Assets[0].Name)
	}
}

func TestLatestRelease_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &HTTPReleaseClient{HTTPClient: srv.Client(), BaseURL: srv.URL}
	_, err := c.LatestRelease(context.Background(), "neovim", "neovim")
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
}

func TestLatestRelease_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	defer srv.Close()

	c := &HTTPReleaseClient{HTTPClient: srv.Client(), BaseURL: srv.URL}
	_, err := c.LatestRelease(context.Background(), "neovim", "neovim")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestNewHTTPReleaseClient_Defaults(t *testing.T) {
	c := NewHTTPReleaseClient()
	if c.BaseURL != "https://api.github.com" {
		t.Errorf("BaseURL = %q, want https://api.github.com", c.BaseURL)
	}
	if c.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}
