package plugin

import (
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (ast.Node, []byte)
		expected string
	}{
		{
			name: "extract_from_text_node",
			setup: func() (ast.Node, []byte) {
				source := []byte("Hello World")
				seg := text.NewSegment(0, len(source))
				textNode := ast.NewTextSegment(seg)
				return textNode, source
			},
			expected: "Hello World",
		},
		{
			name: "extract_from_paragraph_with_text",
			setup: func() (ast.Node, []byte) {
				source := []byte("Test paragraph content")
				paragraph := ast.NewParagraph()
				seg := text.NewSegment(0, len(source))
				textNode := ast.NewTextSegment(seg)
				paragraph.AppendChild(paragraph, textNode)
				return paragraph, source
			},
			expected: "Test paragraph content",
		},
		{
			name: "extract_from_heading_with_text",
			setup: func() (ast.Node, []byte) {
				source := []byte("Heading Text")
				heading := ast.NewHeading(1)
				seg := text.NewSegment(0, len(source))
				textNode := ast.NewTextSegment(seg)
				heading.AppendChild(heading, textNode)
				return heading, source
			},
			expected: "Heading Text",
		},
		{
			name: "extract_from_empty_paragraph",
			setup: func() (ast.Node, []byte) {
				source := []byte("")
				paragraph := ast.NewParagraph()
				return paragraph, source
			},
			expected: "",
		},
		{
			name: "extract_from_nested_nodes",
			setup: func() (ast.Node, []byte) {
				source := []byte("First Second")
				paragraph := ast.NewParagraph()

				// Add first text
				seg1 := text.NewSegment(0, 5) // "First"
				text1 := ast.NewTextSegment(seg1)
				paragraph.AppendChild(paragraph, text1)

				// Add second text
				seg2 := text.NewSegment(6, 12) // "Second"
				text2 := ast.NewTextSegment(seg2)
				paragraph.AppendChild(paragraph, text2)

				return paragraph, source
			},
			expected: "FirstSecond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, source := tt.setup()
			result := ExtractText(node, source)

			if result != tt.expected {
				t.Errorf("ExtractText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWalkChildNodes(t *testing.T) {
	t.Run("walk_paragraph_children", func(t *testing.T) {
		source := []byte("Text content here")
		paragraph := ast.NewParagraph()

		// Add multiple text children
		for i := 0; i < 3; i++ {
			textNode := ast.NewTextSegment(text.NewSegment(0, len(source)))
			paragraph.AppendChild(paragraph, textNode)
		}

		var visitedCount int
		WalkChildNodes(paragraph, func(node ast.Node) bool {
			visitedCount++
			return true // continue walking
		})

		if visitedCount != 3 {
			t.Errorf("visited %d nodes, want 3", visitedCount)
		}
	})

	t.Run("walk_with_early_stop", func(t *testing.T) {
		source := []byte("Text")
		paragraph := ast.NewParagraph()

		for i := 0; i < 5; i++ {
			textNode := ast.NewTextSegment(text.NewSegment(0, len(source)))
			paragraph.AppendChild(paragraph, textNode)
		}

		var visitedCount int
		WalkChildNodes(paragraph, func(node ast.Node) bool {
			visitedCount++
			return visitedCount < 2 // stop after 2
		})

		if visitedCount != 2 {
			t.Errorf("visited %d nodes, want 2 (early stop)", visitedCount)
		}
	})

	t.Run("walk_empty_node", func(t *testing.T) {
		paragraph := ast.NewParagraph()

		var visitedCount int
		WalkChildNodes(paragraph, func(node ast.Node) bool {
			visitedCount++
			return true
		})

		if visitedCount != 0 {
			t.Errorf("visited %d nodes, want 0 for empty node", visitedCount)
		}
	})

	t.Run("walk_nested_children", func(t *testing.T) {
		// Create a document with nested structure
		doc := ast.NewDocument()
		paragraph := ast.NewParagraph()
		source := []byte("Text")
		textNode := ast.NewTextSegment(text.NewSegment(0, len(source)))

		paragraph.AppendChild(paragraph, textNode)
		doc.AppendChild(doc, paragraph)

		var visitedNodes []ast.NodeKind
		WalkChildNodes(doc, func(node ast.Node) bool {
			visitedNodes = append(visitedNodes, node.Kind())
			return true
		})

		// Should visit paragraph and text
		if len(visitedNodes) != 2 {
			t.Errorf("visited %d nodes, want 2", len(visitedNodes))
		}
	})
}

func TestCreateHeading(t *testing.T) {
	tests := []struct {
		name  string
		level int
	}{
		{"h1", 1},
		{"h2", 2},
		{"h3", 3},
		{"h4", 4},
		{"h5", 5},
		{"h6", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heading := ast.NewHeading(tt.level)

			if heading == nil {
				t.Fatal("NewHeading returned nil")
			}

			if heading.Kind() != ast.KindHeading {
				t.Errorf("heading kind = %v, want %v", heading.Kind(), ast.KindHeading)
			}

			if heading.Level != tt.level {
				t.Errorf("heading level = %d, want %d", heading.Level, tt.level)
			}
		})
	}
}

func TestBasePlugin(t *testing.T) {
	t.Run("create_base_plugin", func(t *testing.T) {
		plugin := NewBasePlugin("test-plugin", "1.0.0", "A test plugin")

		if plugin == nil {
			t.Fatal("NewBasePlugin returned nil")
		}

		if plugin.Name() != "test-plugin" {
			t.Errorf("Name() = %q, want %q", plugin.Name(), "test-plugin")
		}

		if plugin.Version() != "1.0.0" {
			t.Errorf("Version() = %q, want %q", plugin.Version(), "1.0.0")
		}

		if plugin.Description() != "A test plugin" {
			t.Errorf("Description() = %q, want %q", plugin.Description(), "A test plugin")
		}
	})

	t.Run("init_returns_nil", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0", "test")

		err := plugin.Init(nil)
		if err != nil {
			t.Errorf("Init() should return nil, got: %v", err)
		}

		err = plugin.Init(map[string]interface{}{"key": "value"})
		if err != nil {
			t.Errorf("Init() with config should return nil, got: %v", err)
		}
	})

	t.Run("cleanup_returns_nil", func(t *testing.T) {
		plugin := NewBasePlugin("test", "1.0", "test")

		err := plugin.Cleanup()
		if err != nil {
			t.Errorf("Cleanup() should return nil, got: %v", err)
		}
	})

	t.Run("empty_values", func(t *testing.T) {
		plugin := NewBasePlugin("", "", "")

		if plugin.Name() != "" {
			t.Errorf("Name() = %q, want empty string", plugin.Name())
		}

		if plugin.Version() != "" {
			t.Errorf("Version() = %q, want empty string", plugin.Version())
		}

		if plugin.Description() != "" {
			t.Errorf("Description() = %q, want empty string", plugin.Description())
		}
	})
}

func TestIsCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (ast.Node, []byte)
		language string
		expected bool
	}{
		{
			name: "fenced_code_block_go",
			setup: func() (ast.Node, []byte) {
				source := []byte("```go\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				// Set the info segment for language
				infoSeg := text.NewSegment(3, 5) // "go" position
				cb.Info = &ast.Text{}
				cb.Info.Segment = infoSeg
				return cb, source
			},
			language: "go",
			expected: true,
		},
		{
			name: "fenced_code_block_wrong_language",
			setup: func() (ast.Node, []byte) {
				source := []byte("```python\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				infoSeg := text.NewSegment(3, 9) // "python" position
				cb.Info = &ast.Text{}
				cb.Info.Segment = infoSeg
				return cb, source
			},
			language: "go",
			expected: false,
		},
		{
			name: "fenced_code_block_no_language",
			setup: func() (ast.Node, []byte) {
				source := []byte("```\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				return cb, source
			},
			language: "",
			expected: true,
		},
		{
			name: "paragraph_not_code_block",
			setup: func() (ast.Node, []byte) {
				source := []byte("not code")
				paragraph := ast.NewParagraph()
				return paragraph, source
			},
			language: "go",
			expected: false,
		},
		{
			name: "heading_not_code_block",
			setup: func() (ast.Node, []byte) {
				source := []byte("# Heading")
				heading := ast.NewHeading(1)
				return heading, source
			},
			language: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, source := tt.setup()
			result := IsCodeBlock(node, tt.language, source)

			if result != tt.expected {
				t.Errorf("IsCodeBlock() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCodeBlockContent(t *testing.T) {
	t.Run("fenced_code_block", func(t *testing.T) {
		source := []byte("```go\nfunc main() {}\n```")
		cb := ast.NewFencedCodeBlock(nil)

		// Add lines to the code block
		lines := text.NewSegments()
		lines.Append(text.NewSegment(6, 21)) // "func main() {}\n"
		cb.SetLines(lines)

		content := GetCodeBlockContent(cb, source)
		if content != "func main() {}\n" {
			t.Errorf("GetCodeBlockContent() = %q, want %q", content, "func main() {}\n")
		}
	})

	t.Run("indented_code_block", func(t *testing.T) {
		source := []byte("    indented code\n")
		cb := ast.NewCodeBlock()

		lines := text.NewSegments()
		lines.Append(text.NewSegment(4, 18)) // "indented code\n"
		cb.SetLines(lines)

		content := GetCodeBlockContent(cb, source)
		if content != "indented code\n" {
			t.Errorf("GetCodeBlockContent() = %q, want %q", content, "indented code\n")
		}
	})

	t.Run("non_code_block_returns_empty", func(t *testing.T) {
		source := []byte("paragraph text")
		paragraph := ast.NewParagraph()

		content := GetCodeBlockContent(paragraph, source)
		if content != "" {
			t.Errorf("GetCodeBlockContent() for non-code block = %q, want empty", content)
		}
	})

	t.Run("empty_code_block", func(t *testing.T) {
		source := []byte("```\n```")
		cb := ast.NewFencedCodeBlock(nil)

		content := GetCodeBlockContent(cb, source)
		if content != "" {
			t.Errorf("GetCodeBlockContent() for empty = %q, want empty", content)
		}
	})
}

func TestGetCodeBlockLanguage(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (ast.Node, []byte)
		expected string
	}{
		{
			name: "go_language",
			setup: func() (ast.Node, []byte) {
				source := []byte("```go\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				infoSeg := text.NewSegment(3, 5)
				cb.Info = &ast.Text{}
				cb.Info.Segment = infoSeg
				return cb, source
			},
			expected: "go",
		},
		{
			name: "javascript_language",
			setup: func() (ast.Node, []byte) {
				source := []byte("```javascript\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				infoSeg := text.NewSegment(3, 13)
				cb.Info = &ast.Text{}
				cb.Info.Segment = infoSeg
				return cb, source
			},
			expected: "javascript",
		},
		{
			name: "no_language",
			setup: func() (ast.Node, []byte) {
				source := []byte("```\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				return cb, source
			},
			expected: "",
		},
		{
			name: "language_with_extra_info",
			setup: func() (ast.Node, []byte) {
				source := []byte("```python filename.py\ncode\n```")
				cb := ast.NewFencedCodeBlock(nil)
				infoSeg := text.NewSegment(3, 21)
				cb.Info = &ast.Text{}
				cb.Info.Segment = infoSeg
				return cb, source
			},
			expected: "python",
		},
		{
			name: "not_fenced_code_block",
			setup: func() (ast.Node, []byte) {
				source := []byte("    indented")
				cb := ast.NewCodeBlock()
				return cb, source
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, source := tt.setup()
			result := GetCodeBlockLanguage(node, source)

			if result != tt.expected {
				t.Errorf("GetCodeBlockLanguage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReplaceNode(t *testing.T) {
	t.Run("replace_child_node", func(t *testing.T) {
		doc := ast.NewDocument()
		oldParagraph := ast.NewParagraph()
		newParagraph := ast.NewParagraph()

		doc.AppendChild(doc, oldParagraph)

		// Replace the node
		ReplaceNode(oldParagraph, newParagraph)

		// Verify replacement
		firstChild := doc.FirstChild()
		if firstChild != newParagraph {
			t.Error("node was not replaced")
		}
	})

	t.Run("replace_node_without_parent", func(t *testing.T) {
		oldNode := ast.NewParagraph()
		newNode := ast.NewParagraph()

		// Should not panic when parent is nil
		ReplaceNode(oldNode, newNode)
	})

	t.Run("replace_middle_child", func(t *testing.T) {
		doc := ast.NewDocument()
		first := ast.NewParagraph()
		middle := ast.NewParagraph()
		last := ast.NewParagraph()
		newMiddle := ast.NewHeading(1)

		doc.AppendChild(doc, first)
		doc.AppendChild(doc, middle)
		doc.AppendChild(doc, last)

		ReplaceNode(middle, newMiddle)

		// Count children and verify types
		var children []ast.Node
		for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		if len(children) != 3 {
			t.Errorf("expected 3 children, got %d", len(children))
		}

		if children[1].Kind() != ast.KindHeading {
			t.Error("middle child should be heading after replacement")
		}
	})
}

func TestCreateEmptyParagraph(t *testing.T) {
	paragraph := CreateEmptyParagraph()

	if paragraph == nil {
		t.Fatal("CreateEmptyParagraph returned nil")
	}

	if paragraph.Kind() != ast.KindParagraph {
		t.Errorf("expected paragraph, got %v", paragraph.Kind())
	}
}

func TestCreateParagraphWithAttribute(t *testing.T) {
	key := "data-test"
	value := []byte("test-value")
	paragraph := CreateParagraphWithAttribute(key, value)

	if paragraph == nil {
		t.Fatal("CreateParagraphWithAttribute returned nil")
	}

	if paragraph.Kind() != ast.KindParagraph {
		t.Errorf("expected paragraph, got %v", paragraph.Kind())
	}

	// Check the attribute was set
	attr, ok := paragraph.Attribute([]byte(key))
	if !ok {
		t.Error("expected attribute to be set")
	}
	if attrBytes, ok := attr.([]byte); ok {
		if string(attrBytes) != string(value) {
			t.Errorf("expected attribute value %q, got %q", value, attrBytes)
		}
	} else {
		t.Error("attribute value should be []byte")
	}
}

func TestCreateTextElement(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		fontSize float64
		style    string
	}{
		{
			name:     "basic_text",
			content:  "Hello",
			fontSize: 12,
			style:    "",
		},
		{
			name:     "bold_text",
			content:  "Bold Text",
			fontSize: 14,
			style:    "B",
		},
		{
			name:     "italic_text",
			content:  "Italic",
			fontSize: 10,
			style:    "I",
		},
		{
			name:     "empty_content",
			content:  "",
			fontSize: 12,
			style:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := CreateTextElement(tt.content, tt.fontSize, tt.style)

			if element == nil {
				t.Fatal("CreateTextElement returned nil")
			}

			textElement, ok := element.(*TextElement)
			if !ok {
				t.Fatal("element is not a TextElement")
			}

			if textElement.Content != tt.content {
				t.Errorf("Content = %q, want %q", textElement.Content, tt.content)
			}

			if textElement.FontSize != tt.fontSize {
				t.Errorf("FontSize = %v, want %v", textElement.FontSize, tt.fontSize)
			}

			if textElement.Style != tt.style {
				t.Errorf("Style = %q, want %q", textElement.Style, tt.style)
			}
		})
	}
}

func TestCreateLineElement(t *testing.T) {
	tests := []struct {
		name  string
		x1    float64
		y1    float64
		x2    float64
		y2    float64
		width float64
	}{
		{
			name:  "horizontal_line",
			x1:    0,
			y1:    10,
			x2:    100,
			y2:    10,
			width: 1,
		},
		{
			name:  "vertical_line",
			x1:    50,
			y1:    0,
			x2:    50,
			y2:    100,
			width: 2,
		},
		{
			name:  "diagonal_line",
			x1:    0,
			y1:    0,
			x2:    100,
			y2:    100,
			width: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			element := CreateLineElement(tt.x1, tt.y1, tt.x2, tt.y2, tt.width)

			if element == nil {
				t.Fatal("CreateLineElement returned nil")
			}

			lineElement, ok := element.(*LineElement)
			if !ok {
				t.Fatal("element is not a LineElement")
			}

			if lineElement.X1 != tt.x1 {
				t.Errorf("X1 = %v, want %v", lineElement.X1, tt.x1)
			}

			if lineElement.Y1 != tt.y1 {
				t.Errorf("Y1 = %v, want %v", lineElement.Y1, tt.y1)
			}

			if lineElement.X2 != tt.x2 {
				t.Errorf("X2 = %v, want %v", lineElement.X2, tt.x2)
			}

			if lineElement.Y2 != tt.y2 {
				t.Errorf("Y2 = %v, want %v", lineElement.Y2, tt.y2)
			}

			if lineElement.LineWidth != tt.width {
				t.Errorf("LineWidth = %v, want %v", lineElement.LineWidth, tt.width)
			}
		})
	}
}

func TestGenerationPhaseConstants(t *testing.T) {
	// Verify generation phase constants are distinct
	phases := []GenerationPhase{BeforeContent, AfterContent, BeforeEachPage, AfterEachPage}

	seen := make(map[GenerationPhase]bool)
	for _, phase := range phases {
		if seen[phase] {
			t.Errorf("duplicate generation phase value: %v", phase)
		}
		seen[phase] = true
	}
}
