// Package scanner provides conversation file scanning and merging across
// multiple Antigravity data directories. It supports both legacy .pb (protobuf)
// and newer .db (SQLite) conversation formats, with priority-based deduplication.
package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ConversationFile represents a single conversation file found on disk.
type ConversationFile struct {
	ID       string    `json:"id"`
	FilePath string    `json:"filePath"`
	ModTime  time.Time `json:"modTime"`
	Format   string    `json:"format"` // "pb" or "db"
}

// CollectConversations scans all given directories for conversation files.
// It accepts .pb (legacy protobuf) and .db (new SQLite) files, skipping
// SQLite journal files (.db-shm, .db-wal).
// First directory wins for deduplication (priority order).
func CollectConversations(dirs []string) map[string]*ConversationFile {
	catalog := make(map[string]*ConversationFile)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()

			var cid, format string
			switch {
			case strings.HasSuffix(name, ".pb"):
				cid = strings.TrimSuffix(name, ".pb")
				format = "pb"
			case strings.HasSuffix(name, ".db") &&
				!strings.HasSuffix(name, ".db-shm") &&
				!strings.HasSuffix(name, ".db-wal"):
				cid = strings.TrimSuffix(name, ".db")
				format = "db"
			default:
				continue
			}

			// First seen wins (directories are in priority order)
			if _, exists := catalog[cid]; exists {
				continue
			}

			fullPath := filepath.Join(dir, name)
			info, err := entry.Info()
			if err != nil {
				continue
			}

			catalog[cid] = &ConversationFile{
				ID:       cid,
				FilePath: fullPath,
				ModTime:  info.ModTime(),
				Format:   format,
			}
		}
	}

	return catalog
}

// SortByModTime returns conversation IDs sorted by modification time (newest first).
func SortByModTime(catalog map[string]*ConversationFile) []string {
	ids := make([]string, 0, len(catalog))
	for id := range catalog {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return catalog[ids[i]].ModTime.After(catalog[ids[j]].ModTime)
	})

	return ids
}

// GetDirSummary returns a count of conversations per parent directory.
func GetDirSummary(catalog map[string]*ConversationFile) map[string]int {
	counts := make(map[string]int)
	for _, cf := range catalog {
		parent := filepath.Dir(cf.FilePath)
		// Use the grandparent folder name for a cleaner label (e.g., "antigravity-ide")
		grandparent := filepath.Base(filepath.Dir(parent))
		counts[grandparent]++
	}
	return counts
}
