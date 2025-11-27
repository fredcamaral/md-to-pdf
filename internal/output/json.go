package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// ConversionResult represents the result of a single file conversion.
type ConversionResult struct {
	Success       bool   `json:"success"`
	Input         string `json:"input"`
	Output        string `json:"output,omitempty"`
	DurationMs    int64  `json:"duration_ms"`
	FileSizeBytes int64  `json:"file_size_bytes,omitempty"`
	Error         string `json:"error,omitempty"`
}

// BatchResult represents results for multiple conversions.
type BatchResult struct {
	Results []ConversionResult `json:"results"`
	Summary Summary            `json:"summary"`
}

// Summary provides aggregate statistics for batch conversions.
type Summary struct {
	Total      int   `json:"total"`
	Succeeded  int   `json:"succeeded"`
	Failed     int   `json:"failed"`
	TotalMs    int64 `json:"total_duration_ms"`
	TotalBytes int64 `json:"total_size_bytes"`
}

// Formatter handles output formatting.
type Formatter struct {
	writer   io.Writer
	jsonMode bool
	results  []ConversionResult
}

// NewFormatter creates a new output formatter.
func NewFormatter(jsonMode bool) *Formatter {
	return &Formatter{
		writer:   os.Stdout,
		jsonMode: jsonMode,
		results:  make([]ConversionResult, 0),
	}
}

// SetWriter sets the output writer (useful for testing).
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// IsJSON returns true if JSON mode is enabled.
func (f *Formatter) IsJSON() bool {
	return f.jsonMode
}

// RecordSuccess records a successful conversion.
func (f *Formatter) RecordSuccess(input, output string, duration time.Duration) {
	var fileSize int64
	if info, err := os.Stat(output); err == nil {
		fileSize = info.Size()
	}

	result := ConversionResult{
		Success:       true,
		Input:         input,
		Output:        output,
		DurationMs:    duration.Milliseconds(),
		FileSizeBytes: fileSize,
	}
	f.results = append(f.results, result)
}

// RecordError records a failed conversion.
func (f *Formatter) RecordError(input string, duration time.Duration, err error) {
	result := ConversionResult{
		Success:    false,
		Input:      input,
		DurationMs: duration.Milliseconds(),
		Error:      err.Error(),
	}
	f.results = append(f.results, result)
}

// Print outputs results in the appropriate format.
func (f *Formatter) Print() error {
	if !f.jsonMode {
		return nil // Text output is handled elsewhere
	}

	if len(f.results) == 1 {
		return f.printJSON(f.results[0])
	}

	return f.printBatchJSON()
}

// printJSON outputs a single result as JSON.
func (f *Formatter) printJSON(result ConversionResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if _, err := fmt.Fprintln(f.writer, string(data)); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}

// printBatchJSON outputs batch results as JSON.
func (f *Formatter) printBatchJSON() error {
	summary := Summary{
		Total: len(f.results),
	}

	for _, r := range f.results {
		if r.Success {
			summary.Succeeded++
			summary.TotalBytes += r.FileSizeBytes
		} else {
			summary.Failed++
		}
		summary.TotalMs += r.DurationMs
	}

	batch := BatchResult{
		Results: f.results,
		Summary: summary,
	}

	data, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if _, err := fmt.Fprintln(f.writer, string(data)); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}

// HasErrors returns true if any conversion failed.
func (f *Formatter) HasErrors() bool {
	for _, r := range f.results {
		if !r.Success {
			return true
		}
	}
	return false
}

// Results returns all recorded results.
func (f *Formatter) Results() []ConversionResult {
	return f.results
}
