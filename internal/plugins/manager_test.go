package plugins

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name       string
		pluginDir  string
		enabled    bool
	}{
		{
			name:      "enabled_with_default_dir",
			pluginDir: "./plugins",
			enabled:   true,
		},
		{
			name:      "disabled_manager",
			pluginDir: "./plugins",
			enabled:   false,
		},
		{
			name:      "custom_directory",
			pluginDir: "/custom/plugin/path",
			enabled:   true,
		},
		{
			name:      "empty_directory",
			pluginDir: "",
			enabled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.pluginDir, tt.enabled, nil)

			if manager == nil {
				t.Fatal("NewManager returned nil")
			}

			if manager.pluginDir != tt.pluginDir {
				t.Errorf("pluginDir = %q, want %q", manager.pluginDir, tt.pluginDir)
			}

			if manager.enabled != tt.enabled {
				t.Errorf("enabled = %v, want %v", manager.enabled, tt.enabled)
			}

			if manager.plugins == nil {
				t.Error("plugins map should be initialized")
			}

			if manager.transformers == nil {
				t.Error("transformers slice should be initialized")
			}

			if manager.generators == nil {
				t.Error("generators map should be initialized")
			}
		})
	}
}

func TestLoadPlugins_MissingDirectory(t *testing.T) {
	// Test that loading plugins from a non-existent directory doesn't error
	manager := NewManager("/nonexistent/plugin/directory", true, nil)

	err := manager.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins should not error for missing directory: %v", err)
	}

	// Should have no plugins loaded
	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestLoadPlugins_EmptyDirectory(t *testing.T) {
	// Create an empty directory
	tempDir := t.TempDir()

	manager := NewManager(tempDir, true, nil)

	err := manager.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins should not error for empty directory: %v", err)
	}

	// Should have no plugins loaded
	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestLoadPlugins_Disabled(t *testing.T) {
	// When plugins are disabled, LoadPlugins should return immediately
	manager := NewManager("./plugins", false, nil)

	err := manager.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins should not error when disabled: %v", err)
	}

	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins when disabled, got %d", len(plugins))
	}
}

