/*
Package main implements catmd, a markdown file concatenation tool.

# ARCHITECTURE OVERVIEW

catmd transforms multiple markdown files into a single concatenated document with
proper navigation and footnote handling. The architecture follows a three-phase pipeline:

Phase 1: PARSING (parser.go)
- Uses goldmark parser with GFM and footnote extensions
- Extracts metadata: headers, links, footnotes from each file
- Creates AST representation for content transformation
- Preserves original source for accurate text extraction

Phase 2: TRANSFORMATION (transform.go)
- Applies header generation/adjustment rules for consistent hierarchy
- Transforms internal links to section anchors for navigation
- Inlines footnotes by replacing references with content
- All transformations work at AST level, never ad-hoc text parsing

Phase 3: RENDERING
- Uses goldmark-markdown renderer to convert transformed AST back to markdown
- Maintains proper markdown syntax throughout the pipeline

KEY DESIGN PRINCIPLES
- Never parse markdown syntax ad-hoc - always use goldmark parser + AST
- Never generate markdown syntax ad-hoc - always use standard renderer
- Always apply transformations at AST level for robustness
- Preserve original source for accurate text segment extraction

DEPENDENCY USAGE
- goldmark: GitHub-compliant markdown parsing with auto heading IDs
- goldmark/extension: GFM features and footnote support
- goldmark-markdown: Standard markdown rendering from AST
- No custom markdown parsing/generation - leverages ecosystem tools

FOOTNOTE HANDLING INNOVATION
- Footnotes store content in child paragraph nodes with line segments
- Re-parse footnote markdown to create fresh, source-independent AST nodes
- Convert Text nodes (segment-based) to String nodes (value-based) for portability
- Enables automatic link transformation within inlined footnote content
*/
package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// LinkInfo represents a link found in markdown content.
type LinkInfo struct {
	URL        string // The link destination
	Text       string // The display text of the link
	IsInternal bool   // True if this is a relative link within scope
	IsFootnote bool   // True if this is a footnote reference
}

// HeaderInfo represents a heading found in markdown content.
type HeaderInfo struct {
	Level int    // Header level (1-6)
	Text  string // Header text content
	ID    string // Header ID attribute if present
}

// FootnoteInfo represents a footnote definition found in markdown content.
type FootnoteInfo struct {
	ID    string     // Footnote identifier (e.g., "1" or "note")
	Nodes []ast.Node // Fresh AST nodes from re-parsed footnote content
}

// ParsedFile contains all extracted information from a markdown file.
type ParsedFile struct {
	Headers   []HeaderInfo   // All headers found in the file
	Links     []LinkInfo     // All links found in the file
	Footnotes []FootnoteInfo // All footnote definitions found
	AST       ast.Node       // The parsed AST for content transformation
	Source    []byte         // Original source content
}

// NewMarkdownParser creates a new Goldmark parser configured for GitHub Flavored Markdown
// with footnote support and automatic heading ID generation.
//
// Key configuration choices:
//   - GFM extension: GitHub-compatible syntax (tables, strikethrough, autolinks, etc.)
//   - Footnote extension: Support for [^1] footnote references and definitions
//   - WithAutoHeadingID(): Generates GitHub-compatible anchors automatically
//     (lowercase, spaces become hyphens, punctuation removed)
func NewMarkdownParser() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

// ParseMarkdownFile parses markdown content and extracts all relevant information
// including headers, links, and footnotes. Links are classified as internal/external
// based on the provided scope directory.
//
// Returns a ParsedFile containing:
// - Headers: Extracted with goldmark's auto-generated IDs for anchor generation
// - Links: Both regular links and footnote references, classified as internal/external
// - Footnotes: Definitions with re-parsed AST nodes for inline transformation
// - AST: Full document tree for content transformation
// - Source: Original bytes for accurate text segment extraction
func ParseMarkdownFile(content []byte, scopeDir string) (*ParsedFile, error) {
	md := NewMarkdownParser()

	doc := md.Parser().Parse(text.NewReader(content))

	// First extract footnotes to get the index->ID mapping
	footnotes := extractFootnotes(doc, content)

	// Create index to ID mapping
	indexToID := make(map[int]string)
	for _, footnote := range footnotes {
		// Find the footnote node to get its index
		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if footnoteNode, ok := n.(*extast.Footnote); ok {
				if string(footnoteNode.Ref) == footnote.ID {
					indexToID[footnoteNode.Index] = footnote.ID
				}
			}
			return ast.WalkContinue, nil
		})
	}

	parsed := &ParsedFile{
		Headers:   extractHeaders(doc, content),
		Links:     extractLinks(doc, content, scopeDir, indexToID),
		Footnotes: footnotes,
		AST:       doc,
		Source:    content,
	}

	return parsed, nil
}

