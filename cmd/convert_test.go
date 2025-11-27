package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fredcamaral/md-to-pdf/internal/config"
	"github.com/fredcamaral/md-to-pdf/internal/core"
)

func TestConfigMergingPriority(t *testing.T) {
	// Test that CLI flags override user config which overrides defaults

	tests := []struct {
		name           string
		defaultValue   float64
		userConfigVal  float64
		cliFlag        float64
		expectedResult float64
		description    string
	}{
		{
			name:           "default_only",
			defaultValue:   12.0,
			userConfigVal:  0,
			cliFlag:        0,
			expectedResult: 12.0,
			description:    "should use default when no override",
		},
		{
			name:           "user_config_overrides_default",
			defaultValue:   12.0,
			userConfigVal:  14.0,
			cliFlag:        0,
			expectedResult: 14.0,
			description:    "user config should override default",
		},
		{
			name:           "cli_overrides_user_config",
			defaultValue:   12.0,
			userConfigVal:  14.0,
			cliFlag:        16.0,
			expectedResult: 16.0,
			description:    "CLI flag should override user config",
		},
		{
			name:           "cli_overrides_default",
			defaultValue:   12.0,
			userConfigVal:  0,
			cliFlag:        18.0,
			expectedResult: 18.0,
			description:    "CLI flag should override default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start with default config
			baseConfig := core.DefaultConfig()

			// Simulate user config application
			userConfig := &config.UserConfig{
				FontSize: tt.userConfigVal,
			}
			config.ApplyUserConfig(baseConfig, userConfig)

			// Simulate CLI flag override (using Changed() pattern)
			// In real usage, Changed() detects if flag was explicitly set
			if tt.cliFlag > 0 {
				baseConfig.Renderer.FontSize = tt.cliFlag
			}

			if baseConfig.Renderer.FontSize != tt.expectedResult {
				t.Errorf("%s: got FontSize=%v, want %v",
					tt.description, baseConfig.Renderer.FontSize, tt.expectedResult)
			}
		})
	}
}

func TestConvertRequiresInput(t *testing.T) {
	// Test that convert command with no args fails
	cmd := newConvertCommand()

	// Execute with no arguments
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("convert command should fail with no arguments")
	}
}

func TestMultipleFilesWithOutputFlagReturnsError(t *testing.T) {
	// When multiple files are provided with a single output path,
	// the command should return an error to prevent silent overwrites

	cmd := newConvertCommand()
	cmd.SetArgs([]string{"file1.md", "file2.md", "-o", "output.pdf"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when using --output with multiple input files")
	}

	expectedMsg := "cannot use --output with multiple input files"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got: %v", expectedMsg, err)
	}
}

