package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	called := false
	convertFunc := func(inputFile string) error {
		called = true
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	if w.fsWatcher == nil {
		t.Error("fsWatcher should not be nil")
	}
	if w.files == nil {
		t.Error("files map should not be nil")
	}
	if w.debounce != 100*time.Millisecond {
		t.Errorf("debounce = %v, want 100ms", w.debounce)
	}
	if called {
		t.Error("convertFunc should not have been called yet")
	}
}

func TestAddFile(t *testing.T) {
	convertFunc := func(inputFile string) error {
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Add the file
	err = w.AddFile(tmpFile)
	if err != nil {
		t.Errorf("AddFile() failed: %v", err)
	}

	// Verify the file was added
	absPath, _ := filepath.Abs(tmpFile)
	if _, ok := w.files[absPath]; !ok {
		t.Error("file was not added to watch list")
	}
}

func TestAddFile_NonExistentDir(t *testing.T) {
	convertFunc := func(inputFile string) error {
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	// Try to add a file in a non-existent directory
	err = w.AddFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("AddFile() should fail for non-existent directory")
	}
}

func TestWatch_ContextCancellation(t *testing.T) {
	convertFunc := func(inputFile string) error {
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Watch should return immediately
	done := make(chan error, 1)
	go func() {
		done <- w.Watch(ctx)
	}()

	select {
	case <-done:
		// Good - Watch returned
	case <-time.After(time.Second):
		t.Error("Watch did not return after context cancellation")
	}
}

func TestWatch_FileChange(t *testing.T) {
	var callCount int32

	convertFunc := func(inputFile string) error {
		atomic.AddInt32(&callCount, 1)
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := w.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Add the file
	if err := w.AddFile(tmpFile); err != nil {
		t.Fatalf("AddFile() failed: %v", err)
	}

	// Start watching in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Watch(ctx)
	}()

	// Give the watcher time to start
	time.Sleep(200 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(tmpFile, []byte("# Modified"), 0644); err != nil {
		t.Fatalf("failed to modify temp file: %v", err)
	}

	// Wait for the conversion to be triggered
	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt32(&callCount)
	if count == 0 {
		t.Error("convertFunc was not called after file modification")
	}
}

func TestClose(t *testing.T) {
	convertFunc := func(inputFile string) error {
		return nil
	}

	w, err := New(convertFunc)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}
