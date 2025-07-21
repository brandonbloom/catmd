package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestFileProcessor_GenerateFileHeader(t *testing.T) {
	fp := &FileProcessor{}

	tests := []struct {
		name     string
		filename string
		headers  []HeaderInfo
		expected string
	}{
		{
			name:     "no headers",
			filename: "/path/to/file.md",
			headers:  []HeaderInfo{},
			expected: "# file.md",
		},
		{
			name:     "single top-level header at start",
			filename: "/path/to/document.md",
			headers: []HeaderInfo{
				{Level: 1, Text: "Main Title", ID: "main-title"},
			},
			expected: "", // No synthetic header needed
		},
		{
			name:     "single top-level header not at start",
			filename: "/path/to/readme.md",
			headers: []HeaderInfo{
				{Level: 2, Text: "Intro", ID: "intro"},
				{Level: 1, Text: "Main Title", ID: "main-title"},
			},
			expected: "# readme.md",
		},
		{
			name:     "multiple top-level headers",
			filename: "/docs/guide.md",
			headers: []HeaderInfo{
				{Level: 1, Text: "First", ID: "first"},
				{Level: 2, Text: "Sub", ID: "sub"},
				{Level: 1, Text: "Second", ID: "second"},
			},
			expected: "# guide.md",
		},
		{
			name:     "only sub-headers",
			filename: "api.md",
			headers: []HeaderInfo{
				{Level: 2, Text: "Methods", ID: "methods"},
				{Level: 3, Text: "GET", ID: "get"},
				{Level: 2, Text: "Examples", ID: "examples"},
			},
			expected: "# api.md",
		},
		{
			name:     "complex filename with spaces",
			filename: "/path/to/my awesome file.md",
			headers:  []HeaderInfo{},
			expected: "# my awesome file.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.generateFileHeader(tt.filename, tt.headers)
			if result != tt.expected {
				t.Errorf("generateFileHeader(%q, headers) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFileProcessor_IsInternalLink(t *testing.T) {
	fp := &FileProcessor{}

	tests := []struct {
		name        string
		url         string
		currentFile string
		expected    bool
	}{
		{
			name:        "http URL",
			url:         "http://example.com",
			currentFile: "/project/file.md",
			expected:    false,
		},
		{
			name:        "https URL",
			url:         "https://github.com",
			currentFile: "/project/file.md",
			expected:    false,
		},
		{
			name:        "mailto link",
			url:         "mailto:test@example.com",
			currentFile: "/project/file.md",
			expected:    false,
		},
		{
			name:        "fragment only",
			url:         "#section",
			currentFile: "/project/file.md",
			expected:    false,
		},
		{
			name:        "absolute path",
			url:         "/absolute/path.md",
			currentFile: "/project/file.md",
			expected:    false,
		},
		{
			name:        "relative path",
			url:         "other.md",
			currentFile: "/project/file.md",
			expected:    true,
		},
		{
			name:        "relative with directory",
			url:         "docs/api.md",
			currentFile: "/project/file.md",
			expected:    true,
		},
		{
			name:        "relative with fragment",
			url:         "file.md#section",
			currentFile: "/project/file.md",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.isInternalLink(tt.url, tt.currentFile)
			if result != tt.expected {
				t.Errorf("isInternalLink(%q, %q) = %v, want %v", tt.url, tt.currentFile, result, tt.expected)
			}
		})
	}
}

func TestFileProcessor_ResolveLink(t *testing.T) {
	fp := &FileProcessor{}
	currentFile := "/project/docs/index.md"

	tests := []struct {
		name     string
		linkURL  string
		wantErr  bool
		contains string // Check if result contains this string
	}{
		{
			name:     "simple relative link",
			linkURL:  "api.md",
			wantErr:  false,
			contains: "api.md",
		},
		{
			name:     "relative with subdirectory",
			linkURL:  "guides/tutorial.md",
			wantErr:  false,
			contains: filepath.Join("guides", "tutorial.md"),
		},
		{
			name:     "link with fragment",
			linkURL:  "api.md#methods",
			wantErr:  false,
			contains: "api.md",
		},
		{
			name:    "fragment only",
			linkURL: "#section",
			wantErr: true,
		},
		{
			name:    "empty link",
			linkURL: "",
			wantErr: true,
		},
		{
			name:     "parent directory",
			linkURL:  "../readme.md",
			wantErr:  false,
			contains: "readme.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fp.resolveLink(currentFile, tt.linkURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLink(%q, %q) error = %v, wantErr %v", currentFile, tt.linkURL, err, tt.wantErr)
			}
			if !tt.wantErr && tt.contains != "" && !strings.Contains(result, tt.contains) {
				t.Errorf("resolveLink(%q, %q) = %q, want to contain %q", currentFile, tt.linkURL, result, tt.contains)
			}
		})
	}
}

func TestFileProcessor_TransformContent(t *testing.T) {
	scopeDir := "/project"
	orderedFiles := []string{
		"/project/index.md",
		"/project/docs/api.md",
		"/project/guide.md",
	}

	fp := NewFileProcessor(scopeDir, orderedFiles)

	tests := []struct {
		name        string
		content     string
		currentFile string
		links       []LinkInfo
		footnotes   []FootnoteInfo
		expected    string
	}{
		{
			name:        "simple link transformation",
			content:     "See the [API](./docs/api.md) for details.",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "./docs/api.md", Text: "API", IsInternal: true, IsFootnote: false},
			},
			footnotes: []FootnoteInfo{},
			expected:  "See the [API](#api.md) for details.",
		},
		{
			name:        "footnote inlining",
			content:     "This is important[^1].\n\n[^1]: Very important indeed!",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "1", Text: "", IsInternal: false, IsFootnote: true},
			},
			footnotes: []FootnoteInfo{
				{ID: "1", Content: "Very important indeed!"},
			},
			expected: "This is important (Very important indeed!).\n",
		},
		{
			name:        "preserve external links",
			content:     "Visit [GitHub](https://github.com) for more.",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "https://github.com", Text: "GitHub", IsInternal: false, IsFootnote: false},
			},
			footnotes: []FootnoteInfo{},
			expected:  "Visit [GitHub](https://github.com) for more.",
		},
		{
			name:        "link with fragment",
			content:     "See [methods](./docs/api.md#methods) section.",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "./docs/api.md#methods", Text: "methods", IsInternal: true, IsFootnote: false},
			},
			footnotes: []FootnoteInfo{},
			expected:  "See [methods](#api.md#methods) section.",
		},
		{
			name:        "multiple footnotes",
			content:     "First[^1] and second[^2].\n\n[^1]: Note one\n[^2]: Note two",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "1", Text: "", IsInternal: false, IsFootnote: true},
				{URL: "2", Text: "", IsInternal: false, IsFootnote: true},
			},
			footnotes: []FootnoteInfo{
				{ID: "1", Content: "Note one"},
				{ID: "2", Content: "Note two"},
			},
			expected: "First (Note one) and second (Note two).\n",
		},
		{
			name:        "unvisited internal link not transformed",
			content:     "See [unvisited](./other.md) file.",
			currentFile: "/project/index.md",
			links: []LinkInfo{
				{URL: "./other.md", Text: "unvisited", IsInternal: true, IsFootnote: false},
			},
			footnotes: []FootnoteInfo{},
			expected:  "See [unvisited](./other.md) file.", // Not transformed because not in visitedFiles
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fp.transformContent(tt.content, tt.currentFile, tt.links, tt.footnotes)
			if result != tt.expected {
				t.Errorf("transformContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}
