package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConvertFunc is the function signature for file conversion.
type ConvertFunc func(inputFile string) error

// Watcher watches files for changes and triggers conversions.
type Watcher struct {
	fsWatcher   *fsnotify.Watcher
	convertFunc ConvertFunc
	files       map[string]struct{}
	debounce    time.Duration
	mu          sync.Mutex
	lastEvent   map[string]time.Time
}

// New creates a new file watcher.
func New(convertFunc ConvertFunc) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Watcher{
		fsWatcher:   fsw,
		convertFunc: convertFunc,
		files:       make(map[string]struct{}),
		debounce:    100 * time.Millisecond,
		lastEvent:   make(map[string]time.Time),
	}, nil
}

// AddFile adds a file to be watched.
func (w *Watcher) AddFile(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", filePath, err)
	}

	// Watch the directory containing the file (for editor save patterns)
	dir := filepath.Dir(absPath)
	if err := w.fsWatcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", dir, err)
	}

	w.mu.Lock()
	w.files[absPath] = struct{}{}
	w.mu.Unlock()

	return nil
}

// Watch starts watching for file changes. Blocks until context is cancelled.
func (w *Watcher) Watch(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return w.Close()

		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}
			w.handleEvent(event)

		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}

// handleEvent processes a file system event.
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Only handle write and create events
	if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
		return
	}

	absPath, err := filepath.Abs(event.Name)
	if err != nil {
		return
	}

	w.mu.Lock()
	_, isWatched := w.files[absPath]
	lastTime := w.lastEvent[absPath]
	w.mu.Unlock()

	if !isWatched {
		return
	}

	// Debounce: ignore events that happen too close together
	if time.Since(lastTime) < w.debounce {
		return
	}

	w.mu.Lock()
	w.lastEvent[absPath] = time.Now()
	w.mu.Unlock()

	// Small delay to ensure file write is complete
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("\nFile changed: %s\n", filepath.Base(absPath))
	fmt.Printf("Re-converting...\n")

	if err := w.convertFunc(absPath); err != nil {
		fmt.Printf("Conversion error: %v\n", err)
	} else {
		fmt.Printf("Conversion complete.\n")
	}
}

// Close stops the watcher and releases resources.
func (w *Watcher) Close() error {
	return w.fsWatcher.Close()
}
