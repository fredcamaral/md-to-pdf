package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuin/goldmark/ast"
)

func TestNewMarkdownParser(t *testing.T) {
	parser := NewMarkdownParser()
	if parser == nil {
		t.Fatal("NewMarkdownParser returned nil")
	}
	if parser.goldmark == nil {
		t.Error("goldmark instance should not be nil")
	}
}

func TestParse_Headings(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLevel int
		expectedText  string
	}{
		{
			name:          "h1_heading",
			input:         "# Heading 1",
			expectedLevel: 1,
			expectedText:  "Heading 1",
		},
		{
			name:          "h2_heading",
			input:         "## Heading 2",
			expectedLevel: 2,
			expectedText:  "Heading 2",
		},
		{
			name:          "h3_heading",
			input:         "### Heading 3",
			expectedLevel: 3,
			expectedText:  "Heading 3",
		},
		{
			name:          "h4_heading",
			input:         "#### Heading 4",
			expectedLevel: 4,
			expectedText:  "Heading 4",
		},
		{
			name:          "h5_heading",
			input:         "##### Heading 5",
			expectedLevel: 5,
			expectedText:  "Heading 5",
		},
		{
			name:          "h6_heading",
			input:         "###### Heading 6",
			expectedLevel: 6,
			expectedText:  "Heading 6",
		},
		{
			name:          "heading_with_special_chars",
			input:         "# Hello World!",
			expectedLevel: 1,
			expectedText:  "Hello World!",
		},
	}

	parser := NewMarkdownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.input)
			node, err := parser.Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Find the heading node
			var headingNode *ast.Heading
			err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && n.Kind() == ast.KindHeading {
					headingNode = n.(*ast.Heading)
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if headingNode == nil {
				t.Fatal("No heading node found")
			}

			if headingNode.Level != tt.expectedLevel {
				t.Errorf("heading level = %d, want %d", headingNode.Level, tt.expectedLevel)
			}

			// Extract text content
			var text string
			for child := headingNode.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == ast.KindText {
					textNode := child.(*ast.Text)
					text += string(textNode.Segment.Value(source))
				}
			}

			if text != tt.expectedText {
				t.Errorf("heading text = %q, want %q", text, tt.expectedText)
			}
		})
	}
}

func TestParse_Paragraphs(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedText string
	}{
		{
			name:         "simple_paragraph",
			input:        "This is a simple paragraph.",
			expectedText: "This is a simple paragraph.",
		},
		{
			name:         "paragraph_with_multiple_words",
			input:        "Hello world this is a test.",
			expectedText: "Hello world this is a test.",
		},
		{
			name:         "paragraph_with_numbers",
			input:        "This costs $100 and has 5 items.",
			expectedText: "This costs $100 and has 5 items.",
		},
	}

	parser := NewMarkdownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.input)
			node, err := parser.Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Find the paragraph node
			var paragraphNode *ast.Paragraph
			err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && n.Kind() == ast.KindParagraph {
					paragraphNode = n.(*ast.Paragraph)
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if paragraphNode == nil {
				t.Fatal("No paragraph node found")
			}

			// Extract text content
			var text string
			for child := paragraphNode.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == ast.KindText {
					textNode := child.(*ast.Text)
					text += string(textNode.Segment.Value(source))
				}
			}

			if text != tt.expectedText {
				t.Errorf("paragraph text = %q, want %q", text, tt.expectedText)
			}
		})
	}
}

