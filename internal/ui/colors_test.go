package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutput_ColorsDisabledByDefault(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// When using buffer writers (not a real TTY), colors should be disabled
	o := NewOutputWithWriters(stdout, stderr)

	if o.IsTTY() {
		t.Error("expected IsTTY() to return false for buffer writer")
	}

	if o.ColorsEnabled() {
		t.Error("expected colors to be disabled for non-TTY writer")
	}
}

func TestOutput_Error(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Error("test error message")

	if !strings.Contains(stderr.String(), "test error message") {
		t.Errorf("expected stderr to contain 'test error message', got: %s", stderr.String())
	}
}

func TestOutput_Errorf(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Errorf("error: %s", "details")

	output := stderr.String()
	if !strings.Contains(output, "Error:") {
		t.Errorf("expected stderr to contain 'Error:', got: %s", output)
	}
	if !strings.Contains(output, "details") {
		t.Errorf("expected stderr to contain 'details', got: %s", output)
	}
}

func TestOutput_Warn(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Warn("test warning")

	if !strings.Contains(stderr.String(), "test warning") {
		t.Errorf("expected stderr to contain 'test warning', got: %s", stderr.String())
	}
}

func TestOutput_Warnf(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Warnf("warning: %s", "info")

	output := stderr.String()
	if !strings.Contains(output, "Warning:") {
		t.Errorf("expected stderr to contain 'Warning:', got: %s", output)
	}
}

func TestOutput_Success(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Success("operation complete")

	if !strings.Contains(stdout.String(), "operation complete") {
		t.Errorf("expected stdout to contain 'operation complete', got: %s", stdout.String())
	}
}

func TestOutput_Successf(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Successf("converted %d files", 5)

	if !strings.Contains(stdout.String(), "converted 5 files") {
		t.Errorf("expected stdout to contain 'converted 5 files', got: %s", stdout.String())
	}
}

func TestOutput_Info(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Info("info message")

	if !strings.Contains(stdout.String(), "info message") {
		t.Errorf("expected stdout to contain 'info message', got: %s", stdout.String())
	}
}

func TestOutput_Print(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Print("plain text %d", 42)

	if !strings.Contains(stdout.String(), "plain text 42") {
		t.Errorf("expected stdout to contain 'plain text 42', got: %s", stdout.String())
	}
}

func TestOutput_Println(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)
	o.Println("line 1")

	if !strings.Contains(stdout.String(), "line 1\n") {
		t.Errorf("expected stdout to contain 'line 1\\n', got: %s", stdout.String())
	}
}

func TestOutput_SetColorsEnabled(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	// Initially disabled for non-TTY
	if o.ColorsEnabled() {
		t.Error("expected colors to be disabled initially")
	}

	// Explicitly enable
	o.SetColorsEnabled(true)
	if !o.ColorsEnabled() {
		t.Error("expected colors to be enabled after SetColorsEnabled(true)")
	}

	// Explicitly disable
	o.SetColorsEnabled(false)
	if o.ColorsEnabled() {
		t.Error("expected colors to be disabled after SetColorsEnabled(false)")
	}
}

func TestOutput_Bold(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	// Without colors, should return plain text
	result := o.Bold("important")
	if result != "important" {
		t.Errorf("expected 'important', got: %s", result)
	}
}

func TestOutput_Highlight(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	// Without colors, should return plain text
	result := o.Highlight("highlighted")
	if result != "highlighted" {
		t.Errorf("expected 'highlighted', got: %s", result)
	}
}

func TestOutput_FileName(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	result := o.FileName("test.md")
	if result != "test.md" {
		t.Errorf("expected 'test.md', got: %s", result)
	}
}

func TestOutput_Number(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	result := o.Number(42)
	if result != "42" {
		t.Errorf("expected '42', got: %s", result)
	}
}

func TestOutput_Writers(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	o := NewOutputWithWriters(stdout, stderr)

	if o.Stdout() != stdout {
		t.Error("expected Stdout() to return the provided stdout writer")
	}

	if o.Stderr() != stderr {
		t.Error("expected Stderr() to return the provided stderr writer")
	}
}
