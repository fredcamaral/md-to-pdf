package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name     string
		jsonMode bool
	}{
		{"text mode", false},
		{"json mode", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.jsonMode)
			if f == nil {
				t.Fatal("NewFormatter returned nil")
			}
			if f.IsJSON() != tt.jsonMode {
				t.Errorf("IsJSON() = %v, want %v", f.IsJSON(), tt.jsonMode)
			}
			if len(f.Results()) != 0 {
				t.Errorf("Results() should be empty, got %d items", len(f.Results()))
			}
		})
	}
}

func TestRecordSuccess(t *testing.T) {
	f := NewFormatter(true)
	duration := 100 * time.Millisecond

	f.RecordSuccess("input.md", "output.pdf", duration)

	results := f.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if !r.Success {
		t.Error("Success should be true")
	}
	if r.Input != "input.md" {
		t.Errorf("Input = %q, want %q", r.Input, "input.md")
	}
	if r.Output != "output.pdf" {
		t.Errorf("Output = %q, want %q", r.Output, "output.pdf")
	}
	if r.DurationMs != 100 {
		t.Errorf("DurationMs = %d, want 100", r.DurationMs)
	}
	if r.Error != "" {
		t.Errorf("Error should be empty, got %q", r.Error)
	}
}

func TestRecordError(t *testing.T) {
	f := NewFormatter(true)
	duration := 50 * time.Millisecond
	err := &testError{"test error"}

	f.RecordError("input.md", duration, err)

	results := f.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Success {
		t.Error("Success should be false")
	}
	if r.Input != "input.md" {
		t.Errorf("Input = %q, want %q", r.Input, "input.md")
	}
	if r.DurationMs != 50 {
		t.Errorf("DurationMs = %d, want 50", r.DurationMs)
	}
	if r.Error != "test error" {
		t.Errorf("Error = %q, want %q", r.Error, "test error")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Formatter)
		wantError bool
	}{
		{
			name:      "no results",
			setup:     func(f *Formatter) {},
			wantError: false,
		},
		{
			name: "only successes",
			setup: func(f *Formatter) {
				f.RecordSuccess("a.md", "a.pdf", time.Millisecond)
				f.RecordSuccess("b.md", "b.pdf", time.Millisecond)
			},
			wantError: false,
		},
		{
			name: "only errors",
			setup: func(f *Formatter) {
				f.RecordError("a.md", time.Millisecond, &testError{"err"})
			},
			wantError: true,
		},
		{
			name: "mixed",
			setup: func(f *Formatter) {
				f.RecordSuccess("a.md", "a.pdf", time.Millisecond)
				f.RecordError("b.md", time.Millisecond, &testError{"err"})
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(true)
			tt.setup(f)
			if got := f.HasErrors(); got != tt.wantError {
				t.Errorf("HasErrors() = %v, want %v", got, tt.wantError)
			}
		})
	}
}

func TestPrint_SingleResult(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(true)
	f.SetWriter(&buf)

	f.RecordSuccess("test.md", "test.pdf", 100*time.Millisecond)

	err := f.Print()
	if err != nil {
		t.Fatalf("Print() error: %v", err)
	}

	// Parse the output as JSON
	var result ConversionResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	if !result.Success {
		t.Error("result.Success should be true")
	}
	if result.Input != "test.md" {
		t.Errorf("result.Input = %q, want %q", result.Input, "test.md")
	}
	if result.Output != "test.pdf" {
		t.Errorf("result.Output = %q, want %q", result.Output, "test.pdf")
	}
}

func TestPrint_BatchResult(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(true)
	f.SetWriter(&buf)

	f.RecordSuccess("a.md", "a.pdf", 100*time.Millisecond)
	f.RecordSuccess("b.md", "b.pdf", 200*time.Millisecond)
	f.RecordError("c.md", 50*time.Millisecond, &testError{"failed"})

	err := f.Print()
	if err != nil {
		t.Fatalf("Print() error: %v", err)
	}

	// Parse the output as JSON
	var batch BatchResult
	if err := json.Unmarshal(buf.Bytes(), &batch); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	if len(batch.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(batch.Results))
	}
	if batch.Summary.Total != 3 {
		t.Errorf("Summary.Total = %d, want 3", batch.Summary.Total)
	}
	if batch.Summary.Succeeded != 2 {
		t.Errorf("Summary.Succeeded = %d, want 2", batch.Summary.Succeeded)
	}
	if batch.Summary.Failed != 1 {
		t.Errorf("Summary.Failed = %d, want 1", batch.Summary.Failed)
	}
	if batch.Summary.TotalMs != 350 {
		t.Errorf("Summary.TotalMs = %d, want 350", batch.Summary.TotalMs)
	}
}

func TestPrint_TextMode(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(false) // text mode
	f.SetWriter(&buf)

	f.RecordSuccess("test.md", "test.pdf", 100*time.Millisecond)

	err := f.Print()
	if err != nil {
		t.Fatalf("Print() error: %v", err)
	}

	// In text mode, Print should not output anything
	if buf.Len() != 0 {
		t.Errorf("expected no output in text mode, got: %s", buf.String())
	}
}

func TestConversionResultJSON(t *testing.T) {
	result := ConversionResult{
		Success:       true,
		Input:         "doc.md",
		Output:        "doc.pdf",
		DurationMs:    234,
		FileSizeBytes: 45000,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"success":true`) {
		t.Errorf("missing success field in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"input":"doc.md"`) {
		t.Errorf("missing input field in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"output":"doc.pdf"`) {
		t.Errorf("missing output field in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"duration_ms":234`) {
		t.Errorf("missing duration_ms field in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"file_size_bytes":45000`) {
		t.Errorf("missing file_size_bytes field in JSON: %s", jsonStr)
	}
}

func TestConversionResultJSON_ErrorCase(t *testing.T) {
	result := ConversionResult{
		Success:    false,
		Input:      "bad.md",
		DurationMs: 10,
		Error:      "file not found",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"success":false`) {
		t.Errorf("missing success field in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"error":"file not found"`) {
		t.Errorf("missing error field in JSON: %s", jsonStr)
	}
	// Output should be omitted (omitempty)
	if strings.Contains(jsonStr, `"output":`) {
		t.Errorf("output should be omitted for error: %s", jsonStr)
	}
}