func TestParse_CodeBlocks(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedLang   string
		expectedCode   string
		isFencedBlock  bool
	}{
		{
			name: "fenced_code_block_go",
			input: "```go\nfunc main() {}\n```",
			expectedLang:  "go",
			expectedCode:  "func main() {}\n",
			isFencedBlock: true,
		},
		{
			name: "fenced_code_block_python",
			input: "```python\nprint('hello')\n```",
			expectedLang:  "python",
			expectedCode:  "print('hello')\n",
			isFencedBlock: true,
		},
		{
			name: "fenced_code_block_no_language",
			input: "```\nsome code\n```",
			expectedLang:  "",
			expectedCode:  "some code\n",
			isFencedBlock: true,
		},
		{
			name: "fenced_code_block_javascript",
			input: "```javascript\nconsole.log('test');\n```",
			expectedLang:  "javascript",
			expectedCode:  "console.log('test');\n",
			isFencedBlock: true,
		},
		{
			name: "indented_code_block",
			input: "    indented code",
			expectedLang:  "",
			expectedCode:  "indented code",
			isFencedBlock: false,
		},
	}

	parser := NewMarkdownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.input)
			node, err := parser.Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Find the code block node
			var codeNode ast.Node
			var foundKind ast.NodeKind
			err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					if n.Kind() == ast.KindFencedCodeBlock || n.Kind() == ast.KindCodeBlock {
						codeNode = n
						foundKind = n.Kind()
						return ast.WalkStop, nil
					}
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if codeNode == nil {
				t.Fatal("No code block node found")
			}

			if tt.isFencedBlock {
				if foundKind != ast.KindFencedCodeBlock {
					t.Errorf("expected fenced code block, got %v", foundKind)
				}

				fencedBlock := codeNode.(*ast.FencedCodeBlock)

				// Check language
				var lang string
				if fencedBlock.Info != nil {
					lang = string(fencedBlock.Info.Segment.Value(source))
				}
				if lang != tt.expectedLang {
					t.Errorf("code language = %q, want %q", lang, tt.expectedLang)
				}

				// Extract code content
				var code string
				lines := fencedBlock.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					code += string(line.Value(source))
				}
				if code != tt.expectedCode {
					t.Errorf("code content = %q, want %q", code, tt.expectedCode)
				}
			} else {
				if foundKind != ast.KindCodeBlock {
					t.Errorf("expected indented code block, got %v", foundKind)
				}

				codeBlock := codeNode.(*ast.CodeBlock)

				// Extract code content
				var code string
				lines := codeBlock.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					code += string(line.Value(source))
				}
				if code != tt.expectedCode {
					t.Errorf("code content = %q, want %q", code, tt.expectedCode)
				}
			}
		})
	}
}

func TestParse_MixedContent(t *testing.T) {
	input := `# Main Title

This is an introduction paragraph.

## Section 1

Some text here.

` + "```go\nfunc example() {}\n```" + `

### Subsection

More text.
`

	parser := NewMarkdownParser()
	source := []byte(input)
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Count different node types
	var headingCount, paragraphCount, codeBlockCount int

	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindHeading:
			headingCount++
		case ast.KindParagraph:
			paragraphCount++
		case ast.KindFencedCodeBlock:
			codeBlockCount++
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if headingCount != 3 {
		t.Errorf("heading count = %d, want 3", headingCount)
	}
	if paragraphCount != 3 {
		t.Errorf("paragraph count = %d, want 3", paragraphCount)
	}
	if codeBlockCount != 1 {
		t.Errorf("code block count = %d, want 1", codeBlockCount)
	}
}

