package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileTraversal struct {
	visited   map[string]bool
	scopeDir  string
	rootFile  string
	queue     []string
	fileOrder []string
}

func NewFileTraversal(rootFile, scopeDir string) *FileTraversal {
	return &FileTraversal{
		visited:   make(map[string]bool),
		scopeDir:  scopeDir,
		rootFile:  rootFile,
		queue:     []string{rootFile},
		fileOrder: []string{},
	}
}

func (ft *FileTraversal) Traverse() ([]string, error) {
	for len(ft.queue) > 0 {
		// Take from the end for depth-first traversal (stack behavior)
		currentFile := ft.queue[len(ft.queue)-1]
		ft.queue = ft.queue[:len(ft.queue)-1]

		if ft.visited[currentFile] {
			continue
		}

		ft.visited[currentFile] = true
		ft.fileOrder = append(ft.fileOrder, currentFile)

		links, err := ft.extractLinksFromFile(currentFile)
		if err != nil {
			return nil, fmt.Errorf("failed to process file %q: %w", currentFile, err)
		}

		// Add links in reverse order so they are processed in forward order
		for i := len(links) - 1; i >= 0; i-- {
			link := links[i]
			if !ft.visited[link] && ft.isWithinScope(link) {
				ft.queue = append(ft.queue, link)
			}
		}
	}

	return ft.fileOrder, nil
}

func (ft *FileTraversal) extractLinksFromFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	parsed, err := ParseMarkdownFile(content, ft.scopeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	var linkedFiles []string
	for _, link := range parsed.Links {
		if link.IsInternal && !link.IsFootnote {
			resolvedPath, err := ft.resolveLink(filename, link.URL)
			if err != nil {
				continue
			}

			if ft.fileExists(resolvedPath) {
				linkedFiles = append(linkedFiles, resolvedPath)
			}
		}
	}

	return linkedFiles, nil
}

func (ft *FileTraversal) resolveLink(currentFile, linkURL string) (string, error) {
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

func (ft *FileTraversal) isWithinScope(filename string) bool {
	absScope, err := filepath.Abs(ft.scopeDir)
	if err != nil {
		return false
	}

	absFile, err := filepath.Abs(filename)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(absScope, absFile)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "../") && relPath != ".."
}

func (ft *FileTraversal) fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func (ft *FileTraversal) isMarkdownFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".md" || ext == ".markdown"
}

func DetermineScopeDir(rootFile string, explicitScope string) (string, error) {
	if explicitScope != "" {
		abs, err := filepath.Abs(explicitScope)
		if err != nil {
			return "", fmt.Errorf("invalid scope directory %q: %w", explicitScope, err)
		}

		info, err := os.Stat(abs)
		if err != nil {
			return "", fmt.Errorf("scope directory %q does not exist: %w", abs, err)
		}

		if !info.IsDir() {
			return "", fmt.Errorf("scope path %q is not a directory", abs)
		}

		return abs, nil
	}

	rootAbs, err := filepath.Abs(rootFile)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root file path: %w", err)
	}

	if _, err := os.Stat(rootAbs); err != nil {
		return "", fmt.Errorf("root file %q does not exist: %w", rootAbs, err)
	}

	return filepath.Dir(rootAbs), nil
}

func ValidateRootFile(rootFile string) error {
	info, err := os.Stat(rootFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("root file %q does not exist", rootFile)
		}
		return fmt.Errorf("failed to access root file %q: %w", rootFile, err)
	}

	if info.IsDir() {
		return fmt.Errorf("root file %q is a directory, not a file", rootFile)
	}

	ext := strings.ToLower(filepath.Ext(rootFile))
	if ext != ".md" && ext != ".markdown" {
		return fmt.Errorf("root file %q is not a markdown file", rootFile)
	}

	return nil
}

func WalkDirectoryForMarkdown(scopeDir string) ([]string, error) {
	var markdownFiles []string

	err := filepath.WalkDir(scopeDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for %q: %w", path, err)
			}
			markdownFiles = append(markdownFiles, absPath)
		}

		return nil
	})

	return markdownFiles, err
}
