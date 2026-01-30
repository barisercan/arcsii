package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file change event
type FileEvent struct {
	Path      string
	Name      string
	Operation string
	Time      time.Time
	Size      int64
	IsGitOp   bool
	GitOp     string // "commit", "push", "pull", "merge", etc.
	Preview   []string // Preview lines of the change
}

// Watcher watches for file changes
type Watcher struct {
	watcher    *fsnotify.Watcher
	root       string
	Events     chan FileEvent
	Errors     chan error
	done       chan bool
	WatchCount int // Number of directories being watched
}

// New creates a new file watcher
func New(root string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:    fsWatcher,
		root:       root,
		Events:     make(chan FileEvent, 100),
		Errors:     make(chan error, 10),
		done:       make(chan bool),
		WatchCount: 0,
	}

	// Get absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}
	w.root = absRoot

	// Add all directories recursively
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		name := info.Name()
		// Skip common ignore patterns but NOT .git (we want to watch it for git ops)
		if name == "node_modules" || name == "vendor" || name == "dist" || name == "__pycache__" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Allow .git and all its subdirectories
		inGitDir := strings.Contains(path, ".git")

		// Skip hidden files/dirs except .git and its contents
		if strings.HasPrefix(name, ".") && name != ".git" && !inGitDir {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if err := fsWatcher.Add(path); err == nil {
				w.WatchCount++
			}
		}
		return nil
	})

	if err != nil {
		fsWatcher.Close()
		return nil, err
	}

	// Explicitly watch key .git directories for git operations
	gitDirs := []string{
		filepath.Join(absRoot, ".git"),
		filepath.Join(absRoot, ".git", "refs"),
		filepath.Join(absRoot, ".git", "refs", "heads"),
		filepath.Join(absRoot, ".git", "refs", "remotes"),
		filepath.Join(absRoot, ".git", "logs"),
		filepath.Join(absRoot, ".git", "logs", "refs"),
		filepath.Join(absRoot, ".git", "logs", "refs", "heads"),
	}
	for _, dir := range gitDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			fsWatcher.Add(dir)
		}
	}

	return w, nil
}

// Start begins watching for file changes
func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					return
				}

				name := filepath.Base(event.Name)

				// Skip temp files used by editors for safe writes
				if strings.Contains(name, ".tmp") || strings.HasSuffix(name, "~") || strings.HasPrefix(name, "#") {
					continue
				}

				// Check if this is a git operation
				isGitOp := false
				gitOp := ""
				if strings.Contains(event.Name, ".git") {
					isGitOp = true
					gitOp = detectGitOperation(event.Name, name)
					if gitOp == "" {
						continue // Skip uninteresting git file changes
					}
				} else if strings.HasPrefix(name, ".") {
					continue // Skip other hidden files
				}

				var op string
				switch {
				case event.Op&fsnotify.Create == fsnotify.Create:
					op = "created"
					// If it's a new directory, watch it
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						w.watcher.Add(event.Name)
					}
				case event.Op&fsnotify.Write == fsnotify.Write:
					op = "modified"
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					op = "deleted"
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					op = "renamed"
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					continue // Skip chmod events
				default:
					continue
				}

				var size int64
				if info, err := os.Stat(event.Name); err == nil {
					size = info.Size()
				}

				rel, _ := filepath.Rel(w.root, event.Name)
				if rel == "" {
					rel = event.Name
				}

				// Get preview for non-git file changes
				var preview []string
				if !isGitOp && (op == "modified" || op == "created") {
					preview = getFilePreview(event.Name, 3)
				}

				w.Events <- FileEvent{
					Path:      rel,
					Name:      name,
					Operation: op,
					Time:      time.Now(),
					Size:      size,
					IsGitOp:   isGitOp,
					GitOp:     gitOp,
					Preview:   preview,
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					return
				}
				w.Errors <- err

			case <-w.done:
				return
			}
		}
	}()
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.done <- true
	w.watcher.Close()
}

// getFilePreview reads the last few lines of a file for preview
func getFilePreview(path string, numLines int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Skip binary files
	if isBinary(data) {
		return []string{"[binary file]"}
	}

	lines := strings.Split(string(data), "\n")

	// Get last N non-empty lines
	var preview []string
	for i := len(lines) - 1; i >= 0 && len(preview) < numLines; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			// Truncate long lines
			if len(line) > 60 {
				line = line[:57] + "..."
			}
			preview = append([]string{line}, preview...)
		}
	}

	return preview
}

// isBinary checks if data appears to be binary
func isBinary(data []byte) bool {
	if len(data) > 512 {
		data = data[:512]
	}
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// detectGitOperation identifies git operations from file changes
func detectGitOperation(path, name string) string {
	// Skip lock files
	if strings.HasSuffix(name, ".lock") {
		return ""
	}

	// Detect commit - COMMIT_EDITMSG is created when committing
	if name == "COMMIT_EDITMSG" || name == "MERGE_MSG" {
		return "commit"
	}

	// refs/heads changes indicate a commit was made
	if strings.Contains(path, "refs/heads") || strings.Contains(path, "logs/refs/heads") {
		return "commit"
	}

	// Detect push/pull (refs/remotes changes)
	if strings.Contains(path, "refs/remotes") || strings.Contains(path, "logs/refs/remotes") {
		return "push"
	}
	if name == "FETCH_HEAD" {
		return "fetch"
	}
	if name == "ORIG_HEAD" {
		return "pull"
	}

	// Detect merge
	if name == "MERGE_HEAD" {
		return "merge"
	}

	// Detect rebase
	if strings.Contains(path, "rebase-merge") || strings.Contains(path, "rebase-apply") {
		return "rebase"
	}

	// Detect stash
	if strings.Contains(path, "refs/stash") || name == "stash" {
		return "stash"
	}

	// Detect checkout/branch - HEAD file changes
	if name == "HEAD" && !strings.Contains(path, "logs") {
		return "checkout"
	}

	return "" // Not an interesting git operation
}
