package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Manager handles plugin lifecycle and coordination
type Manager struct {
	plugins        map[string]Plugin
	transformers   []ASTTransformer
	generators     map[GenerationPhase][]ContentGenerator
	pluginDir      string
	enabled        bool
	pluginConfigs  map[string]map[string]interface{}
	securityConfig *SecurityConfig
	allowlist      *PluginAllowlist
	logger         *PluginSecurityLogger
}

// NewManager creates a new plugin manager with the specified directory and enabled state.
// pluginConfigs is an optional map of plugin-specific configurations keyed by plugin name.
func NewManager(pluginDir string, enabled bool, pluginConfigs map[string]map[string]interface{}) *Manager {
	if pluginConfigs == nil {
		pluginConfigs = make(map[string]map[string]interface{})
	}
	return &Manager{
		plugins:        make(map[string]Plugin),
		transformers:   make([]ASTTransformer, 0),
		generators:     make(map[GenerationPhase][]ContentGenerator),
		pluginDir:      pluginDir,
		enabled:        enabled,
		pluginConfigs:  pluginConfigs,
		securityConfig: DefaultSecurityConfig(),
		logger:         NewPluginSecurityLogger(),
	}
}

// NewManagerWithSecurity creates a new plugin manager with explicit security configuration
func NewManagerWithSecurity(pluginDir string, enabled bool, securityConfig *SecurityConfig) (*Manager, error) {
	if securityConfig == nil {
		securityConfig = DefaultSecurityConfig()
	}

	// Load allowlist if path is specified
	var allowlist *PluginAllowlist
	var err error
	if securityConfig.AllowlistPath != "" {
		allowlist, err = LoadAllowlistFromFile(securityConfig.AllowlistPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load plugin allowlist: %w", err)
		}
	} else {
		allowlist = NewPluginAllowlist()
	}

	return &Manager{
		plugins:        make(map[string]Plugin),
		transformers:   make([]ASTTransformer, 0),
		generators:     make(map[GenerationPhase][]ContentGenerator),
		pluginDir:      pluginDir,
		enabled:        enabled,
		securityConfig: securityConfig,
		allowlist:      allowlist,
		logger:         NewPluginSecurityLogger(),
	}, nil
}

// SetSecurityConfig updates the security configuration for the manager
func (m *Manager) SetSecurityConfig(config *SecurityConfig) error {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	m.securityConfig = config

	// Reload allowlist if path changed
	if config.AllowlistPath != "" {
		allowlist, err := LoadAllowlistFromFile(config.AllowlistPath)
		if err != nil {
			return fmt.Errorf("failed to load plugin allowlist: %w", err)
		}
		m.allowlist = allowlist
	} else {
		m.allowlist = NewPluginAllowlist()
	}

	return nil
}

// GetSecurityEvents returns all logged security events
func (m *Manager) GetSecurityEvents() []PluginLoadEvent {
	if m.logger == nil {
		return nil
	}
	return m.logger.GetEvents()
}

// LoadPlugins discovers and loads all plugins from the configured directory
func (m *Manager) LoadPlugins() error {
	if !m.enabled {
		return nil
	}

	// Validate and canonicalize the plugin directory path
	validatedPath, err := m.validatePluginDirectory()
	if err != nil {
		return err
	}

	// Update the plugin directory to the validated path
	m.pluginDir = validatedPath

	if _, err := os.Stat(m.pluginDir); os.IsNotExist(err) {
		// Plugin directory doesn't exist, skip loading
		return nil
	}

	files, err := os.ReadDir(m.pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".so") {
			continue
		}

		pluginPath := filepath.Join(m.pluginDir, file.Name())
		loadErr := m.loadPlugin(pluginPath)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load plugin %s: %v\n", file.Name(), loadErr)
			continue
		}
	}

	// Sort transformers by priority
	sort.Slice(m.transformers, func(i, j int) bool {
		return m.transformers[i].Priority() < m.transformers[j].Priority()
	})

	return nil
}

// validatePluginDirectory validates the plugin directory for security issues
func (m *Manager) validatePluginDirectory() (string, error) {
	// Validate path for traversal attacks
	validatedPath, err := ValidatePluginPath(m.pluginDir)
	if err != nil {
		return "", fmt.Errorf("invalid plugin directory: %w", err)
	}

	// Check if directory is within working directory or trusted paths
	inWorkDir, checkErr := IsPathInWorkingDirectory(validatedPath)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "[SECURITY WARNING] Could not verify plugin directory safety: %v\n", checkErr)
	} else if !inWorkDir {
		// Check if in trusted directories
		inTrusted := IsPathInTrustedDirectory(validatedPath, m.securityConfig.TrustedDirectories)
		if !inTrusted {
			fmt.Fprintf(os.Stderr, "[SECURITY WARNING] Plugin directory '%s' is outside current working directory and trusted paths\n", validatedPath)
		}
	}

	return validatedPath, nil
}

