// Package github is a minimal client for the one GitHub REST API endpoint
// nvimforge needs (a repo's latest release). It deliberately avoids a full
// SDK dependency for a single endpoint.
package github

// Release is the subset of GitHub's release API response nvimforge needs.
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset is one downloadable file attached to a Release.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}
