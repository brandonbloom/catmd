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

type LinkInfo struct {
	URL        string
	Text       string
	IsInternal bool
	IsFootnote bool
}

type HeaderInfo struct {
	Level int
	Text  string
	ID    string
}

type ParsedFile struct {
	Headers []HeaderInfo
	Links   []LinkInfo
	AST     ast.Node
	Source  []byte
}

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

func ParseMarkdownFile(content []byte, scopeDir string) (*ParsedFile, error) {
	md := NewMarkdownParser()

	doc := md.Parser().Parse(text.NewReader(content))

	parsed := &ParsedFile{
		Headers: extractHeaders(doc, content),
		Links:   extractLinks(doc, content, scopeDir),
		AST:     doc,
		Source:  content,
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

func extractLinks(doc ast.Node, source []byte, scopeDir string) []LinkInfo {
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

			links = append(links, LinkInfo{
				URL:        fmt.Sprintf("[%d]", node.Index),
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

func GenerateSectionLink(filename string) string {
	base := filepath.Base(filename)
	// Keep the full filename including extension
	return "#" + base
}