func TestParse_EmptyContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty_string",
			input: "",
		},
		{
			name:  "only_whitespace",
			input: "   ",
		},
		{
			name:  "only_newlines",
			input: "\n\n\n",
		},
		{
			name:  "only_tabs",
			input: "\t\t",
		},
	}

	parser := NewMarkdownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.input)
			node, err := parser.Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if node == nil {
				t.Fatal("Parse returned nil node for empty content")
			}

			// Should be a document node with no meaningful children
			if node.Kind() != ast.KindDocument {
				t.Errorf("expected document node, got %v", node.Kind())
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Run("parse_valid_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.md")

		content := "# Test File\n\nThis is content."
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		parser := NewMarkdownParser()
		node, err := parser.ParseFile(testFile)
		if err != nil {
			t.Fatalf("ParseFile failed: %v", err)
		}

		if node == nil {
			t.Fatal("ParseFile returned nil node")
		}

		// Verify content was parsed
		var hasHeading bool
		err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if entering && n.Kind() == ast.KindHeading {
				hasHeading = true
				return ast.WalkStop, nil
			}
			return ast.WalkContinue, nil
		})
		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		if !hasHeading {
			t.Error("expected heading node in parsed file")
		}
	})

	t.Run("parse_nonexistent_file", func(t *testing.T) {
		parser := NewMarkdownParser()
		_, err := parser.ParseFile("/nonexistent/file.md")
		if err == nil {
			t.Error("ParseFile should fail for nonexistent file")
		}
	})

	t.Run("parse_empty_file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.md")

		if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		parser := NewMarkdownParser()
		node, err := parser.ParseFile(testFile)
		if err != nil {
			t.Fatalf("ParseFile failed for empty file: %v", err)
		}

		if node == nil {
			t.Fatal("ParseFile returned nil node for empty file")
		}
	})
}

func TestParse_Lists(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		ordered   bool
		itemCount int
	}{
		{
			name:      "unordered_list",
			input:     "- Item 1\n- Item 2\n- Item 3",
			ordered:   false,
			itemCount: 3,
		},
		{
			name:      "ordered_list",
			input:     "1. First\n2. Second\n3. Third",
			ordered:   true,
			itemCount: 3,
		},
		{
			name:      "single_item_list",
			input:     "- Single item",
			ordered:   false,
			itemCount: 1,
		},
	}

	parser := NewMarkdownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.input)
			node, err := parser.Parse(source)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			var listNode *ast.List
			err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering && n.Kind() == ast.KindList {
					listNode = n.(*ast.List)
					return ast.WalkStop, nil
				}
				return ast.WalkContinue, nil
			})
			if err != nil {
				t.Fatalf("Walk failed: %v", err)
			}

			if listNode == nil {
				t.Fatal("No list node found")
			}

			if listNode.IsOrdered() != tt.ordered {
				t.Errorf("list ordered = %v, want %v", listNode.IsOrdered(), tt.ordered)
			}

			// Count list items
			itemCount := 0
			for child := listNode.FirstChild(); child != nil; child = child.NextSibling() {
				if child.Kind() == ast.KindListItem {
					itemCount++
				}
			}

			if itemCount != tt.itemCount {
				t.Errorf("list item count = %d, want %d", itemCount, tt.itemCount)
			}
		})
	}
}

func TestParse_Blockquotes(t *testing.T) {
	input := "> This is a quote\n> continued here"

	parser := NewMarkdownParser()
	source := []byte(input)
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var blockquoteNode *ast.Blockquote
	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindBlockquote {
			blockquoteNode = n.(*ast.Blockquote)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if blockquoteNode == nil {
		t.Fatal("No blockquote node found")
	}
}

func TestParse_Links(t *testing.T) {
	input := "[Link Text](https://example.com)"

	parser := NewMarkdownParser()
	source := []byte(input)
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var linkNode *ast.Link
	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindLink {
			linkNode = n.(*ast.Link)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if linkNode == nil {
		t.Fatal("No link node found")
	}

	destination := string(linkNode.Destination)
	if destination != "https://example.com" {
		t.Errorf("link destination = %q, want %q", destination, "https://example.com")
	}
}

func TestParse_Images(t *testing.T) {
	input := "![Alt Text](image.png)"

	parser := NewMarkdownParser()
	source := []byte(input)
	node, err := parser.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var imageNode *ast.Image
	err = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindImage {
			imageNode = n.(*ast.Image)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if imageNode == nil {
		t.Fatal("No image node found")
	}

	destination := string(imageNode.Destination)
	if destination != "image.png" {
		t.Errorf("image destination = %q, want %q", destination, "image.png")
	}
}