func extractHeaders(doc ast.Node, source []byte) []HeaderInfo {
	var headers []HeaderInfo

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			text := extractTextFromNode(heading, source)
			id := ""
			if idAttr, exists := heading.Attribute([]byte("id")); exists {
				if idBytes, ok := idAttr.([]byte); ok {
					id = string(idBytes)
				} else if idStr, ok := idAttr.(string); ok {
					id = idStr
				}
			}

			headers = append(headers, HeaderInfo{
				Level: heading.Level,
				Text:  text,
				ID:    id,
			})
		}

		return ast.WalkContinue, nil
	})

	return headers
}

func extractLinks(doc ast.Node, source []byte, scopeDir string, indexToID map[int]string) []LinkInfo {
	var links []LinkInfo

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Link:
			url := string(node.Destination)
			text := extractTextFromNode(node, source)
			isInternal := isInternalLink(url, scopeDir)

			links = append(links, LinkInfo{
				URL:        url,
				Text:       text,
				IsInternal: isInternal,
				IsFootnote: false,
			})

		case *extast.FootnoteLink:
			text := extractTextFromNode(node, source)

			// Get the actual footnote reference ID from mapping
			footnoteID := indexToID[node.Index]
			if footnoteID == "" {
				footnoteID = fmt.Sprintf("%d", node.Index) // fallback to index
			}

			links = append(links, LinkInfo{
				URL:        footnoteID,
				Text:       text,
				IsInternal: false,
				IsFootnote: true,
			})
		}

		return ast.WalkContinue, nil
	})

	return links
}

func extractTextFromNode(node ast.Node, source []byte) string {
	if node == nil {
		return ""
	}

	var buf strings.Builder

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if textNode, ok := n.(*ast.Text); ok {
			segment := textNode.Segment
			buf.Write(segment.Value(source))
		}

		return ast.WalkContinue, nil
	})

	return strings.TrimSpace(buf.String())
}

func isInternalLink(url, scopeDir string) bool {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return false
	}

	if strings.HasPrefix(url, "#") {
		return false
	}

	if strings.HasPrefix(url, "mailto:") {
		return false
	}

	if filepath.IsAbs(url) {
		return false
	}

	cleanPath := filepath.Clean(url)
	if strings.HasPrefix(cleanPath, "../") {
		absPath := filepath.Join(scopeDir, cleanPath)
		relPath, err := filepath.Rel(scopeDir, absPath)
		if err != nil || strings.HasPrefix(relPath, "../") {
			return false
		}
	}

	return true
}

// GenerateSectionLink creates a section anchor link from a filename.
// For example, "dir/file.md" becomes "#file.md".
func GenerateSectionLink(filename string) string {
	base := filepath.Base(filename)
	// Keep the full filename including extension
	return "#" + base
}

// extractFootnotes finds all footnote definitions in the document and extracts
// their content as fresh AST nodes for later inline transformation.
//
// Critical design choice: We store AST nodes instead of raw text to enable
// automatic link transformation within footnote content during the transform phase.
func extractFootnotes(doc ast.Node, source []byte) []FootnoteInfo {
	var footnotes []FootnoteInfo

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if footnoteNode, ok := n.(*extast.Footnote); ok {
			id := string(footnoteNode.Ref)
			nodes := extractFootnoteNodes(footnoteNode, source)

			footnotes = append(footnotes, FootnoteInfo{
				ID:    id,
				Nodes: nodes,
			})
		}

		return ast.WalkContinue, nil
	})

	return footnotes
}

