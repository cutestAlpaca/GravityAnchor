// Package platform provides multi-platform path detection for Antigravity IDE data.
// It auto-detects the operating system and locates state databases, conversation
// directories, and brain directories across macOS, Linux, and Windows (including WSL).
package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Antigravity application name variants to search for.
var antigravityNames = []string{"Antigravity IDE", "antigravity", "Antigravity"}

// Conversation subfolder variants within ~/.gemini/.
var conversationVariants = []string{"antigravity-ide", "antigravity", "antigravity-backup"}

// PathInfo holds all detected paths for the current platform.
type PathInfo struct {
	OS                  string   // "darwin", "linux", or "windows"
	IsWSL               bool     // True if running under Windows Subsystem for Linux
	DBPaths             []string // All existing state.vscdb paths
	ConversationDirs    []string // All existing conversation directories
	BrainDirs           []string // All existing brain directories
	WorkspaceStorageDir string   // First existing workspaceStorage directory
}

// DetectPaths auto-detects all relevant paths based on the current OS.
// It scans for multiple Antigravity name variants and conversation subfolder
// variants, returning only paths that actually exist on disk.
func DetectPaths() *PathInfo {
	info := &PathInfo{
		OS:    runtime.GOOS,
		IsWSL: detectWSL(),
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return info
	}

	// Determine the application data root directory based on OS.
	var appDataRoots []string
	switch runtime.GOOS {
	case "darwin":
		appDataRoots = []string{filepath.Join(homeDir, "Library", "Application Support")}
	case "linux":
		appDataRoots = []string{filepath.Join(homeDir, ".config")}
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			appDataRoots = []string{appData}
		}
	}

	// Collect DB paths: <appdata>/<AntigravityName>/User/globalStorage/state.vscdb
	var dbCandidates []string
	for _, root := range appDataRoots {
		for _, name := range antigravityNames {
			dbCandidates = append(dbCandidates,
				filepath.Join(root, name, "User", "globalStorage", "state.vscdb"))
		}
	}
	info.DBPaths = existingPaths(dbCandidates...)

	// Collect workspace storage dir: <appdata>/<AntigravityName>/User/workspaceStorage
	var wsCandidates []string
	for _, root := range appDataRoots {
		for _, name := range antigravityNames {
			wsCandidates = append(wsCandidates,
				filepath.Join(root, name, "User", "workspaceStorage"))
		}
	}
	info.WorkspaceStorageDir = firstExisting(wsCandidates...)

	// Conversation and brain dirs under ~/.gemini/<variant>/
	geminiBase := filepath.Join(homeDir, ".gemini")

	var convCandidates []string
	var brainCandidates []string
	for _, variant := range conversationVariants {
		convCandidates = append(convCandidates,
			filepath.Join(geminiBase, variant, "conversations"))
		brainCandidates = append(brainCandidates,
			filepath.Join(geminiBase, variant, "brain"))
	}
	info.ConversationDirs = existingPaths(convCandidates...)
	info.BrainDirs = existingPaths(brainCandidates...)

	return info
}

// detectWSL checks if the current Linux environment is Windows Subsystem for Linux.
// It looks for WSL-specific indicators in /proc/version.
func detectWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// firstExisting returns the first path from the candidates that exists on disk.
// Returns an empty string if none exist.
func firstExisting(candidates ...string) string {
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// existingPaths returns all paths from the candidates that exist on disk.
// The order of candidates is preserved.
func existingPaths(candidates ...string) []string {
	var result []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			result = append(result, p)
		}
	}
	return result
}
