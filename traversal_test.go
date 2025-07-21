package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRootFile(t *testing.T) {
	// Create temp directory for test files
	tempDir := t.TempDir()

	// Create test files
	validMD := filepath.Join(tempDir, "valid.md")
	if err := os.WriteFile(validMD, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	notMD := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(notMD, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid markdown file",
			path:    validMD,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tempDir, "missing.md"),
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name:    "directory instead of file",
			path:    testDir,
			wantErr: true,
			errMsg:  "is a directory",
		},
		{
			name:    "non-markdown file",
			path:    notMD,
			wantErr: true,
			errMsg:  "is not a markdown file",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRootFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRootFile(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateRootFile(%q) error = %v, want error containing %q", tt.path, err, tt.errMsg)
			}
		})
	}
}

func TestDetermineScopeDir(t *testing.T) {
	// Create temp directory structure
	tempDir := t.TempDir()

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	rootFile := filepath.Join(subDir, "root.md")
	if err := os.WriteFile(rootFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	notADir := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		rootFile      string
		explicitScope string
		want          string
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "default to root file directory",
			rootFile:      rootFile,
			explicitScope: "",
			want:          subDir,
			wantErr:       false,
		},
		{
			name:          "explicit scope directory",
			rootFile:      rootFile,
			explicitScope: tempDir,
			want:          tempDir,
			wantErr:       false,
		},
		{
			name:          "non-existent explicit scope",
			rootFile:      rootFile,
			explicitScope: filepath.Join(tempDir, "missing"),
			wantErr:       true,
			errMsg:        "does not exist",
		},
		{
			name:          "explicit scope is file not directory",
			rootFile:      rootFile,
			explicitScope: notADir,
			wantErr:       true,
			errMsg:        "is not a directory",
		},
		{
			name:          "non-existent root file",
			rootFile:      filepath.Join(tempDir, "missing.md"),
			explicitScope: "",
			wantErr:       true,
			errMsg:        "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetermineScopeDir(tt.rootFile, tt.explicitScope)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineScopeDir(%q, %q) error = %v, wantErr %v", tt.rootFile, tt.explicitScope, err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
				t.Errorf("DetermineScopeDir(%q, %q) error = %v, want error containing %q", tt.rootFile, tt.explicitScope, err, tt.errMsg)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DetermineScopeDir(%q, %q) = %q, want %q", tt.rootFile, tt.explicitScope, got, tt.want)
			}
		})
	}
}

func TestFileTraversal_IsWithinScope(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	scopeDir := filepath.Join(tempDir, "project")
	if err := os.Mkdir(scopeDir, 0755); err != nil {
		t.Fatal(err)
	}

	ft := &FileTraversal{
		scopeDir: scopeDir,
	}

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "file within scope",
			filename: filepath.Join(scopeDir, "file.md"),
			expected: true,
		},
		{
			name:     "file in subdirectory of scope",
			filename: filepath.Join(scopeDir, "docs", "api.md"),
			expected: true,
		},
		{
			name:     "file outside scope (parent)",
			filename: filepath.Join(tempDir, "outside.md"),
			expected: false,
		},
		{
			name:     "file outside scope (sibling)",
			filename: filepath.Join(tempDir, "other", "file.md"),
			expected: false,
		},
		{
			name:     "scope directory itself",
			filename: scopeDir,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ft.isWithinScope(tt.filename)
			if result != tt.expected {
				t.Errorf("isWithinScope(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFileTraversal_IsMarkdownFile(t *testing.T) {
	ft := &FileTraversal{}

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "lowercase .md extension",
			filename: "file.md",
			expected: true,
		},
		{
			name:     "uppercase .MD extension",
			filename: "FILE.MD",
			expected: true,
		},
		{
			name:     "mixed case .Md extension",
			filename: "File.Md",
			expected: true,
		},
		{
			name:     "lowercase .markdown extension",
			filename: "readme.markdown",
			expected: true,
		},
		{
			name:     "uppercase .MARKDOWN extension",
			filename: "README.MARKDOWN",
			expected: true,
		},
		{
			name:     "no extension",
			filename: "README",
			expected: false,
		},
		{
			name:     "different extension",
			filename: "script.js",
			expected: false,
		},
		{
			name:     "md in filename but not extension",
			filename: "md-file.txt",
			expected: false,
		},
		{
			name:     "hidden markdown file",
			filename: ".hidden.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ft.isMarkdownFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isMarkdownFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestFileTraversal_ResolveLink(t *testing.T) {
	tempDir := t.TempDir()
	currentFile := filepath.Join(tempDir, "docs", "index.md")

	ft := &FileTraversal{}

	tests := []struct {
		name        string
		currentFile string
		linkURL     string
		wantErr     bool
		checkResult func(string) bool
	}{
		{
			name:        "relative link in same directory",
			currentFile: currentFile,
			linkURL:     "api.md",
			wantErr:     false,
			checkResult: func(result string) bool {
				return filepath.Base(result) == "api.md"
			},
		},
		{
			name:        "relative link with subdirectory",
			currentFile: currentFile,
			linkURL:     "subfolder/guide.md",
			wantErr:     false,
			checkResult: func(result string) bool {
				return contains(result, "subfolder") && filepath.Base(result) == "guide.md"
			},
		},
		{
			name:        "relative link going up",
			currentFile: currentFile,
			linkURL:     "../readme.md",
			wantErr:     false,
			checkResult: func(result string) bool {
				return filepath.Base(result) == "readme.md"
			},
		},
		{
			name:        "link with fragment",
			currentFile: currentFile,
			linkURL:     "api.md#section",
			wantErr:     false,
			checkResult: func(result string) bool {
				return filepath.Base(result) == "api.md" // Fragment should be removed
			},
		},
		{
			name:        "empty link after fragment",
			currentFile: currentFile,
			linkURL:     "#section",
			wantErr:     true,
		},
		{
			name:        "absolute path",
			currentFile: currentFile,
			linkURL:     "/absolute/path.md",
			wantErr:     false,
			checkResult: func(result string) bool {
				return result == "/absolute/path.md"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ft.resolveLink(tt.currentFile, tt.linkURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLink(%q, %q) error = %v, wantErr %v", tt.currentFile, tt.linkURL, err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkResult != nil && !tt.checkResult(result) {
				t.Errorf("resolveLink(%q, %q) = %q, check failed", tt.currentFile, tt.linkURL, result)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
