package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestProgress_DisabledForNonTTY(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	p := NewProgress(o)

	if p.IsEnabled() {
		t.Error("expected progress to be disabled for non-TTY output")
	}
}

func TestProgress_SetEnabled(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	p := NewProgress(o)

	// Cannot enable for non-TTY
	p.SetEnabled(true)
	if p.IsEnabled() {
		t.Error("expected progress to remain disabled for non-TTY even when SetEnabled(true)")
	}
}

func TestProgress_StartStopSafeForNonTTY(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	p := NewProgress(o)

	// These should not panic for non-TTY
	p.Start("Processing...")
	p.Update("Still processing...")
	p.Stop()
	p.StopWithSuccess("Done!")
	p.StopWithError("Failed!")
}

func TestBatchProgress_DisabledForNonTTY(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 5)

	if bp.IsEnabled() {
		t.Error("expected batch progress to be disabled for non-TTY output")
	}
}

func TestBatchProgress_SetEnabled(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 5)

	// Cannot enable for non-TTY
	bp.SetEnabled(true)
	if bp.IsEnabled() {
		t.Error("expected batch progress to remain disabled for non-TTY")
	}
}

func TestBatchProgress_FileComplete(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 3)

	// For non-TTY, FileComplete should print to stdout
	bp.StartFile("test1.md")
	bp.FileComplete("test1.md", "test1.pdf")

	output := stdout.String()
	if !strings.Contains(output, "Converted: test1.md -> test1.pdf") {
		t.Errorf("expected FileComplete output, got: %s", output)
	}
}

func TestBatchProgress_Complete(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 3)

	// These should not panic
	bp.StartFile("test.md")
	bp.Complete()
}

func TestBatchProgress_CompleteWithMessage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 1)

	bp.StartFile("test.md")
	bp.CompleteWithMessage("Custom complete message")

	// No output for non-TTY with CompleteWithMessage
	// (the message is only shown when progress indicators are visible)
}

func TestBatchProgress_Error(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	bp := NewBatchProgress(o, 3)

	bp.StartFile("test.md")
	bp.Error(nil)

	// Should print error prefix to stderr
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("expected error output, got: %s", stderr.String())
	}
}
