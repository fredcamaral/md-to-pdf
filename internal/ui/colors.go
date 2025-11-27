// Package ui provides terminal UI utilities including colored output and progress indicators.
package ui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// Output provides colored terminal output with TTY detection.
type Output struct {
	// Writers for different output types
	stdout io.Writer
	stderr io.Writer

	// Color functions
	errorColor   *color.Color
	warnColor    *color.Color
	successColor *color.Color
	infoColor    *color.Color
	boldColor    *color.Color
	dimColor     *color.Color

	// State
	colorsEnabled bool
	isTTY         bool
}

// NewOutput creates a new Output with automatic TTY and NO_COLOR detection.
func NewOutput() *Output {
	return NewOutputWithWriters(os.Stdout, os.Stderr)
}

// NewOutputWithWriters creates a new Output with custom writers (useful for testing).
func NewOutputWithWriters(stdout, stderr io.Writer) *Output {
	o := &Output{
		stdout: stdout,
		stderr: stderr,
	}

	// Detect TTY status
	if f, ok := stdout.(*os.File); ok {
		o.isTTY = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}

	// Check NO_COLOR environment variable (https://no-color.org/)
	_, noColor := os.LookupEnv("NO_COLOR")

	// Enable colors only if TTY and NO_COLOR not set
	o.colorsEnabled = o.isTTY && !noColor

	// Initialize color functions
	o.initColors()

	return o
}

// initColors sets up the color functions based on whether colors are enabled.
func (o *Output) initColors() {
	if o.colorsEnabled {
		o.errorColor = color.New(color.FgRed)
		o.warnColor = color.New(color.FgYellow)
		o.successColor = color.New(color.FgGreen)
		o.infoColor = color.New(color.FgCyan)
		o.boldColor = color.New(color.Bold)
		o.dimColor = color.New(color.Faint)
	} else {
		// No-op colors when disabled
		o.errorColor = color.New()
		o.warnColor = color.New()
		o.successColor = color.New()
		o.infoColor = color.New()
		o.boldColor = color.New()
		o.dimColor = color.New()

		// Disable color output
		color.NoColor = true
	}
}

// SetColorsEnabled explicitly enables or disables colors.
func (o *Output) SetColorsEnabled(enabled bool) {
	o.colorsEnabled = enabled
	color.NoColor = !enabled
	o.initColors()
}

// IsTTY returns true if stdout is a terminal.
func (o *Output) IsTTY() bool {
	return o.isTTY
}

// ColorsEnabled returns true if colored output is enabled.
func (o *Output) ColorsEnabled() bool {
	return o.colorsEnabled
}

// Error prints an error message to stderr in red.
func (o *Output) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.errorColor.Fprint(o.stderr, msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		_, _ = fmt.Fprintln(o.stderr)
	}
}

// Errorf prints a formatted error message to stderr in red with "Error: " prefix.
func (o *Output) Errorf(format string, args ...interface{}) {
	_, _ = o.errorColor.Fprint(o.stderr, "Error: ")
	_, _ = fmt.Fprintf(o.stderr, format, args...)
	_, _ = fmt.Fprintln(o.stderr)
}

// Warn prints a warning message to stderr in yellow.
func (o *Output) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.warnColor.Fprint(o.stderr, msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		_, _ = fmt.Fprintln(o.stderr)
	}
}

// Warnf prints a formatted warning message to stderr in yellow with "Warning: " prefix.
func (o *Output) Warnf(format string, args ...interface{}) {
	_, _ = o.warnColor.Fprint(o.stderr, "Warning: ")
	_, _ = fmt.Fprintf(o.stderr, format, args...)
	_, _ = fmt.Fprintln(o.stderr)
}

// Success prints a success message to stdout in green.
func (o *Output) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.successColor.Fprint(o.stdout, msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		_, _ = fmt.Fprintln(o.stdout)
	}
}

// Successf prints a formatted success message with a checkmark prefix.
func (o *Output) Successf(format string, args ...interface{}) {
	if o.colorsEnabled {
		_, _ = o.successColor.Fprint(o.stdout, "[OK] ")
	}
	_, _ = fmt.Fprintf(o.stdout, format, args...)
	_, _ = fmt.Fprintln(o.stdout)
}

// Info prints an info message to stdout in cyan.
func (o *Output) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = o.infoColor.Fprint(o.stdout, msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		_, _ = fmt.Fprintln(o.stdout)
	}
}

// Infof prints a formatted info message to stdout in cyan.
func (o *Output) Infof(format string, args ...interface{}) {
	_, _ = o.infoColor.Fprintf(o.stdout, format, args...)
	_, _ = fmt.Fprintln(o.stdout)
}

// Bold prints bold text to stdout.
func (o *Output) Bold(format string, args ...interface{}) string {
	return o.boldColor.Sprintf(format, args...)
}

// Dim prints dimmed text to stdout.
func (o *Output) Dim(format string, args ...interface{}) string {
	return o.dimColor.Sprintf(format, args...)
}

// Print prints plain text to stdout.
func (o *Output) Print(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(o.stdout, format, args...)
}

// Println prints plain text to stdout with a newline.
func (o *Output) Println(args ...interface{}) {
	_, _ = fmt.Fprintln(o.stdout, args...)
}

// Stdout returns the stdout writer.
func (o *Output) Stdout() io.Writer {
	return o.stdout
}

// Stderr returns the stderr writer.
func (o *Output) Stderr() io.Writer {
	return o.stderr
}

// Highlight returns text formatted as highlighted (bold + cyan).
func (o *Output) Highlight(text string) string {
	if o.colorsEnabled {
		return color.New(color.Bold, color.FgCyan).Sprint(text)
	}
	return text
}

// FileName returns a filename formatted with bold styling.
func (o *Output) FileName(name string) string {
	return o.Bold(name)
}

// Number returns a number formatted with cyan styling.
func (o *Output) Number(n int) string {
	if o.colorsEnabled {
		return o.infoColor.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d", n)
}
