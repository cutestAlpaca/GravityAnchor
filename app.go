package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gravity-anchor/internal/database"
	pb "gravity-anchor/internal/protobuf"
	"gravity-anchor/internal/platform"
	"gravity-anchor/internal/scanner"
	"gravity-anchor/internal/title"
	"gravity-anchor/internal/updater"
	"gravity-anchor/internal/workspace"
)

// App struct holds the application state and provides methods bound to the frontend.
type App struct {
	ctx              context.Context
	pathInfo         *platform.PathInfo
	catalog          map[string]*scanner.ConversationFile
	sortedIDs        []string
	resolved         []ResolvedConversation
	titles           map[string]string
	blobs            map[string][]byte
	knownURIs        []string
	manualWorkspaces map[string]string
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{
		manualWorkspaces: make(map[string]string),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ── Data Types ──────────────────────────────────────────────────────────────

// SystemInfo is returned by GetSystemInfo.
type SystemInfo struct {
	OS                  string   `json:"os"`
	DBPaths             []string `json:"dbPaths"`
	ConversationDirs    []string `json:"conversationDirs"`
	BrainDirs           []string `json:"brainDirs"`
	WorkspaceStorageDir string   `json:"workspaceStorageDir"`
}

// ResolvedConversation represents a single conversation after scanning.
type ResolvedConversation struct {
	Index        int    `json:"index"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	Source       string `json:"source"`       // "brain", "preserved", "fallback"
	HasWorkspace bool   `json:"hasWorkspace"` // true if workspace metadata exists
}

// ScanResult is returned by ScanConversations.
type ScanResult struct {
	Conversations []ResolvedConversation `json:"conversations"`
	Stats         map[string]int         `json:"stats"`         // brain, preserved, fallback counts
	DirSummary    map[string]int         `json:"dirSummary"`    // per-folder counts
	TotalExisting int                    `json:"totalExisting"` // existing titles preserved
	WSCount       int                    `json:"wsCount"`       // conversations with workspace
	KnownWSCount  int                    `json:"knownWSCount"`  // known workspace URIs loaded
}

// FixOptions is the input to RunFix.
type FixOptions struct {
	Mode string `json:"mode"` // "auto" or "auto_manual"
}

// DBResult is the result of writing to one database.
type DBResult struct {
	Path       string `json:"path"`
	AppName    string `json:"appName"`
	BackupFile string `json:"backupFile"`
}

// FixResult is returned by RunFix.
type FixResult struct {
	Success            bool       `json:"success"`
	Total              int        `json:"total"`
	WorkspaceMapped    int        `json:"workspaceMapped"`
	TimestampsInjected int        `json:"timestampsInjected"`
	DBResults          []DBResult `json:"dbResults"`
	Error              string     `json:"error,omitempty"`
}

// ── Bound Methods (called from frontend) ────────────────────────────────────

// GetSystemInfo detects and returns system path information.
func (a *App) GetSystemInfo() SystemInfo {
	a.pathInfo = platform.DetectPaths()
	return SystemInfo{
		OS:                  a.pathInfo.OS,
		DBPaths:             a.pathInfo.DBPaths,
		ConversationDirs:    a.pathInfo.ConversationDirs,
		BrainDirs:           a.pathInfo.BrainDirs,
		WorkspaceStorageDir: a.pathInfo.WorkspaceStorageDir,
	}
}

// ScanConversations scans all conversation directories and resolves titles.
func (a *App) ScanConversations() ScanResult {
	if a.pathInfo == nil {
		a.pathInfo = platform.DetectPaths()
	}

	a.emitLog("info", "Scanning conversation directories...")

	// Collect conversations from all dirs
	a.catalog = scanner.CollectConversations(a.pathInfo.ConversationDirs)
	a.sortedIDs = scanner.SortByModTime(a.catalog)
	dirSummary := scanner.GetDirSummary(a.catalog)

	a.emitLog("info", fmt.Sprintf("Found %d unique conversations", len(a.sortedIDs)))

	// Extract existing metadata from databases
	a.emitLog("info", "Reading metadata from database(s)...")
	a.titles, a.blobs = database.ExtractMetadataFromPaths(a.pathInfo.DBPaths)

	wsCount := 0
	for _, blob := range a.blobs {
		if workspace.ExtractWorkspaceHint(blob) != "" {
			wsCount++
		}
	}

	// Load known workspace URIs
	a.knownURIs = workspace.LoadKnownURIs(a.pathInfo.WorkspaceStorageDir)

	// Resolve titles for all conversations
	a.resolved = make([]ResolvedConversation, 0, len(a.sortedIDs))
	stats := map[string]int{"brain": 0, "preserved": 0, "fallback": 0}

	for i, cid := range a.sortedIDs {
		cf := a.catalog[cid]
		t, source := title.Resolve(cid, a.titles, a.pathInfo.BrainDirs, cf.FilePath)
		innerData := a.blobs[cid]
		hasWS := innerData != nil && workspace.ExtractWorkspaceHint(innerData) != ""

		a.resolved = append(a.resolved, ResolvedConversation{
			Index:        i + 1,
			ID:           cid,
			Title:        t,
			Source:       source,
			HasWorkspace: hasWS,
		})
		stats[source]++
	}

	a.emitLog("info", fmt.Sprintf("Titles: %d brain, %d preserved, %d fallback",
		stats["brain"], stats["preserved"], stats["fallback"]))

	return ScanResult{
		Conversations: a.resolved,
		Stats:         stats,
		DirSummary:    dirSummary,
		TotalExisting: len(a.titles),
		WSCount:       wsCount,
		KnownWSCount:  len(a.knownURIs),
	}
}

// RunFix executes the conversation fix process.
func (a *App) RunFix(options FixOptions) FixResult {
	if a.pathInfo == nil || len(a.resolved) == 0 {
		return FixResult{Success: false, Error: "Please scan conversations first"}
	}

	if len(a.pathInfo.DBPaths) == 0 {
		return FixResult{Success: false, Error: "No database found"}
	}

	total := len(a.resolved)
	a.emitProgress(0, total, "Starting fix process...", 0)

	// Pre-populate wsAssignments with manual selections
	wsAssignments := make(map[string]string)
	for cid, path := range a.manualWorkspaces {
		wsAssignments[cid] = path
	}

	unmappedCount := 0
	for _, conv := range a.resolved {
		// Skip if already manually assigned or has workspace
		if _, exists := wsAssignments[conv.ID]; exists || conv.HasWorkspace {
			continue
		}
		unmappedCount++
	}

	if unmappedCount > 0 {
		a.emitLog("info", fmt.Sprintf("Auto-assigning workspaces for %d conversations...", unmappedCount))
		autoCount := 0
		for _, conv := range a.resolved {
			if conv.HasWorkspace {
				continue
			}
			if _, exists := wsAssignments[conv.ID]; exists {
				continue
			}
			inferred := workspace.InferFromBrain(conv.ID, a.pathInfo.BrainDirs,
				a.knownURIs, a.pathInfo.OS, a.pathInfo.IsWSL)
			if inferred != "" {
				if workspace.IsRemoteURI(inferred) || dirExists(inferred) {
					wsAssignments[conv.ID] = inferred
					autoCount++
				}
			}
		}
		a.emitLog("info", fmt.Sprintf("Auto-assigned %d workspace(s)", autoCount))
	}

	// Build the new index
	a.emitProgress(0, total, "Building index...", 10)
	a.emitLog("info", "Building final index...")

	var resultBytes []byte
	wsTotal := 0
	tsInjected := 0

	for i, conv := range a.resolved {
		innerData := a.blobs[conv.ID]
		wsPath := wsAssignments[conv.ID]

		cf := a.catalog[conv.ID]
		var pbMtime float64
		if cf != nil {
			pbMtime = float64(cf.ModTime.Unix())
		}

		entry := buildTrajectoryEntry(conv.ID, conv.Title, innerData, wsPath,
			pbMtime, a.pathInfo.OS, a.pathInfo.IsWSL)
		resultBytes = append(resultBytes, pb.EncodeLengthDelimited(1, entry)...)

		if conv.HasWorkspace || wsPath != "" {
			wsTotal++
		}
		if pbMtime > 0 && (innerData == nil || !pb.HasTimestampFields(innerData)) {
			tsInjected++
		}

		if (i+1)%10 == 0 || i+1 == total {
			percent := int(float64(i+1) / float64(total) * 70) + 10
			a.emitProgress(i+1, total,
				fmt.Sprintf("Processing %d/%d...", i+1, total), percent)
		}
	}

	a.emitLog("info", fmt.Sprintf("Workspace: %d mapped | Timestamps injected: %d", wsTotal, tsInjected))

	// Write to all databases
	encoded := base64.StdEncoding.EncodeToString(resultBytes)
	a.emitProgress(total, total, "Writing to database(s)...", 85)

	var dbResults []DBResult
	backupDir, _ := os.Getwd()

	for _, dbPath := range a.pathInfo.DBPaths {
		appName := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(dbPath))))
		re := regexp.MustCompile(`[^A-Za-z0-9]+`)
		suffix := strings.ToLower(strings.Trim(re.ReplaceAllString(appName, "_"), "_"))

		backupFile, err := database.WriteIndex(dbPath, encoded, backupDir, suffix)
		if err != nil {
			a.emitLog("error", fmt.Sprintf("Error writing to %s: %v", appName, err))
			continue
		}

		a.emitLog("info", fmt.Sprintf("Updated: %s", appName))
		dbResults = append(dbResults, DBResult{
			Path:       dbPath,
			AppName:    appName,
			BackupFile: backupFile,
		})
	}

	a.emitProgress(total, total, "Complete!", 100)
	a.emitLog("info", fmt.Sprintf("SUCCESS! Rebuilt index with %d conversations.", total))

	return FixResult{
		Success:            true,
		Total:              total,
		WorkspaceMapped:    wsTotal,
		TimestampsInjected: tsInjected,
		DBResults:          dbResults,
	}
}

// CheckForUpdates checks GitHub for a newer release.
func (a *App) CheckForUpdates() *updater.UpdateInfo {
	return updater.CheckForUpdates()
}

// AssignWorkspace manually associates a workspace folder path with a conversation.
func (a *App) AssignWorkspace(cid, wsPath string) bool {
	if a.manualWorkspaces == nil {
		a.manualWorkspaces = make(map[string]string)
	}
	a.manualWorkspaces[cid] = wsPath

	// Update the resolved conversation state so it reflects in the next UI render or scan result
	for i, conv := range a.resolved {
		if conv.ID == cid {
			a.resolved[i].HasWorkspace = true
			return true
		}
	}
	return false
}

// SelectFolder opens a native folder selection dialog with a custom title.
func (a *App) SelectFolder(title string) string {
	if title == "" {
		title = "Select Workspace Folder"
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return ""
	}
	return dir
}

// ── Internal Helpers ────────────────────────────────────────────────────────

// buildTrajectoryEntry builds a single trajectory summary protobuf entry.
func buildTrajectoryEntry(conversationID, titleStr string, existingInner []byte,
	workspacePath string, pbMtime float64, system string, isWSL bool) []byte {

	var innerInfo []byte

	if existingInner != nil {
		// Strip old title (field 1) and prepend new title
		preservedFields := pb.StripField(existingInner, 1)
		innerInfo = append(pb.EncodeStringField(1, titleStr), preservedFields...)

		// Decode %20/%3A in existing workspace URIs
		if workspacePath == "" {
			existingWS := workspace.ExtractWorkspaceHint(innerInfo)
			if existingWS != "" && (strings.Contains(existingWS, "%20") ||
				strings.Contains(existingWS, "%3A") || strings.Contains(existingWS, "%3a")) {
				decodedWS, _ := url.PathUnescape(existingWS)
				if decodedWS != "" {
					innerInfo = pb.StripField(innerInfo, 9)
					innerInfo = append(innerInfo, workspace.BuildWorkspaceField(decodedWS, system, isWSL)...)
				}
			}
		}

		// Override workspace if assigned
		if workspacePath != "" {
			innerInfo = pb.StripField(innerInfo, 9)
			innerInfo = append(innerInfo, workspace.BuildWorkspaceField(workspacePath, system, isWSL)...)
		}

		// Inject timestamps if missing
		if pbMtime > 0 && !pb.HasTimestampFields(existingInner) {
			innerInfo = append(innerInfo, pb.BuildTimestampFields(int64(pbMtime))...)
		}
	} else {
		innerInfo = pb.EncodeStringField(1, titleStr)
		if workspacePath != "" {
			innerInfo = append(innerInfo, workspace.BuildWorkspaceField(workspacePath, system, isWSL)...)
		}
		if pbMtime > 0 {
			innerInfo = append(innerInfo, pb.BuildTimestampFields(int64(pbMtime))...)
		}
	}

	infoB64 := base64.StdEncoding.EncodeToString(innerInfo)
	subMessage := pb.EncodeStringField(1, infoB64)

	entry := pb.EncodeStringField(1, conversationID)
	entry = append(entry, pb.EncodeLengthDelimited(2, subMessage)...)
	return entry
}

// emitProgress sends a progress event to the frontend.
func (a *App) emitProgress(step, total int, message string, percent int) {
	runtime.EventsEmit(a.ctx, "fix:progress", map[string]interface{}{
		"step":    step,
		"total":   total,
		"message": message,
		"percent": percent,
	})
}

// emitLog sends a log event to the frontend.
func (a *App) emitLog(level, message string) {
	runtime.EventsEmit(a.ctx, "fix:log", map[string]interface{}{
		"level":   level,
		"message": message,
	})
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
