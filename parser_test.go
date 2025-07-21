package main

import (
	"testing"
)

func TestGenerateSectionLink(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "simple filename",
			filename: "file.md",
			expected: "#file.md",
		},
		{
			name:     "filename with path",
			filename: "/path/to/document.md",
			expected: "#document.md",
		},
		{
			name:     "relative path",
			filename: "./docs/readme.md",
			expected: "#readme.md",
		},
		{
			name:     "complex filename with spaces",
			filename: "/path/to/my file.md",
			expected: "#my file.md",
		},
		{
			name:     "non-markdown file",
			filename: "script.js",
			expected: "#script.js",
		},
		{
			name:     "deeply nested path",
			filename: "../../docs/api/endpoints.md",
			expected: "#endpoints.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSectionLink(tt.filename)
			if result != tt.expected {
				t.Errorf("GenerateSectionLink(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestIsInternalLink(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		scopeDir string
		expected bool
	}{
		// External links
		{
			name:     "http URL",
			url:      "http://example.com",
			scopeDir: "/project",
			expected: false,
		},
		{
			name:     "https URL",
			url:      "https://github.com/user/repo",
			scopeDir: "/project",
			expected: false,
		},
		{
			name:     "mailto link",
			url:      "mailto:test@example.com",
			scopeDir: "/project",
			expected: false,
		},
		{
			name:     "fragment-only link",
			url:      "#section",
			scopeDir: "/project",
			expected: false,
		},
		{
			name:     "absolute path",
			url:      "/usr/local/file.md",
			scopeDir: "/project",
			expected: false,
		},
		// Internal links
		{
			name:     "relative file in same directory",
			url:      "file.md",
			scopeDir: "/project",
			expected: true,
		},
		{
			name:     "relative path with subdirectory",
			url:      "docs/api.md",
			scopeDir: "/project",
			expected: true,
		},
		{
			name:     "relative path with ./",
			url:      "./readme.md",
			scopeDir: "/project",
			expected: true,
		},
		{
			name:     "relative path going up one level",
			url:      "../sibling/file.md",
			scopeDir: "/project/docs",
			expected: false, // Goes outside scope
		},
		{
			name:     "path with fragment",
			url:      "file.md#section",
			scopeDir: "/project",
			expected: true,
		},
		{
			name:     "path with spaces",
			url:      "my%20file.md",
			scopeDir: "/project",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInternalLink(tt.url, tt.scopeDir)
			if result != tt.expected {
				t.Errorf("isInternalLink(%q, %q) = %v, want %v", tt.url, tt.scopeDir, result, tt.expected)
			}
		})
	}
}

func TestExtractTextFromNode(t *testing.T) {
	// This would require creating AST nodes, which is more complex
	// For now, we'll add a simple test case
	t.Run("nil node returns empty string", func(t *testing.T) {
		result := extractTextFromNode(nil, []byte("test"))
		if result != "" {
			t.Errorf("extractTextFromNode(nil, source) = %q, want empty string", result)
		}
	})
}
