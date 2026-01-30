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
}

// Watcher watches for file changes
type Watcher struct {
	watcher *fsnotify.Watcher
	root    string
	Events  chan FileEvent
	Errors  chan error
	done    chan bool
}

// New creates a new file watcher
func New(root string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher: fsWatcher,
		root:    root,
		Events:  make(chan FileEvent, 100),
		Errors:  make(chan error, 10),
		done:    make(chan bool),
	}

	// Add all directories recursively
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

		// Skip hidden files/dirs except .git
		if strings.HasPrefix(name, ".") && name != ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return fsWatcher.Add(path)
		}
		return nil
	})

	if err != nil {
		fsWatcher.Close()
		return nil, err
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

				w.Events <- FileEvent{
					Path:      rel,
					Name:      name,
					Operation: op,
					Time:      time.Now(),
					Size:      size,
					IsGitOp:   isGitOp,
					GitOp:     gitOp,
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

// detectGitOperation identifies git operations from file changes
func detectGitOperation(path, name string) string {
	// Detect commit
	if name == "COMMIT_EDITMSG" || name == "MERGE_MSG" {
		return "commit"
	}
	if strings.Contains(path, "refs/heads") && !strings.HasSuffix(name, ".lock") {
		return "commit"
	}

	// Detect push/pull (refs/remotes changes)
	if strings.Contains(path, "refs/remotes") {
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
	if strings.Contains(path, "refs/stash") {
		return "stash"
	}

	// Detect checkout/branch
	if name == "HEAD" {
		return "checkout"
	}

	return "" // Not an interesting git operation
}
