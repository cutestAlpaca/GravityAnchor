// Package workspace handles workspace URI construction, extraction, and inference
// for Antigravity conversation entries. It converts filesystem paths to file:/// URIs,
// builds protobuf workspace fields, and infers workspace associations from brain artifacts.
package workspace

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	pb "gravity-anchor/internal/protobuf"
)

// PathToURI converts a local folder path to a file:/// URI matching Antigravity's format.
// Passes through remote URIs (vscode-remote://, file:///) unchanged.
func PathToURI(folderPath string, system string, isWSL bool) string {
	// Pass through URIs already in correct format
	if IsRemoteURI(folderPath) {
		return folderPath
	}

	// WSL: convert /mnt/<drive>/... to file:///<drive>:/...
	if isWSL && strings.HasPrefix(folderPath, "/mnt/") {
		parts := strings.SplitN(folderPath, "/", 4)
		if len(parts) >= 3 && len(parts[2]) == 1 {
			drive := strings.ToLower(parts[2])
			rest := ""
			if len(parts) == 4 {
				rest = "/" + parts[3]
			}
			return "file:///" + drive + ":" + rest
		}
	}

	// Normalize backslashes
	p := strings.ReplaceAll(folderPath, "\\", "/")

	// Windows drive letter
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		rest := p[2:]
		return "file:///" + drive + ":" + rest
	}

	// Unix path
	return "file:///" + strings.TrimLeft(p, "/")
}

// IsRemoteURI checks if a string is a remote/absolute URI (not a local path).
func IsRemoteURI(pathOrURI string) bool {
	return strings.HasPrefix(pathOrURI, "vscode-remote://") ||
		strings.HasPrefix(pathOrURI, "file:///")
}

// BuildWorkspaceField builds protobuf field 9 (workspace sub-message) from a path.
// Sub-message structure: sub-field 1 (string) = URI, sub-field 2 (string) = URI.
func BuildWorkspaceField(folderPath string, system string, isWSL bool) []byte {
	uri := PathToURI(folderPath, system, isWSL)
	subMsg := append(
		pb.EncodeStringField(1, uri),
		pb.EncodeStringField(2, uri)...,
	)
	return pb.EncodeLengthDelimited(9, subMsg)
}

// ExtractWorkspaceHint tries to extract a workspace URI from a protobuf inner blob.
// It scans length-delimited fields for strings containing file:/// or vscode-remote://.
func ExtractWorkspaceHint(innerBlob []byte) string {
	if len(innerBlob) == 0 {
		return ""
	}

	pos := 0
	for pos < len(innerBlob) {
		tag, newPos, err := pb.DecodeVarint(innerBlob, pos)
		if err != nil {
			break
		}
		wireType := int(tag & 0x07)
		fieldNum := int(tag >> 3)

		switch wireType {
		case pb.WireLengthDelimited:
			length, np, err := pb.DecodeVarint(innerBlob, newPos)
			if err != nil {
				return ""
			}
			endPos := np + int(length)
			if endPos > len(innerBlob) {
				return ""
			}
			content := innerBlob[np:endPos]
			newPos = endPos

			// Look for workspace URIs in fields > 1
			if fieldNum > 1 {
				text := string(content)
				if strings.Contains(text, "file:///") || strings.Contains(text, "vscode-remote://") {
					return text
				}
			}
			pos = newPos

		case pb.WireVarint:
			_, np, err := pb.DecodeVarint(innerBlob, newPos)
			if err != nil {
				return ""
			}
			pos = np

		case pb.Wire64Bit:
			pos = newPos + 8

		case pb.Wire32Bit:
			pos = newPos + 4

		default:
			return ""
		}
	}

	return ""
}

// LoadKnownURIs loads all known workspace URIs from Antigravity's workspaceStorage.
// Each subfolder contains a workspace.json with a "folder" or "workspace" URI.
// Returns URIs sorted longest-first for prefix matching.
func LoadKnownURIs(storageDir string) []string {
	if storageDir == "" {
		return nil
	}

	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil
	}

	var uris []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		wsJSON := filepath.Join(storageDir, entry.Name(), "workspace.json")
		data, err := os.ReadFile(wsJSON)
		if err != nil {
			continue
		}

		var wsData struct {
			Folder    string `json:"folder"`
			Workspace string `json:"workspace"`
		}
		if err := json.Unmarshal(data, &wsData); err != nil {
			continue
		}

		uri := wsData.Folder
		if uri == "" {
			uri = wsData.Workspace
		}
		if uri != "" {
			uris = append(uris, uri)
		}
	}

	// Sort longest first for more-specific prefix matching
	sort.Slice(uris, func(i, j int) bool {
		return len(uris[i]) > len(uris[j])
	})

	return uris
}