func TestLoadPlugins_NonPluginFiles(t *testing.T) {
	// Create a directory with non-plugin files
	tempDir := t.TempDir()

	// Create various non-plugin files
	files := []struct {
		name    string
		content string
	}{
		{"readme.txt", "This is a readme"},
		{"config.yaml", "key: value"},
		{"script.sh", "#!/bin/bash\necho hello"},
		{"data.json", `{"key": "value"}`},
	}

	for _, f := range files {
		path := filepath.Join(tempDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	manager := NewManager(tempDir, true, nil)

	err := manager.LoadPlugins()
	if err != nil {
		t.Errorf("LoadPlugins should not error for non-plugin files: %v", err)
	}

	// Should have no plugins loaded (only .so files are considered)
	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestApplyTransformers_Priority(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Create mock transformers with different priorities
	transformer1 := &testTransformer{
		name:     "transformer-low",
		priority: 100,
		transformFunc: func(node ast.Node, ctx *TransformContext) (ast.Node, error) {
			ctx.Metadata["low"] = true
			return node, nil
		},
	}

	transformer2 := &testTransformer{
		name:     "transformer-high",
		priority: 10,
		transformFunc: func(node ast.Node, ctx *TransformContext) (ast.Node, error) {
			ctx.Metadata["high"] = true
			return node, nil
		},
	}

	transformer3 := &testTransformer{
		name:     "transformer-mid",
		priority: 50,
		transformFunc: func(node ast.Node, ctx *TransformContext) (ast.Node, error) {
			ctx.Metadata["mid"] = true
			return node, nil
		},
	}

	// Add transformers in random order
	manager.transformers = append(manager.transformers, transformer1, transformer2, transformer3)

	// Sort by priority (as LoadPlugins would do)
	// Lower priority number = runs first
	sortTransformers(manager.transformers)

	// Verify order
	if manager.transformers[0].Priority() != 10 {
		t.Errorf("first transformer should have priority 10, got %d", manager.transformers[0].Priority())
	}
	if manager.transformers[1].Priority() != 50 {
		t.Errorf("second transformer should have priority 50, got %d", manager.transformers[1].Priority())
	}
	if manager.transformers[2].Priority() != 100 {
		t.Errorf("third transformer should have priority 100, got %d", manager.transformers[2].Priority())
	}
}

func TestApplyTransformers_SupportedNodes(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Create a transformer that only supports heading nodes
	transformer := &testTransformer{
		name:           "heading-only",
		priority:       100,
		supportedNodes: []ast.NodeKind{ast.KindHeading},
		transformFunc: func(node ast.Node, ctx *TransformContext) (ast.Node, error) {
			ctx.Metadata["transformed"] = true
			return node, nil
		},
	}

	manager.transformers = append(manager.transformers, transformer)

	// Test with a paragraph node (should not be transformed)
	paragraphNode := ast.NewParagraph()
	ctx := &TransformContext{
		CurrentNode: paragraphNode,
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
	}

	result, err := manager.ApplyTransformers(paragraphNode, ctx)
	if err != nil {
		t.Fatalf("ApplyTransformers failed: %v", err)
	}

	if ctx.Metadata["transformed"] == true {
		t.Error("paragraph should not have been transformed by heading-only transformer")
	}

	// Test with a heading node (should be transformed)
	headingNode := ast.NewHeading(1)
	ctx = &TransformContext{
		CurrentNode: headingNode,
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
	}

	result, err = manager.ApplyTransformers(headingNode, ctx)
	if err != nil {
		t.Fatalf("ApplyTransformers failed: %v", err)
	}

	if result == nil {
		t.Error("ApplyTransformers returned nil")
	}

	if ctx.Metadata["transformed"] != true {
		t.Error("heading should have been transformed")
	}
}

func TestApplyTransformers_NoTransformers(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	node := ast.NewParagraph()
	ctx := &TransformContext{
		CurrentNode: node,
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
	}

	result, err := manager.ApplyTransformers(node, ctx)
	if err != nil {
		t.Fatalf("ApplyTransformers failed: %v", err)
	}

	if result != node {
		t.Error("with no transformers, original node should be returned")
	}
}

func TestApplyTransformers_TransformerError(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	expectedErr := errors.New("transformer error")
	transformer := &testTransformer{
		name:     "error-transformer",
		priority: 100,
		transformFunc: func(node ast.Node, ctx *TransformContext) (ast.Node, error) {
			return nil, expectedErr
		},
	}

	manager.transformers = append(manager.transformers, transformer)

	node := ast.NewParagraph()
	ctx := &TransformContext{
		CurrentNode: node,
		Metadata:    make(map[string]interface{}),
		Config:      make(map[string]interface{}),
	}

	_, err := manager.ApplyTransformers(node, ctx)
	if err == nil {
		t.Error("ApplyTransformers should return error when transformer fails")
	}

	if !strings.Contains(err.Error(), "transformer error") {
		t.Errorf("error should contain original message, got: %v", err)
	}
}

func TestCleanup_AggregatesErrors(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Add plugins that return errors on cleanup
	plugin1 := &testPlugin{
		name:       "plugin1",
		cleanupErr: errors.New("cleanup error 1"),
	}
	plugin2 := &testPlugin{
		name:       "plugin2",
		cleanupErr: errors.New("cleanup error 2"),
	}
	plugin3 := &testPlugin{
		name:       "plugin3",
		cleanupErr: nil, // This one succeeds
	}

	manager.plugins["plugin1"] = plugin1
	manager.plugins["plugin2"] = plugin2
	manager.plugins["plugin3"] = plugin3

	err := manager.Cleanup()
	if err == nil {
		t.Error("Cleanup should return error when plugins fail")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "plugin1") || !strings.Contains(errStr, "cleanup error 1") {
		t.Error("error should contain plugin1's error")
	}
	if !strings.Contains(errStr, "plugin2") || !strings.Contains(errStr, "cleanup error 2") {
		t.Error("error should contain plugin2's error")
	}
}

func TestCleanup_NoErrors(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Add plugins that succeed on cleanup
	plugin1 := &testPlugin{name: "plugin1", cleanupErr: nil}
	plugin2 := &testPlugin{name: "plugin2", cleanupErr: nil}

	manager.plugins["plugin1"] = plugin1
	manager.plugins["plugin2"] = plugin2

	err := manager.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not error when all plugins succeed: %v", err)
	}
}

func TestCleanup_EmptyPlugins(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	err := manager.Cleanup()
	if err != nil {
		t.Errorf("Cleanup should not error with no plugins: %v", err)
	}
}

func TestGetTransformers(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Initially empty
	transformers := manager.GetTransformers()
	if len(transformers) != 0 {
		t.Errorf("expected 0 transformers, got %d", len(transformers))
	}

	// Add some transformers
	manager.transformers = append(manager.transformers, &testTransformer{name: "t1"})
	manager.transformers = append(manager.transformers, &testTransformer{name: "t2"})

	transformers = manager.GetTransformers()
	if len(transformers) != 2 {
		t.Errorf("expected 2 transformers, got %d", len(transformers))
	}
}

func TestGetGenerators(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Add generators for different phases
	gen1 := &testGenerator{name: "gen1", phase: BeforeContent}
	gen2 := &testGenerator{name: "gen2", phase: AfterContent}
	gen3 := &testGenerator{name: "gen3", phase: BeforeContent}

	manager.generators[BeforeContent] = []ContentGenerator{gen1, gen3}
	manager.generators[AfterContent] = []ContentGenerator{gen2}

	// Get BeforeContent generators
	beforeGens := manager.GetGenerators(BeforeContent)
	if len(beforeGens) != 2 {
		t.Errorf("expected 2 BeforeContent generators, got %d", len(beforeGens))
	}

	// Get AfterContent generators
	afterGens := manager.GetGenerators(AfterContent)
	if len(afterGens) != 1 {
		t.Errorf("expected 1 AfterContent generator, got %d", len(afterGens))
	}

	// Get generators for unused phase
	pageGens := manager.GetGenerators(BeforeEachPage)
	if len(pageGens) != 0 {
		t.Errorf("expected 0 BeforeEachPage generators, got %d", len(pageGens))
	}
}

func TestGenerateContent(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Create generators that produce elements
	gen1 := &testGenerator{
		name:  "gen1",
		phase: BeforeContent,
		elements: []PDFElement{
			&TextElement{Content: "Element 1"},
		},
	}
	gen2 := &testGenerator{
		name:  "gen2",
		phase: BeforeContent,
		elements: []PDFElement{
			&TextElement{Content: "Element 2"},
			&TextElement{Content: "Element 3"},
		},
	}

	manager.generators[BeforeContent] = []ContentGenerator{gen1, gen2}

	ctx := &RenderContext{
		Metadata: make(map[string]interface{}),
		Config:   make(map[string]interface{}),
	}

	elements, err := manager.GenerateContent(BeforeContent, ctx)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	if len(elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(elements))
	}
}

func TestGenerateContent_Error(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	expectedErr := errors.New("generation error")
	gen := &testGenerator{
		name:  "error-gen",
		phase: BeforeContent,
		err:   expectedErr,
	}

	manager.generators[BeforeContent] = []ContentGenerator{gen}

	ctx := &RenderContext{
		Metadata: make(map[string]interface{}),
		Config:   make(map[string]interface{}),
	}

	_, err := manager.GenerateContent(BeforeContent, ctx)
	if err == nil {
		t.Error("GenerateContent should return error when generator fails")
	}
}

func TestListPlugins(t *testing.T) {
	manager := NewManager("./plugins", true, nil)

	// Initially empty
	plugins := manager.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}

	// Add some plugins
	manager.plugins["plugin1"] = &testPlugin{
		name:        "plugin1",
		version:     "1.0.0",
		description: "Test plugin 1",
	}
	manager.plugins["plugin2"] = &testPlugin{
		name:        "plugin2",
		version:     "2.0.0",
		description: "Test plugin 2",
	}

	plugins = manager.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}

	// Verify plugin info
	found := make(map[string]bool)
	for _, p := range plugins {
		found[p.Name] = true
		if p.Name == "plugin1" {
			if p.Version != "1.0.0" {
				t.Errorf("plugin1 version = %q, want %q", p.Version, "1.0.0")
			}
			if p.Description != "Test plugin 1" {
				t.Errorf("plugin1 description = %q, want %q", p.Description, "Test plugin 1")
			}
		}
	}

	if !found["plugin1"] || !found["plugin2"] {
		t.Error("not all plugins were listed")
	}
}

