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

type Manager struct {
	plugins      map[string]Plugin
	transformers []ASTTransformer
	generators   map[GenerationPhase][]ContentGenerator
	pluginDir    string
	enabled      bool
}

func NewManager(pluginDir string, enabled bool) *Manager {
	return &Manager{
		plugins:      make(map[string]Plugin),
		transformers: make([]ASTTransformer, 0),
		generators:   make(map[GenerationPhase][]ContentGenerator),
		pluginDir:    pluginDir,
		enabled:      enabled,
	}
}

func (m *Manager) LoadPlugins() error {
	if !m.enabled {
		return nil
	}

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
		err := m.loadPlugin(pluginPath)
		if err != nil {
			fmt.Printf("Warning: failed to load plugin %s: %v\n", file.Name(), err)
			continue
		}
	}

	// Sort transformers by priority
	sort.Slice(m.transformers, func(i, j int) bool {
		return m.transformers[i].Priority() < m.transformers[j].Priority()
	})

	return nil
}

func (m *Manager) loadPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for NewPlugin function
	newPluginSymbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin missing NewPlugin function: %w", err)
	}

	newPluginFunc, ok := newPluginSymbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin has invalid signature")
	}

	pluginInstance := newPluginFunc()
	
	// Initialize plugin
	err = pluginInstance.Init(make(map[string]interface{}))
	if err != nil {
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

	return nil
}

func (m *Manager) GetTransformers() []ASTTransformer {
	return m.transformers
}

func (m *Manager) GetGenerators(phase GenerationPhase) []ContentGenerator {
	return m.generators[phase]
}

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

func (m *Manager) ListPlugins() []PluginInfo {
	var plugins []PluginInfo
	
	for _, p := range m.plugins {
		plugins = append(plugins, PluginInfo{
			Name:        p.Name(),
			Version:     p.Version(),
			Description: p.Description(),
		})
	}
	
	return plugins
}