// URIToLocalPath converts a file:/// URI to a local filesystem path.
// Returns empty string for non-file URIs.
func URIToLocalPath(fileURI string, system string, isWSL bool) string {
	if !strings.HasPrefix(fileURI, "file:///") {
		return ""
	}

	raw, err := url.PathUnescape(fileURI[len("file://"):])
	if err != nil {
		raw = fileURI[len("file://"):]
	}

	// Windows: file:///C:/... -> C:/...
	if system == "windows" && len(raw) >= 3 && raw[0] == '/' && raw[2] == ':' {
		raw = raw[1:]
	}

	// WSL: file:///C:/... -> /mnt/c/...
	if isWSL && len(raw) >= 3 && raw[0] == '/' && raw[2] == ':' {
		drive := strings.ToLower(string(raw[1]))
		raw = "/mnt/" + drive + raw[3:]
	}

	return raw
}

// InferFromBrain scans brain .md files for file:/// and vscode-remote:// paths
// and infers the workspace by matching against known workspace URIs.
func InferFromBrain(conversationID string, brainDirs []string, knownURIs []string, system string, isWSL bool) string {
	// Find brain path
	var brainPath string
	for _, dir := range brainDirs {
		p := filepath.Join(dir, conversationID)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			brainPath = p
			break
		}
	}
	if brainPath == "" {
		return ""
	}

	// Patterns for file URIs
	var localPattern *regexp.Regexp
	if system == "windows" {
		localPattern = regexp.MustCompile(`file:///([A-Za-z](?:%3A|:)/[^\s"'\]>)]+)`)
	} else {
		localPattern = regexp.MustCompile(`file:///([^\s"'\]>)]+)`)
	}
	remotePattern := regexp.MustCompile(`(vscode-remote://[^\s"'\]>)]+)`)

	var foundURIs []string
	var foundRemote []string

	entries, err := os.ReadDir(brainPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, ".") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(brainPath, name))
		if err != nil {
			continue
		}
		// Only read first 16KB
		content := string(data)
		if len(content) > 16384 {
			content = content[:16384]
		}

		for _, match := range remotePattern.FindAllStringSubmatch(content, -1) {
			foundRemote = append(foundRemote, match[1])
		}
		for _, match := range localPattern.FindAllStringSubmatch(content, -1) {
			foundURIs = append(foundURIs, "file:///"+match[1])
		}
	}

	if len(foundURIs) == 0 && len(foundRemote) == 0 {
		return ""
	}

	// Strategy 1: Match against known workspace URIs
	if len(knownURIs) > 0 {
		wsCounts := make(map[string]int)

		normalizeURI := func(uri string) string {
			uri = strings.ReplaceAll(uri, "%3A", ":")
			uri = strings.ReplaceAll(uri, "%3a", ":")
			uri = strings.ReplaceAll(uri, "%20", " ")
			return uri
		}

		for _, fileURI := range foundURIs {
			normalized := normalizeURI(fileURI)
			for _, wsURI := range knownURIs {
				wsNorm := normalizeURI(wsURI)
				if strings.HasPrefix(normalized, wsNorm+"/") || normalized == wsNorm {
					wsCounts[wsURI]++
					break
				}
			}
		}

		for _, remoteURI := range foundRemote {
			for _, wsURI := range knownURIs {
				if strings.HasPrefix(remoteURI, wsURI+"/") || remoteURI == wsURI {
					wsCounts[wsURI]++
					break
				}
			}
		}

		if len(wsCounts) > 0 {
			bestURI := maxByCount(wsCounts)
			local := URIToLocalPath(bestURI, system, isWSL)
			if local != "" {
				return local
			}
			return bestURI
		}
	}

	// Strategy 2: Heuristic depth-based approach
	pathCounts := make(map[string]int)

	for _, fileURI := range foundURIs {
		raw := fileURI[len("file:///"):]
		raw = strings.ReplaceAll(raw, "%3A", ":")
		raw = strings.ReplaceAll(raw, "%3a", ":")
		raw = strings.ReplaceAll(raw, "%20", " ")

		if isWSL && len(raw) >= 2 && raw[1] == ':' {
			drive := strings.ToLower(string(raw[0]))
			raw = "mnt/" + drive + "/" + raw[3:]
		}

		parts := strings.Split(strings.ReplaceAll(raw, "\\", "/"), "/")

		depth := 4
		if system == "windows" || (isWSL && strings.HasPrefix(raw, "mnt/")) {
			depth = 5
		}

		if len(parts) >= depth {
			ws := strings.Join(parts[:depth], "/")
			if system != "windows" && !strings.HasPrefix(ws, "/") {
				ws = "/" + ws
			}
			pathCounts[ws]++
		}
	}

	for _, remoteURI := range foundRemote {
		pathCounts[remoteURI]++
	}

	if len(pathCounts) == 0 {
		return ""
	}

	best := maxByCount(pathCounts)
	if strings.HasPrefix(best, "vscode-remote://") {
		return best
	}
	return strings.ReplaceAll(best, "/", string(filepath.Separator))
}

// maxByCount returns the key with the highest count in the map.
func maxByCount(m map[string]int) string {
	var bestKey string
	bestCount := -1
	for k, v := range m {
		if v > bestCount {
			bestKey = k
			bestCount = v
		}
	}
	return bestKey
}
