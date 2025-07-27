package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
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
	scopeDir     string          // Directory boundary for scope checking
	fileOrder    map[string]int  // Order index of each file in traversal
	visitedFiles map[string]bool // Set of files included in concatenation
}

// NewFileProcessor creates a new file processor for the given scope directory
// and list of files in traversal order.
func NewFileProcessor(scopeDir string, orderedFiles []string) *FileProcessor {
	fileOrder := make(map[string]int)
	for i, file := range orderedFiles {
		fileOrder[file] = i
	}

	visited := make(map[string]bool)
	for _, file := range orderedFiles {
		visited[file] = true
	}

	return &FileProcessor{
		scopeDir:     scopeDir,
		fileOrder:    fileOrder,
		visitedFiles: visited,
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
	transformedContent := fp.renderModifiedContent(parsed.AST, parsed.Source, filename, parsed.Links, parsed.Footnotes, needsHeaderAdjustment)

	var result strings.Builder
	if header != "" {
		result.WriteString(header)
		result.WriteString("\n\n")
	}
	result.WriteString(transformedContent)

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
func (fp *FileProcessor) renderModifiedContent(doc ast.Node, source []byte, filename string, links []LinkInfo, footnotes []FootnoteInfo, needsHeaderAdjustment bool) string {
	// Implement Header Adjustment Rules: Increment ALL headers by 1 level when
	// a synthetic header is added AND the original document had exactly 1 level-1 header
	if needsHeaderAdjustment {
		// Count existing level 1 headers
		level1Count := 0
		ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
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
			adjustHeaderLevelsInAST(doc)
		}
	}

	// Render the modified AST back to markdown with link and footnote transformations
	return fp.renderModifiedASTToMarkdownWithTransforms(doc, source, filename, links, footnotes)
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
func (fp *FileProcessor) renderModifiedASTToMarkdownWithTransforms(doc ast.Node, source []byte, filename string, links []LinkInfo, footnotes []FootnoteInfo) string {
	// Create lookup maps for transformations
	footnoteMap := make(map[string]string)
	for _, footnote := range footnotes {
		footnoteMap[footnote.ID] = footnote.Content
	}

	linkMap := make(map[string]string)
	for _, link := range links {
		if link.IsInternal && !link.IsFootnote {
			if resolvedPath, err := fp.resolveLink(filename, link.URL); err == nil {
				if fp.visitedFiles[resolvedPath] {
					fragment := ""
					if strings.Contains(link.URL, "#") {
						parts := strings.Split(link.URL, "#")
						if len(parts) > 1 {
							fragment = "#" + strings.Join(parts[1:], "#")
						}
					}
					sectionLink := GenerateSectionLink(resolvedPath) + fragment
					linkMap[link.URL] = sectionLink
				}
			}
		}
	}

	var buf strings.Builder
	footnoteIndex := 0 // Track which footnote link we're processing

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		case *ast.Document:
			// Document container - no output needed

		case *ast.Heading:
			if entering {
				// Write the heading marker
				for i := 0; i < node.Level; i++ {
					buf.WriteByte('#')
				}
				buf.WriteByte(' ')
			} else {
				buf.WriteString("\n\n")
			}

		case *ast.Paragraph:
			if !entering {
				buf.WriteString("\n")
			}

		case *ast.Text:
			if entering {
				segment := node.Segment
				text := string(source[segment.Start:segment.Stop])

				// Handle footnote references in text
				for footnoteID, content := range footnoteMap {
					footnoteRef := "[^" + footnoteID + "]"
					inlineContent := " (" + content + ")"
					text = strings.ReplaceAll(text, footnoteRef, inlineContent)
				}

				buf.WriteString(text)
			}

		case *ast.Link:
			if entering {
				buf.WriteByte('[')
			} else {
				buf.WriteString("](")

				// Transform the destination if needed
				destination := string(node.Destination)
				if newDest, exists := linkMap[destination]; exists {
					buf.WriteString(newDest)
				} else {
					buf.WriteString(destination)
				}

				if node.Title != nil {
					buf.WriteString(` "`)
					buf.Write(node.Title)
					buf.WriteByte('"')
				}
				buf.WriteByte(')')
			}

		case *ast.CodeSpan:
			if entering {
				buf.WriteByte('`')
				// CodeSpan content is in child Text nodes, let them handle the content
			} else {
				buf.WriteByte('`')
			}

		case *ast.Emphasis:
			if entering {
				if node.Level == 1 {
					buf.WriteByte('*')
				} else {
					buf.WriteString("**")
				}
			} else {
				if node.Level == 1 {
					buf.WriteByte('*')
				} else {
					buf.WriteString("**")
				}
			}

		// Note: ast.Strong doesn't exist, Emphasis with Level=2 is used for strong

		case *ast.FencedCodeBlock:
			if entering {
				buf.WriteString("```")
				if node.Language(source) != nil {
					buf.Write(node.Language(source))
				}
				buf.WriteByte('\n')
				lines := node.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					buf.Write(source[line.Start:line.Stop])
				}
				buf.WriteString("```\n")
			}

		case *ast.List:
			if !entering {
				buf.WriteByte('\n')
			}

		case *ast.ListItem:
			if entering {
				if node.Parent().(*ast.List).IsOrdered() {
					buf.WriteString("1. ")
				} else {
					buf.WriteString("- ")
				}
			} else {
				buf.WriteByte('\n')
			}

		case *ast.Blockquote:
			if entering {
				buf.WriteString("> ")
			}

		case *ast.ThematicBreak:
			if entering {
				buf.WriteString("---\n")
			}

		case *ast.HTMLBlock:
			if entering {
				lines := node.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					buf.Write(source[line.Start:line.Stop])
				}
			}

		case *ast.RawHTML:
			if entering {
				// RawHTML stores content in Segments
				for i := 0; i < node.Segments.Len(); i++ {
					segment := node.Segments.At(i)
					buf.Write(source[segment.Start:segment.Stop])
				}
			}

		default:
			// Check for footnote-related nodes by kind
			switch n.Kind().String() {
			case "FootnoteLink":
				if entering {
					// Find the corresponding footnote in our links array
					currentIndex := 0
					for _, link := range links {
						if link.IsFootnote {
							if currentIndex == footnoteIndex {
								// Found our footnote
								if content, exists := footnoteMap[link.URL]; exists {
									buf.WriteString(" (")
									buf.WriteString(content)
									buf.WriteString(")")
								}
								footnoteIndex++
								break
							}
							currentIndex++
						}
					}
				}
				return ast.WalkSkipChildren, nil
			case "Footnote", "FootnoteList":
				// Skip footnote definitions entirely
				return ast.WalkSkipChildren, nil
			}
			// For any other unhandled node types, just continue walking
		}

		return ast.WalkContinue, nil
	})

	result := buf.String()
	// Ensure output ends with a newline
	if len(result) > 0 && result[len(result)-1] != '\n' {
		result += "\n"
	}
	return result
}
