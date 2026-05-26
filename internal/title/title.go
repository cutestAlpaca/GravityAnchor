// Package title provides conversation title resolution with a three-level
// priority system: database-preserved titles, brain artifact headings, and
// date-based fallback titles.
package title

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Resolve determines the best title for a conversation using a three-level priority:
//  1. Existing title from database (source="preserved")
//  2. Brain artifact .md heading (source="brain")
//  3. Fallback: date + short UUID (source="fallback")
func Resolve(cid string, existingTitles map[string]string, brainDirs []string, pbPath string) (string, string) {
	// Priority 1: Canonical title from database
	if t, ok := existingTitles[cid]; ok {
		return t, "preserved"
	}

	// Priority 2: Brain artifact heading
	if brainTitle := GetFromBrain(cid, brainDirs); brainTitle != "" {
		return brainTitle, "brain"
	}

	// Priority 3: Fallback with date
	if pbPath != "" {
		if info, err := os.Stat(pbPath); err == nil {
			modTime := info.ModTime().Format("Jan 02")
			return fmt.Sprintf("Conversation (%s) %s", modTime, cid[:min(8, len(cid))]), "fallback"
		}
	}

	return fmt.Sprintf("Conversation %s", cid[:min(8, len(cid))]), "fallback"
}

// GetFromBrain extracts a title from brain artifact .md files.
// It looks for the first markdown heading (line starting with #) in
// alphabetically-sorted .md files within the conversation's brain folder.
func GetFromBrain(cid string, brainDirs []string) string {
	// Find brain path across all dirs
	var brainPath string
	for _, dir := range brainDirs {
		p := filepath.Join(dir, cid)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			brainPath = p
			break
		}
	}
	if brainPath == "" {
		return ""
	}

	entries, err := os.ReadDir(brainPath)
	if err != nil {
		return ""
	}

	// Sort entries alphabetically
	names := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".md") {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		filePath := filepath.Join(brainPath, name)
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") {
				title := strings.TrimLeft(line, "# ")
				title = strings.TrimSpace(title)
				if len(title) > 80 {
					title = title[:80]
				}
				f.Close()
				return title
			}
		}
		f.Close()
	}

	return ""
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatModTime formats a time for display in fallback titles.
func FormatModTime(t time.Time) string {
	return t.Format("Jan 02")
}
