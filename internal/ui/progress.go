package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// Progress provides progress indication for long-running operations.
type Progress struct {
	output   *Output
	spinner  *spinner.Spinner
	enabled  bool
	isActive bool
}

// NewProgress creates a new Progress indicator.
// Progress is automatically disabled if output is not a TTY.
func NewProgress(output *Output) *Progress {
	p := &Progress{
		output:  output,
		enabled: output.IsTTY(),
	}

	if p.enabled {
		// Use a clean character set that works across platforms
		charSet := spinner.CharSets[14] // dots
		p.spinner = spinner.New(charSet, 100*time.Millisecond, spinner.WithWriter(output.Stdout()))
		if output.ColorsEnabled() {
			_ = p.spinner.Color("cyan")
		}
	}

	return p
}

// SetEnabled explicitly enables or disables the progress indicator.
func (p *Progress) SetEnabled(enabled bool) {
	p.enabled = enabled && p.output.IsTTY()
}

// IsEnabled returns true if progress indication is enabled.
func (p *Progress) IsEnabled() bool {
	return p.enabled
}

// Start begins showing a progress spinner with the given message.
func (p *Progress) Start(message string) {
	if !p.enabled || p.spinner == nil {
		return
	}
	p.spinner.Suffix = " " + message
	p.spinner.Start()
	p.isActive = true
}

// Update changes the progress message without stopping the spinner.
func (p *Progress) Update(message string) {
	if !p.enabled || p.spinner == nil || !p.isActive {
		return
	}
	p.spinner.Suffix = " " + message
}

// Stop stops the spinner.
func (p *Progress) Stop() {
	if !p.enabled || p.spinner == nil || !p.isActive {
		return
	}
	p.spinner.Stop()
	p.isActive = false
}

// StopWithSuccess stops the spinner and shows a success message.
func (p *Progress) StopWithSuccess(message string) {
	p.Stop()
	if p.enabled && p.output != nil {
		p.output.Success(message)
	}
}

// StopWithError stops the spinner and shows an error message.
func (p *Progress) StopWithError(message string) {
	p.Stop()
	if p.output != nil {
		p.output.Error(message)
	}
}

// BatchProgress tracks progress across multiple files in a batch operation.
type BatchProgress struct {
	progress *Progress
	output   *Output
	total    int
	current  int
	enabled  bool
}

// NewBatchProgress creates a progress tracker for batch operations.
func NewBatchProgress(output *Output, total int) *BatchProgress {
	return &BatchProgress{
		progress: NewProgress(output),
		output:   output,
		total:    total,
		current:  0,
		enabled:  output.IsTTY(),
	}
}

// SetEnabled explicitly enables or disables batch progress.
func (b *BatchProgress) SetEnabled(enabled bool) {
	b.enabled = enabled && b.output.IsTTY()
	b.progress.SetEnabled(enabled)
}

// IsEnabled returns true if batch progress is enabled.
func (b *BatchProgress) IsEnabled() bool {
	return b.enabled
}

// StartFile begins processing a file and updates progress.
func (b *BatchProgress) StartFile(filename string) {
	b.current++
	if b.enabled && b.total > 1 {
		msg := fmt.Sprintf("Converting %d/%d: %s", b.current, b.total, filename)
		if b.current == 1 {
			b.progress.Start(msg)
		} else {
			b.progress.Update(msg)
		}
	} else if b.enabled {
		b.progress.Start(fmt.Sprintf("Converting %s", filename))
	}
}

// FileComplete marks the current file as complete (for non-TTY output).
func (b *BatchProgress) FileComplete(filename string, outputPath string) {
	if !b.enabled {
		// For non-TTY, print completion message directly
		b.output.Print("Converted: %s -> %s\n", filename, outputPath)
	}
}

// Complete stops progress and shows a summary.
func (b *BatchProgress) Complete() {
	b.progress.Stop()
	if b.enabled && b.total > 1 {
		b.output.Successf("Converted %d files successfully", b.total)
	}
}

// CompleteWithMessage stops progress and shows a custom message.
func (b *BatchProgress) CompleteWithMessage(message string) {
	b.progress.Stop()
	if b.enabled {
		b.output.Success(message)
	}
}

// Error stops progress and shows an error.
func (b *BatchProgress) Error(err error) {
	b.progress.Stop()
	b.output.Errorf("%v", err)
}
