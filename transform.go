package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FileProcessor struct {
	scopeDir     string
	fileOrder    map[string]int
	visitedFiles map[string]bool
}

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

func (fp *FileProcessor) ProcessFile(filename string, content []byte) ([]byte, error) {
	parsed, err := ParseMarkdownFile(content, fp.scopeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %q: %w", filename, err)
	}

	header := fp.generateFileHeader(filename, parsed.Headers)

	// Transform the content by replacing internal links and inlining footnotes
	transformedContent := fp.transformContent(string(content), filename, parsed.Links, parsed.Footnotes)

	var result strings.Builder
	if header != "" {
		result.WriteString(header)
		result.WriteString("\n\n")
	}
	result.WriteString(transformedContent)

	return []byte(result.String()), nil
}

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

func (fp *FileProcessor) transformContent(content string, currentFile string, links []LinkInfo, footnotes []FootnoteInfo) string {
	result := content

	// First, remove footnote definitions before inlining
	lines := strings.Split(result, "\n")
	var filteredLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isFootnoteDef := strings.HasPrefix(trimmed, "[^") &&
			strings.Contains(trimmed, "]:") &&
			strings.Index(trimmed, "]:") > 2 // Ensure there's content after [^
		if !isFootnoteDef {
			filteredLines = append(filteredLines, line)
		}
	}
	result = strings.Join(filteredLines, "\n")

	// Then, inline footnotes
	footnoteMap := make(map[string]string)
	for _, footnote := range footnotes {
		footnoteMap[footnote.ID] = footnote.Content
	}

	// Replace footnote references with inline content
	for _, link := range links {
		if link.IsFootnote {
			footnoteID := link.URL // URL now contains the footnote ID directly
			if content, exists := footnoteMap[footnoteID]; exists {
				// Replace [^id] with (content)
				footnoteRef := fmt.Sprintf("[^%s]", footnoteID)
				inlineContent := fmt.Sprintf(" (%s)", content)
				result = strings.ReplaceAll(result, footnoteRef, inlineContent)
			}
		}
	}

	// Process each link

	for _, link := range links {
		if link.IsInternal && !link.IsFootnote {
			// Extract fragment from original URL before resolution
			fragment := ""
			if strings.Contains(link.URL, "#") {
				parts := strings.Split(link.URL, "#")
				if len(parts) > 1 {
					fragment = "#" + strings.Join(parts[1:], "#")
				}
			}

			resolvedPath, err := fp.resolveLink(currentFile, link.URL)
			if err != nil {
				continue
			}

			if fp.visitedFiles[resolvedPath] {
				sectionLink := GenerateSectionLink(resolvedPath) + fragment

				// Find the link in the content
				// This is a simple approach - just replace the URL
				oldPattern := fmt.Sprintf("(%s)", link.URL)
				newPattern := fmt.Sprintf("(%s)", sectionLink)
				result = strings.ReplaceAll(result, oldPattern, newPattern)
			}
		}
	}

	return result
}