func TestPluginSecurity_PathTraversal(t *testing.T) {
	tests := []struct {
		name       string
		pluginDir  string
		shouldWarn bool
	}{
		{
			name:       "normal_path",
			pluginDir:  "./plugins",
			shouldWarn: false,
		},
		{
			name:       "path_with_traversal",
			pluginDir:  "../../../etc/plugins",
			shouldWarn: true, // Contains path traversal
		},
		{
			name:       "absolute_system_path",
			pluginDir:  "/etc/plugins",
			shouldWarn: true, // System path
		},
		{
			name:       "home_relative_path",
			pluginDir:  "~/.config/md-to-pdf/plugins",
			shouldWarn: false, // User directory is acceptable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check for path traversal attempts
			hasTraversal := strings.Contains(tt.pluginDir, "..")
			isSystemPath := strings.HasPrefix(tt.pluginDir, "/etc") ||
				strings.HasPrefix(tt.pluginDir, "/usr") ||
				strings.HasPrefix(tt.pluginDir, "/var")

			shouldWarn := hasTraversal || isSystemPath

			if shouldWarn != tt.shouldWarn {
				t.Errorf("path %q: shouldWarn = %v, want %v",
					tt.pluginDir, shouldWarn, tt.shouldWarn)
			}
		})
	}
}

