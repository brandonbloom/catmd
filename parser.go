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
	ID      string // Footnote identifier (e.g., "1" or "note")
	Content string // The footnote content text
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

func extractFootnotes(doc ast.Node, source []byte) []FootnoteInfo {
	var footnotes []FootnoteInfo

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if footnoteNode, ok := n.(*extast.Footnote); ok {
			id := string(footnoteNode.Ref)
			content := extractFootnoteMarkdown(footnoteNode, source)

			footnotes = append(footnotes, FootnoteInfo{
				ID:      id,
				Content: content,
			})
		}

		return ast.WalkContinue, nil
	})

	return footnotes
}

func extractFootnoteMarkdown(footnoteNode *extast.Footnote, source []byte) string {
	// The footnote node contains child nodes that represent its content
	// We need to extract the text content from these nodes

	// If the footnote has no children, return empty
	if footnoteNode.FirstChild() == nil {
		return ""
	}

	// Build the content by walking through child nodes
	var content strings.Builder

	for child := footnoteNode.FirstChild(); child != nil; child = child.NextSibling() {
		// For paragraph nodes, extract their text content
		if para, ok := child.(*ast.Paragraph); ok {
			// Extract text from all children of the paragraph
			for pChild := para.FirstChild(); pChild != nil; pChild = pChild.NextSibling() {
				switch node := pChild.(type) {
				case *ast.Text:
					content.Write(node.Segment.Value(source))
				case *ast.String:
					content.Write(node.Value)
				case *ast.Link:
					// For links, we need to preserve the link text
					linkText := extractTextFromNode(node, source)
					content.WriteString(linkText)
				default:
					// For other inline elements, try to extract their text
					text := extractTextFromNode(node, source)
					if text != "" {
						content.WriteString(text)
					}
				}
			}
		}
	}

	return strings.TrimSpace(content.String())
}
