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

	// Transform the content by replacing internal links
	transformedContent := fp.transformContent(string(content), filename, parsed.Links)

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

	if len(topLevelHeaders) == 1 {
		firstNode := true
		for _, h := range headers {
			if firstNode && h.Level == 1 {
				return ""
			}
			if h.Level > 0 {
				firstNode = false
			}
		}
	}

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

func (fp *FileProcessor) transformContent(content string, currentFile string, links []LinkInfo) string {
	result := content

	// Process each link

	for _, link := range links {
		if link.IsInternal && !link.IsFootnote {
			resolvedPath, err := fp.resolveLink(currentFile, link.URL)
			if err != nil {
				continue
			}

			if fp.visitedFiles[resolvedPath] {
				sectionLink := GenerateSectionLink(resolvedPath)

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
