package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	markdown "github.com/teekennedy/goldmark-markdown"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

/*
Header Processing Rules

Goal: Every input file produces exactly one top-level header at the start of that file in the concatenated output.

Header Generation Rules - Add synthetic header `# filename.md` when:
1. 0 level-1 headers - No top-level header exists
2. Multiple level-1 headers - Multiple top-level headers would conflict
3. Single level-1 header not at start - Content before the header prevents it from being the file's opening

Header Adjustment Rules - Increment ALL existing headers by 1 level when:
- A synthetic header is added AND the original file had any level-1 headers

Logic Summary:
- 0 level-1 headers: Add synthetic `#`, keep existing headers unchanged (no conflicts)
- 1 level-1 header at start: Use existing header as-is (goal already achieved)
- 1 level-1 header not at start: Add synthetic `#`, increment all headers (resolve conflict)
- Multiple level-1 headers: Add synthetic `#`, increment all headers (resolve conflicts)

This ensures every file section in the concatenated output starts with exactly one `#` header, with proper hierarchy maintained throughout.
*/

// FileProcessor handles content transformation of markdown files,
// including header generation, link rewriting, and footnote inlining.
type FileProcessor struct {
	scopeDir     string                  // Directory boundary for scope checking
	fileOrder    map[string]int          // Order index of each file in traversal
	visitedFiles map[string]bool         // Set of files included in concatenation
	fileHeaders  map[string][]HeaderInfo // Cached header info for each file
}

// NewFileProcessor creates a new file processor for the given scope directory
// and list of files in traversal order. Pre-loads header information for all files.
func NewFileProcessor(scopeDir string, orderedFiles []string) *FileProcessor {
	fileOrder := make(map[string]int)
	for i, file := range orderedFiles {
		fileOrder[file] = i
	}

	visited := make(map[string]bool)
	for _, file := range orderedFiles {
		visited[file] = true
	}

	// Pre-load header information for all files
	fileHeaders := make(map[string][]HeaderInfo)
	for _, file := range orderedFiles {
		if content, err := os.ReadFile(file); err == nil {
			if parsed, err := ParseMarkdownFile(content, scopeDir); err == nil {
				fileHeaders[file] = parsed.Headers
			}
		}
		// If we can't read/parse a file, it will have empty headers slice
	}

	return &FileProcessor{
		scopeDir:     scopeDir,
		fileOrder:    fileOrder,
		visitedFiles: visited,
		fileHeaders:  fileHeaders,
	}
}

// ProcessFile transforms a markdown file's content by:
// 1. Generating appropriate headers according to the header rules
// 2. Converting internal links to section anchors
// 3. Inlining footnotes and removing footnote definitions
// Returns the transformed content ready for output.
func (fp *FileProcessor) ProcessFile(filename string, content []byte) ([]byte, error) {
	parsed, err := ParseMarkdownFile(content, fp.scopeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %q: %w", filename, err)
	}

	header := fp.generateFileHeader(filename, parsed.Headers)

	// Always use unified processing for consistency
	needsHeaderAdjustment := header != ""
	transformedContent, err := fp.renderModifiedContent(parsed, filename, needsHeaderAdjustment)
	if err != nil {
		return nil, fmt.Errorf("failed to render modified content for %q: %w", filename, err)
	}

	var result strings.Builder
	if header != "" {
		result.WriteString(header)
		result.WriteString("\n\n")
	}
	result.Write(transformedContent)

	return []byte(result.String()), nil
}

// generateFileHeader implements the Header Generation Rules above.
// Returns a synthetic header string (e.g., "# filename.md") if needed, or empty string if not.
// Determines when to add synthetic headers based on the count and position of level-1 headers.
func (fp *FileProcessor) generateFileHeader(filename string, headers []HeaderInfo) string {
	topLevelHeaders := make([]HeaderInfo, 0)
	for _, h := range headers {
		if h.Level == 1 {
			topLevelHeaders = append(topLevelHeaders, h)
		}
	}

	// If there are 0 or more than 1 top-level headers, create synthetic header
	if len(topLevelHeaders) != 1 {
		base := filepath.Base(filename)
		return "# " + base
	}

	// There's exactly 1 top-level header - check if it's at the start
	firstHeaderIsTopLevel := false
	for _, h := range headers {
		if h.Level > 0 {
			if h.Level == 1 {
				firstHeaderIsTopLevel = true
			}
			break
		}
	}

	if firstHeaderIsTopLevel {
		return "" // Use the existing header
	}

	// Top-level header exists but not at start, create synthetic header
	base := filepath.Base(filename)
	return "# " + base
}

func (fp *FileProcessor) isInternalLink(url, currentFile string) bool {
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

	return true
}

func (fp *FileProcessor) resolveLink(currentFile, linkURL string) (string, error) {
	currentDir := filepath.Dir(currentFile)

	if strings.Contains(linkURL, "#") {
		linkURL = strings.Split(linkURL, "#")[0]
	}

	if linkURL == "" {
		return "", fmt.Errorf("empty link after fragment removal")
	}

	var resolvedPath string
	if filepath.IsAbs(linkURL) {
		resolvedPath = linkURL
	} else {
		resolvedPath = filepath.Join(currentDir, linkURL)
	}

	cleanPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return cleanPath, nil
}

