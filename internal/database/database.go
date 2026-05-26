// Package database provides SQLite operations for reading and writing the
// Antigravity sidebar conversation index. It uses pure-Go SQLite via modernc.org/sqlite.
package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pb "gravity-anchor/internal/protobuf"

	_ "modernc.org/sqlite"
)

const trajectorySummariesKey = "antigravityUnifiedStateSync.trajectorySummaries"

// ExtractMetadata reads conversation metadata from a state.vscdb database.
// It returns:
//   - titles: map of conversation ID → title (only non-fallback titles)
//   - innerBlobs: map of conversation ID → raw inner protobuf bytes
func ExtractMetadata(dbPath string) (map[string]string, map[string][]byte, error) {
	titles := make(map[string]string)
	innerBlobs := make(map[string][]byte)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return titles, innerBlobs, fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	var value sql.NullString
	err = db.QueryRow(
		"SELECT value FROM ItemTable WHERE key=?", trajectorySummariesKey,
	).Scan(&value)
	if err != nil || !value.Valid || value.String == "" {
		return titles, innerBlobs, nil // No data yet, not an error
	}

	decoded, err := base64.StdEncoding.DecodeString(value.String)
	if err != nil {
		return titles, innerBlobs, fmt.Errorf("base64 decode: %w", err)
	}

	// Parse outer protobuf: repeated field 1 (length-delimited entries)
	pos := 0
	for pos < len(decoded) {
		tag, newPos, err := pb.DecodeVarint(decoded, pos)
		if err != nil {
			break
		}
		wireType := int(tag & 0x07)
		if wireType != pb.WireLengthDelimited {
			break
		}

		length, newPos, err := pb.DecodeVarint(decoded, newPos)
		if err != nil {
			break
		}
		entryEnd := newPos + int(length)
		if entryEnd > len(decoded) {
			break
		}
		entry := decoded[newPos:entryEnd]
		pos = entryEnd

		// Parse entry: field 1 (UUID string), field 2 (info sub-message)
		uid, infoB64 := parseEntry(entry)
		if uid == "" || infoB64 == "" {
			continue
		}

		rawInner, err := base64.StdEncoding.DecodeString(infoB64)
		if err != nil {
			continue
		}
		innerBlobs[uid] = rawInner

		// Extract title from inner blob: field 1 (string)
		title := extractTitleFromInner(rawInner)
		if title != "" &&
			!strings.HasPrefix(title, "Conversation (") &&
			!strings.HasPrefix(title, "Conversation ") {
			titles[uid] = title
		}
	}

	return titles, innerBlobs, nil
}

// parseEntry extracts the conversation UUID and base64-encoded info from a protobuf entry.
func parseEntry(entry []byte) (uid, infoB64 string) {
	pos := 0
	for pos < len(entry) {
		tag, newPos, err := pb.DecodeVarint(entry, pos)
		if err != nil {
			break
		}
		fieldNum := int(tag >> 3)
		wireType := int(tag & 0x07)

		if wireType == pb.WireLengthDelimited {
			length, np, err := pb.DecodeVarint(entry, newPos)
			if err != nil {
				break
			}
			content := entry[np : np+int(length)]
			newPos = np + int(length)

			if fieldNum == 1 {
				uid = string(content)
			} else if fieldNum == 2 {
				// Field 2 is a sub-message; extract field 1 (string) from it
				infoB64 = extractSubField1(content)
			}
		} else if wireType == pb.WireVarint {
			_, np, err := pb.DecodeVarint(entry, newPos)
			if err != nil {
				break
			}
			newPos = np
		} else {
			break
		}
		pos = newPos
	}
	return
}

// extractSubField1 extracts field 1 (string) from a protobuf sub-message.
func extractSubField1(data []byte) string {
	pos := 0
	if pos >= len(data) {
		return ""
	}
	// Read tag
	_, newPos, err := pb.DecodeVarint(data, pos)
	if err != nil {
		return ""
	}
	// Read length
	length, newPos, err := pb.DecodeVarint(data, newPos)
	if err != nil {
		return ""
	}
	endPos := newPos + int(length)
	if endPos > len(data) {
		return ""
	}
	return string(data[newPos:endPos])
}

// extractTitleFromInner extracts the title (field 1, string) from the inner protobuf blob.
func extractTitleFromInner(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	pos := 0
	// Read tag
	tag, newPos, err := pb.DecodeVarint(data, pos)
	if err != nil {
		return ""
	}
	fieldNum := int(tag >> 3)
	wireType := int(tag & 0x07)
	if fieldNum != 1 || wireType != pb.WireLengthDelimited {
		return ""
	}
	// Read length
	length, newPos, err := pb.DecodeVarint(data, newPos)
	if err != nil {
		return ""
	}
	endPos := newPos + int(length)
	if endPos > len(data) {
		return ""
	}
	return string(data[newPos:endPos])
}

// ExtractMetadataFromPaths reads metadata from ALL existing Antigravity databases.
// First DB wins for each conversation ID.
func ExtractMetadataFromPaths(dbPaths []string) (map[string]string, map[string][]byte) {
	mergedTitles := make(map[string]string)
	mergedBlobs := make(map[string][]byte)

	for _, dbPath := range dbPaths {
		titles, blobs, err := ExtractMetadata(dbPath)
		if err != nil {
			continue
		}
		for cid, title := range titles {
			if _, exists := mergedTitles[cid]; !exists {
				mergedTitles[cid] = title
			}
		}
		for cid, blob := range blobs {
			if _, exists := mergedBlobs[cid]; !exists {
				mergedBlobs[cid] = blob
			}
		}
	}

	return mergedTitles, mergedBlobs
}

// WriteIndex backs up the existing index and writes the rebuilt one to the database.
// Returns the backup filename (empty if no backup was needed) and any error.
func WriteIndex(dbPath string, encodedValue string, backupDir string, backupSuffix string) (string, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", fmt.Errorf("opening database: %w", err)
	}
	defer db.Close()

	// Read existing value for backup
	var oldValue sql.NullString
	err = db.QueryRow(
		"SELECT value FROM ItemTable WHERE key=?", trajectorySummariesKey,
	).Scan(&oldValue)

	hasExisting := err == nil && oldValue.Valid && oldValue.String != ""

	// Create backup if there was existing data
	var backupName string
	if hasExisting {
		if backupSuffix != "" {
			backupName = fmt.Sprintf("trajectorySummaries_backup_%s.txt", backupSuffix)
		} else {
			backupName = "trajectorySummaries_backup.txt"
		}
		backupPath := filepath.Join(backupDir, backupName)
		if err := os.WriteFile(backupPath, []byte(oldValue.String), 0644); err != nil {
			return "", fmt.Errorf("writing backup: %w", err)
		}
	}

	// Write the new value
	if hasExisting {
		_, err = db.Exec(
			"UPDATE ItemTable SET value=? WHERE key=?",
			encodedValue, trajectorySummariesKey,
		)
	} else {
		_, err = db.Exec(
			"INSERT INTO ItemTable (key, value) VALUES (?, ?)",
			trajectorySummariesKey, encodedValue,
		)
	}
	if err != nil {
		return backupName, fmt.Errorf("writing index: %w", err)
	}

	return backupName, nil
}