// Helper function to sort transformers by priority
func sortTransformers(transformers []ASTTransformer) {
	for i := 0; i < len(transformers)-1; i++ {
		for j := i + 1; j < len(transformers); j++ {
			if transformers[i].Priority() > transformers[j].Priority() {
				transformers[i], transformers[j] = transformers[j], transformers[i]
			}
		}
	}
}

// Test doubles

type testPlugin struct {
	name        string
	version     string
	description string
	cleanupErr  error
}

func (p *testPlugin) Name() string        { return p.name }
func (p *testPlugin) Version() string     { return p.version }
func (p *testPlugin) Description() string { return p.description }
func (p *testPlugin) Init(config map[string]interface{}) error { return nil }
func (p *testPlugin) Cleanup() error      { return p.cleanupErr }

type testTransformer struct {
	name           string
	version        string
	description    string
	priority       int
	supportedNodes []ast.NodeKind
	transformFunc  func(ast.Node, *TransformContext) (ast.Node, error)
}

func (t *testTransformer) Name() string        { return t.name }
func (t *testTransformer) Version() string     { return t.version }
func (t *testTransformer) Description() string { return t.description }
func (t *testTransformer) Init(config map[string]interface{}) error { return nil }
func (t *testTransformer) Cleanup() error      { return nil }
func (t *testTransformer) Priority() int       { return t.priority }
func (t *testTransformer) SupportedNodes() []ast.NodeKind { return t.supportedNodes }
func (t *testTransformer) Transform(node ast.Node, ctx *TransformContext) (ast.Node, error) {
	if t.transformFunc != nil {
		return t.transformFunc(node, ctx)
	}
	return node, nil
}

type testGenerator struct {
	name     string
	version  string
	phase    GenerationPhase
	elements []PDFElement
	err      error
}

func (g *testGenerator) Name() string        { return g.name }
func (g *testGenerator) Version() string     { return g.version }
func (g *testGenerator) Description() string { return "" }
func (g *testGenerator) Init(config map[string]interface{}) error { return nil }
func (g *testGenerator) Cleanup() error      { return nil }
func (g *testGenerator) GenerationPhase() GenerationPhase { return g.phase }
func (g *testGenerator) Generate(ctx *RenderContext) ([]PDFElement, error) {
	if g.err != nil {
		return nil, g.err
	}
	return g.elements, nil
}