// loadPlugin loads a single plugin with security verification
func (m *Manager) loadPlugin(path string) error {
	// Perform security verification before loading
	event, verifyErr := VerifyPlugin(path, m.securityConfig, m.allowlist)

	// Log the attempt regardless of outcome
	if m.logger != nil && event != nil {
		defer func() {
			m.logger.LogLoadAttempt(*event)
		}()
	}

	if verifyErr != nil {
		if event != nil {
			event.Success = false
		}
		return verifyErr
	}

	// Actually load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		if event != nil {
			event.Success = false
			event.Error = fmt.Sprintf("failed to open plugin: %v", err)
		}
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for NewPlugin function
	newPluginSymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		if event != nil {
			event.Success = false
			event.Error = fmt.Sprintf("plugin missing NewPlugin function: %v", err)
		}
		return fmt.Errorf("plugin missing NewPlugin function: %w", err)
	}

	newPluginFunc, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		if event != nil {
			event.Success = false
			event.Error = "NewPlugin has invalid signature"
		}
		return fmt.Errorf("NewPlugin has invalid signature")
	}

	pluginInstance := newPluginFunc()

	// Update event with actual plugin name
	if event != nil {
		event.PluginName = pluginInstance.Name()
	}

	// Get plugin-specific configuration, or use empty map if none provided
	pluginConfig := m.pluginConfigs[pluginInstance.Name()]
	if pluginConfig == nil {
		pluginConfig = make(map[string]interface{})
	}

	// Initialize plugin with its configuration
	err = pluginInstance.Init(pluginConfig)
	if err != nil {
		if event != nil {
			event.Success = false
			event.Error = fmt.Sprintf("failed to initialize plugin: %v", err)
		}
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Register plugin
	m.plugins[pluginInstance.Name()] = pluginInstance

	// Check for additional capabilities
	if transformer, ok := pluginInstance.(ASTTransformer); ok {
		m.transformers = append(m.transformers, transformer)
	}

	if generator, ok := pluginInstance.(ContentGenerator); ok {
		phase := generator.GenerationPhase()
		if m.generators[phase] == nil {
			m.generators[phase] = make([]ContentGenerator, 0)
		}
		m.generators[phase] = append(m.generators[phase], generator)
	}

	// Mark as successful
	if event != nil {
		event.Success = true
	}

	return nil
}

// GetTransformers returns all registered AST transformers
func (m *Manager) GetTransformers() []ASTTransformer {
	return m.transformers
}

// GetGenerators returns all content generators for a specific phase
func (m *Manager) GetGenerators(phase GenerationPhase) []ContentGenerator {
	return m.generators[phase]
}

// ApplyTransformers applies all registered transformers to a node
func (m *Manager) ApplyTransformers(node ast.Node, ctx *TransformContext) (ast.Node, error) {
	result := node

	for _, transformer := range m.transformers {
		// Check if transformer supports this node type
		supportedNodes := transformer.SupportedNodes()
		if len(supportedNodes) > 0 {
			supported := false
			for _, nodeKind := range supportedNodes {
				if result.Kind() == nodeKind {
					supported = true
					break
				}
			}
			if !supported {
				continue
			}
		}

		transformedNode, err := transformer.Transform(result, ctx)
		if err != nil {
			return result, fmt.Errorf("transformer %s failed: %w", transformer.Name(), err)
		}

		result = transformedNode
	}

	return result, nil
}

// GenerateContent runs all content generators for a specific phase
func (m *Manager) GenerateContent(phase GenerationPhase, ctx *RenderContext) ([]PDFElement, error) {
	var elements []PDFElement

	generators := m.GetGenerators(phase)
	for _, generator := range generators {
		generatedElements, err := generator.Generate(ctx)
		if err != nil {
			return elements, fmt.Errorf("generator %s failed: %w", generator.Name(), err)
		}

		elements = append(elements, generatedElements...)
	}

	return elements, nil
}

// Cleanup performs cleanup for all loaded plugins
func (m *Manager) Cleanup() error {
	var errors []string

	for name, p := range m.plugins {
		if err := p.Cleanup(); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("plugin cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ListPlugins returns information about all loaded plugins
func (m *Manager) ListPlugins() []PluginInfo {
	var pluginList []PluginInfo

	for _, p := range m.plugins {
		pluginList = append(pluginList, PluginInfo{
			Name:        p.Name(),
			Version:     p.Version(),
			Description: p.Description(),
		})
	}

	return pluginList
}