// renderModifiedContent implements the Header Adjustment Rules above.
// Applies content transformations consistently for all files, including conditional
// header level adjustment when synthetic headers are added to files with exactly 1 level-1 header.
func (fp *FileProcessor) renderModifiedContent(parsed *ParsedFile, filename string, needsHeaderAdjustment bool) ([]byte, error) {
	// Implement Header Adjustment Rules: Increment ALL headers by 1 level when
	// a synthetic header is added AND the original document had exactly 1 level-1 header
	if needsHeaderAdjustment {
		// Count existing level 1 headers
		level1Count := 0
		ast.Walk(parsed.AST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if heading, ok := n.(*ast.Heading); ok && heading.Level == 1 {
				level1Count++
			}
			return ast.WalkContinue, nil
		})

		// Adjust headers when adding synthetic header and any level-1 headers exist
		// This prevents conflicts by ensuring only the synthetic header is level-1
		if level1Count > 0 {
			adjustHeaderLevelsInAST(parsed.AST)
		}
	}

	// Render the modified AST back to markdown with link and footnote transformations
	return fp.renderModifiedASTToMarkdownWithTransforms(parsed, filename)
}

// adjustHeaderLevelsInAST increments ALL header levels by 1 to resolve conflicts.
// Called only when a synthetic header is added to a file with exactly 1 level-1 header.
// This ensures proper hierarchy: synthetic # header becomes parent, existing headers become children.
// Headers at level 6 remain at level 6 (markdown maximum).
func adjustHeaderLevelsInAST(doc ast.Node) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			if heading.Level < 6 {
				heading.Level++
			}
		}

		return ast.WalkContinue, nil
	})
}

// renderModifiedASTToMarkdownWithTransforms converts the modified AST back to markdown with link and footnote transformations
func (fp *FileProcessor) renderModifiedASTToMarkdownWithTransforms(parsed *ParsedFile, filename string) ([]byte, error) {
	// Pass 1: Inline footnotes
	if err := fp.inlineFootnotes(parsed, filename); err != nil {
		return nil, err
	}

	// Pass 2: Transform links
	if err := fp.transformLinks(parsed.AST, filename); err != nil {
		return nil, err
	}

	// Pass 3: Render to markdown using the standard renderer
	renderer := markdown.NewRenderer()
	var buf bytes.Buffer
	if err := renderer.Render(&buf, parsed.Source, parsed.AST); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// inlineFootnotes replaces footnote references with their content and removes footnote definitions
func (fp *FileProcessor) inlineFootnotes(parsed *ParsedFile, filename string) error {
	// First, create a map of footnote content with transformed links
	footnoteContentMap := make(map[string]string)

	for _, footnote := range parsed.Footnotes {
		// Footnote content now preserves original markdown syntax including links.
		//
		// Previously this code tried to parse footnote content and re-render it with
		// goldmark-markdown, but that renderer panics on individual nodes because it
		// expects full document context with renderContext state. The improved
		// extractFootnoteMarkdown() now preserves original syntax using line segments,
		// so no re-parsing/rendering is needed.
		//
		// TODO: Re-enable link transformation for internal links in footnotes
		// (currently [other.md](other.md) should become [other.md](#other-document))
		footnoteContentMap[footnote.ID] = footnote.Content
	}

	// Create index to ID mapping
	footnoteIndexToID := make(map[int]string)
	ast.Walk(parsed.AST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if fn, ok := n.(*extast.Footnote); ok {
			footnoteIndexToID[fn.Index] = string(fn.Ref)
		}
		return ast.WalkContinue, nil
	})

	// Now walk the AST and replace footnote references and remove definitions
	var nodesToRemove []ast.Node

	ast.Walk(parsed.AST, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *extast.FootnoteLink:
			// Replace footnote reference with inline content
			footnoteID := footnoteIndexToID[node.Index]
			if content, exists := footnoteContentMap[footnoteID]; exists {
				inlineContent := fmt.Sprintf(" (%s)", content)
				textNode := ast.NewString([]byte(inlineContent))
				parent := node.Parent()
				if parent != nil {
					parent.ReplaceChild(parent, node, textNode)
				}
			}
			return ast.WalkSkipChildren, nil

		case *extast.Footnote, *extast.FootnoteList:
			// Mark for removal (can't remove during walk)
			nodesToRemove = append(nodesToRemove, n)
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	// Remove footnote definitions
	for _, node := range nodesToRemove {
		if parent := node.Parent(); parent != nil {
			parent.RemoveChild(parent, node)
		}
	}

	return nil
}

// transformLinks converts internal links to section anchors
func (fp *FileProcessor) transformLinks(doc ast.Node, filename string) error {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if link, ok := n.(*ast.Link); ok {
			if fp.isInternalLink(string(link.Destination), filename) {
				if resolvedPath, err := fp.resolveLink(filename, string(link.Destination)); err == nil {
					if fp.visitedFiles[resolvedPath] {
						fragment := ""
						if strings.Contains(string(link.Destination), "#") {
							parts := strings.Split(string(link.Destination), "#")
							if len(parts) > 1 {
								fragment = "#" + strings.Join(parts[1:], "#")
							}
						}
						sectionLink := fp.generateTargetAnchor(resolvedPath) + fragment
						link.Destination = []byte(sectionLink)
					}
				}
			}
		}

		return ast.WalkContinue, nil
	})

	return nil
}

// generateTargetAnchor creates the appropriate anchor for a target file.
// If the file has an H1 header, use that header's anchor. Otherwise, use filename.
func (fp *FileProcessor) generateTargetAnchor(targetPath string) string {
	// Use cached header information
	headers, exists := fp.fileHeaders[targetPath]
	if !exists {
		// Fallback to filename if no cached headers
		return GenerateSectionLink(targetPath)
	}

	// Look for the first H1 header
	for _, header := range headers {
		if header.Level == 1 {
			// File has an H1 header, use its anchor from goldmark's auto-generated ID
			return "#" + header.ID
		}
	}

	// No H1 header found, use filename (catmd will add synthetic header)
	return GenerateSectionLink(targetPath)
}