func extractFootnoteNodes(footnoteNode *extast.Footnote, source []byte) []ast.Node {
	// Extract original source text from the footnote's paragraph children, then
	// re-parse it to create fresh AST nodes that can be safely inserted elsewhere.
	//
	// This approach handles ALL possible node types (links, emphasis, code, tables, etc.)
	// by leveraging goldmark's own parsing logic, making it future-proof and robust.

	originalText := extractFootnoteMarkdown(footnoteNode, source)
	if originalText == "" {
		return nil
	}

	// Re-parse the footnote content to get fresh AST nodes
	parser := NewMarkdownParser()
	tempDoc := parser.Parser().Parse(text.NewReader([]byte(originalText)))

	// Extract inline content from paragraphs and convert Text nodes to String nodes
	// to make them source-independent (Text nodes use segments, String nodes store content)
	var nodes []ast.Node
	for child := tempDoc.FirstChild(); child != nil; child = child.NextSibling() {
		if paragraph, ok := child.(*ast.Paragraph); ok {
			// Collect all inline children first to avoid iteration issues while modifying
			var inlineChildren []ast.Node
			for inlineChild := paragraph.FirstChild(); inlineChild != nil; inlineChild = inlineChild.NextSibling() {
				inlineChildren = append(inlineChildren, inlineChild)
			}
			// Remove all children from paragraph
			for _, inlineChild := range inlineChildren {
				paragraph.RemoveChild(paragraph, inlineChild)
			}
			// Convert and collect all children
			for _, inlineChild := range inlineChildren {
				convertedNode := convertToSourceIndependent(inlineChild, []byte(originalText))
				nodes = append(nodes, convertedNode)
			}
		} else {
			// For non-paragraph content (code blocks, etc.), keep as-is
			tempDoc.RemoveChild(tempDoc, child)
			nodes = append(nodes, child)
		}
	}

	return nodes
}

// convertToSourceIndependent converts Text nodes to String nodes and recursively
// processes children to make the entire subtree source-independent
func convertToSourceIndependent(node ast.Node, source []byte) ast.Node {
	switch n := node.(type) {
	case *ast.Text:
		// Convert Text node (segment-based) to String node (value-based)
		return ast.NewString(n.Segment.Value(source))

	case *ast.Link:
		// Links are already source-independent for destinations, but process children
		// Collect all children first to avoid iteration issues while modifying
		var children []ast.Node
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}
		// Remove all children
		for _, child := range children {
			n.RemoveChild(n, child)
		}
		// Convert and re-add children
		for _, child := range children {
			convertedChild := convertToSourceIndependent(child, source)
			n.AppendChild(n, convertedChild)
		}
		return n

	case *ast.Emphasis:
		// Process children of formatting nodes
		// Collect all children first to avoid iteration issues while modifying
		var children []ast.Node
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}
		// Remove all children
		for _, child := range children {
			n.RemoveChild(n, child)
		}
		// Convert and re-add children
		for _, child := range children {
			convertedChild := convertToSourceIndependent(child, source)
			n.AppendChild(n, convertedChild)
		}
		return n

	default:
		// For other node types, return as-is (CodeSpan, etc. should work)
		return n
	}
}

// extractFootnoteMarkdown extracts the original markdown source text from a footnote
// definition by reading line segments from child paragraph nodes.
//
// Key architectural insight: Footnotes (*extast.Footnote) don't store line segments
// themselves (footnoteNode.Lines().Len() == 0), but their child paragraph nodes DO
// have line segments that point back to the original source. This preserves exact
// markdown syntax including links like [text](url).
//
// We avoid using goldmark-markdown renderer here because it expects full document
// context and panics when rendering individual nodes or subtrees.
func extractFootnoteMarkdown(footnoteNode *extast.Footnote, source []byte) string {
	if footnoteNode.FirstChild() == nil {
		return ""
	}

	var content strings.Builder

	// Walk through footnote children to find paragraphs with line segments
	for child := footnoteNode.FirstChild(); child != nil; child = child.NextSibling() {
		if paragraph, ok := child.(*ast.Paragraph); ok {
			lines := paragraph.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				content.Write(segment.Value(source))
				if i < lines.Len()-1 {
					content.WriteByte('\n')
				}
			}
		}
	}

	return strings.TrimSpace(content.String())
}
