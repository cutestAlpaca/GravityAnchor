// Package updater checks for newer versions of the application on GitHub Releases.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CurrentVersion is the current version of this application.
const CurrentVersion = "1.0.0"

// GitHubRepo is the GitHub repository for update checks.
const GitHubRepo = "FutureisinPast/gravity-anchor"

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	URL            string `json:"url"`
}

// CheckForUpdates queries the GitHub Releases API to check for a newer version.
// Returns nil if no update is available or if the check fails (e.g., no internet).
func CheckForUpdates() *UpdateInfo {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "GravityAnchor")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	tag := strings.TrimLeft(release.TagName, "Vv")
	if tag == "" {
		return nil
	}

	if !isNewer(tag, CurrentVersion) {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: CurrentVersion,
			LatestVersion:  tag,
			URL:            release.HTMLURL,
		}
	}

	return &UpdateInfo{
		Available:      true,
		CurrentVersion: CurrentVersion,
		LatestVersion:  tag,
		URL:            release.HTMLURL,
	}
}

// isNewer returns true if remote version is newer than local version.
func isNewer(remote, local string) bool {
	remoteParts := parseVersion(remote)
	localParts := parseVersion(local)

	for i := 0; i < len(remoteParts) || i < len(localParts); i++ {
		r, l := 0, 0
		if i < len(remoteParts) {
			r = remoteParts[i]
		}
		if i < len(localParts) {
			l = localParts[i]
		}
		if r > l {
			return true
		}
		if r < l {
			return false
		}
	}
	return false
}

// parseVersion splits a version string into integer parts.
func parseVersion(version string) []int {
	parts := strings.Split(version, ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			result[i] = 0
		} else {
			result[i] = n
		}
	}
	return result
}