func TestMultipleFilesWithoutOutputFlag(t *testing.T) {
	// When multiple files are provided without -o flag,
	// each file should generate its own PDF

	tempDir := t.TempDir()

	// Create test markdown files
	file1 := filepath.Join(tempDir, "test1.md")
	file2 := filepath.Join(tempDir, "test2.md")

	if err := os.WriteFile(file1, []byte("# Test 1"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("# Test 2"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := core.ConversionOptions{
		InputFiles: []string{file1, file2},
		OutputPath: "", // Let each file generate its own output
		Verbose:    false,
	}

	cfg := core.DefaultConfig()
	cfg.Plugins.Enabled = false

	engine, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Change to temp dir for output
	originalWd, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Logf("warning: failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	err = engine.Convert(opts)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	// Verify both PDFs were created
	if _, err := os.Stat(filepath.Join(tempDir, "test1.pdf")); os.IsNotExist(err) {
		t.Error("test1.pdf was not created")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "test2.pdf")); os.IsNotExist(err) {
		t.Error("test2.pdf was not created")
	}
}

func TestZeroMarginCanBeSet(t *testing.T) {
	// Verify that 0 margin values are actually applied and not ignored
	// This is important because zero is a valid margin value for full-bleed printing

	tests := []struct {
		name         string
		marginTop    float64
		marginBottom float64
		marginLeft   float64
		marginRight  float64
	}{
		{
			name:         "all_zero_margins",
			marginTop:    0,
			marginBottom: 0,
			marginLeft:   0,
			marginRight:  0,
		},
		{
			name:         "mixed_zero_margins",
			marginTop:    0,
			marginBottom: 10,
			marginLeft:   0,
			marginRight:  15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := core.DefaultConfig()

			// Apply zero margins directly (simulating CLI flag behavior with Changed())
			cfg.Renderer.Margins.Top = tt.marginTop
			cfg.Renderer.Margins.Bottom = tt.marginBottom
			cfg.Renderer.Margins.Left = tt.marginLeft
			cfg.Renderer.Margins.Right = tt.marginRight

			// Verify the values were set
			if cfg.Renderer.Margins.Top != tt.marginTop {
				t.Errorf("margin top: got %v, want %v", cfg.Renderer.Margins.Top, tt.marginTop)
			}
			if cfg.Renderer.Margins.Bottom != tt.marginBottom {
				t.Errorf("margin bottom: got %v, want %v", cfg.Renderer.Margins.Bottom, tt.marginBottom)
			}
			if cfg.Renderer.Margins.Left != tt.marginLeft {
				t.Errorf("margin left: got %v, want %v", cfg.Renderer.Margins.Left, tt.marginLeft)
			}
			if cfg.Renderer.Margins.Right != tt.marginRight {
				t.Errorf("margin right: got %v, want %v", cfg.Renderer.Margins.Right, tt.marginRight)
			}

			// Validate config allows zero margins
			err := core.ValidateConfig(cfg)
			if err != nil {
				t.Errorf("config with zero margins should be valid: %v", err)
			}
		})
	}
}

func TestZeroMarginViaCLIFlag(t *testing.T) {
	// Test that zero margins can be set via CLI flags using Changed() detection
	cmd := newConvertCommand()

	// Simulate setting margin-top=0 via CLI
	if err := cmd.Flags().Set("margin-top", "0"); err != nil {
		t.Fatalf("failed to set flag: %v", err)
	}

	// Verify Changed() returns true for explicitly set flag
	if !cmd.Flags().Changed("margin-top") {
		t.Error("Changed() should return true for explicitly set margin-top=0")
	}

	// Verify Changed() returns false for unset flags
	if cmd.Flags().Changed("margin-bottom") {
		t.Error("Changed() should return false for unset margin-bottom")
	}
}

func TestVerboseOutput(t *testing.T) {
	// Test that verbose flag produces expected output
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Hello World"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "test.pdf")

	cfg := core.DefaultConfig()
	cfg.Plugins.Enabled = false

	engine, err := core.NewEngine(cfg)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := core.ConversionOptions{
		InputFiles: []string{testFile},
		OutputPath: outputFile,
		Verbose:    true,
	}

	err = engine.Convert(opts)
	if err != nil {
		if closeErr := w.Close(); closeErr != nil {
			t.Logf("warning: failed to close pipe writer: %v", closeErr)
		}
		os.Stdout = oldStdout
		t.Fatalf("conversion failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Logf("warning: failed to close pipe writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("warning: failed to read from pipe: %v", err)
	}
	os.Stdout = oldStdout

	output := buf.String()
	if !strings.Contains(output, "Converted:") {
		t.Errorf("verbose output should contain 'Converted:', got: %s", output)
	}
}

func TestConvertCmdArgsValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no_args",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "one_arg",
			args:      []string{"file.md"},
			wantError: false,
		},
		{
			name:      "multiple_args",
			args:      []string{"file1.md", "file2.md"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newConvertCommand()
			err := cmd.Args(cmd, tt.args)
			if (err != nil) != tt.wantError {
				t.Errorf("Args() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestApplyOverridesWithChanged(t *testing.T) {
	// Test that applyOverrides correctly uses Changed() for all numeric flags
	c := &convertCommand{
		marginTop:    0,
		marginBottom: 0,
		fontSize:     0,
		mermaidScale: 0,
	}

	cmd := newConvertCommand()
	cfg := core.DefaultConfig()

	// Set some flags explicitly (including zero values)
	if err := cmd.Flags().Set("margin-top", "0"); err != nil {
		t.Fatalf("failed to set margin-top: %v", err)
	}
	if err := cmd.Flags().Set("font-size", "0"); err != nil {
		t.Fatalf("failed to set font-size: %v", err)
	}

	// Store original values
	originalMarginBottom := cfg.Renderer.Margins.Bottom
	originalMermaidScale := cfg.Renderer.Mermaid.Scale

	// Apply overrides
	c.applyOverrides(cmd, cfg)

	// margin-top was explicitly set to 0, should be 0
	if cfg.Renderer.Margins.Top != 0 {
		t.Errorf("margin-top should be 0, got %v", cfg.Renderer.Margins.Top)
	}

	// font-size was explicitly set to 0, should be 0
	if cfg.Renderer.FontSize != 0 {
		t.Errorf("font-size should be 0, got %v", cfg.Renderer.FontSize)
	}

	// margin-bottom was NOT set, should retain default
	if cfg.Renderer.Margins.Bottom != originalMarginBottom {
		t.Errorf("margin-bottom should retain default %v, got %v",
			originalMarginBottom, cfg.Renderer.Margins.Bottom)
	}

	// mermaid-scale was NOT set, should retain default
	if cfg.Renderer.Mermaid.Scale != originalMermaidScale {
		t.Errorf("mermaid-scale should retain default %v, got %v",
			originalMermaidScale, cfg.Renderer.Mermaid.Scale)
	}
}
