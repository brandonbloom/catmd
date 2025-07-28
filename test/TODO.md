# Comprehensive Wild Markdown Testing Checklist for catmd

**Test Status Legend:**
- ✅ Feature tested and PASS
- ❌ Feature tested and FAIL 
- ⚪ Feature not yet tested (no status)

## Core Link Processing & Navigation
- ✅ **Internal links with various formats**: `./file.md`, `file.md`, `../dir/file.md`, `/abs/path/file.md` *(basic-link, scope-boundary - PASS)*
- ✅ **Fragment links**: `file.md#section`, `./guide.md#getting-started` *(fragment-links - PASS)*
- ⚪ **Query parameters**: `file.md?param=value`, `./api.md?version=2`
- ✅ **Links with spaces**: `./my file.md`, `[text](./file with spaces.md)` *(complex-filenames - PASS)*
- ⚪ **URL-encoded links**: `./file%20name.md`, `./caf%C3%A9.md`
- ✅ **Mixed internal/external in same file**: GitHub URLs, absolute URLs, relative paths *(external-links - PASS)*
- ✅ **Circular references**: A→B→C→A link chains *(cyclic-refs - PASS)*
- ❌ **Self-references**: `[link](./current-file.md)` *(self-reference - FAIL)*
- ⚪ **Case sensitivity**: `File.MD`, `FILE.md`, `file.MD`
- ✅ **Non-existent targets**: `[broken](./missing.md)` *(error-handling - PASS)*

## Header Management & Structure
- ✅ **Files without headers**: Plain text, lists only, code blocks only *(no-headers - PASS)*
- ✅ **Multiple H1 headers**: `# First` and `# Second` in same file *(multiple-headers - PASS)*
- ✅ **Mixed header levels**: Starting with H2, H3, then H1 *(header-not-at-start - PASS)*
- ✅ **Headers with special chars**: `# API's & Services`, `# C++ Guide` *(complex-filenames - PASS)*
- ⚪ **Headers with emojis**: `# 🚀 Getting Started`, `# API 📚 Reference`
- ⚪ **Headers with inline code**: `# Using \`git status\``
- ⚪ **Headers with links**: `# See [GitHub](https://github.com)`
- ⚪ **Headers with formatting**: `# **Bold** and *italic*`
- ⚪ **Very long headers**: 200+ character titles
- ⚪ **Duplicate header names**: Multiple `## Installation` sections

## Footnotes & References
- ✅ **Basic footnotes**: `[^1]`, `[^note]`, `[^long-name]` *(footnote-basic - PASS)*
- ⚪ **Multi-line footnotes**: With line breaks and formatting
- ❌ **Footnotes with links**: `[^1]: See [GitHub](https://github.com)` *(footnote-with-links - FAIL)*
- ⚪ **Unused footnotes**: Defined but never referenced
- ⚪ **Undefined footnotes**: Referenced but not defined
- ⚪ **Footnotes in tables**: Inside table cells
- ⚪ **Footnotes in code blocks**: Should be ignored
- ⚪ **Footnotes with special chars**: `[^café]`, `[^123]`
- ⚪ **Nested footnotes**: Footnotes referencing other footnotes

## GitHub Flavored Markdown (GFM)
- ⚪ **Tables**: Basic, complex alignment, escaped pipes, nested formatting
- ⚪ **Task lists**: `- [ ]`, `- [x]`, mixed with regular lists
- ⚪ **Strikethrough**: `~~text~~`, nested with other formatting
- ⚪ **Autolinks**: URLs, emails, GitHub references
- ⚪ **Code syntax highlighting**: ```javascript, ```python, unknown languages
- ⚪ **Emoji shortcodes**: `:smile:`, `:+1:`, invalid codes
- ⚪ **GitHub alerts**: `> [!NOTE]`, `> [!WARNING]`, etc.
- ⚪ **HTML in markdown**: `<details>`, `<img>`, `<br>`

## Code Blocks & Syntax
- ⚪ **Fenced code blocks**: Triple backticks with/without language
- ⚪ **Indented code blocks**: 4-space indentation
- ⚪ **Nested code blocks**: Code within lists, blockquotes
- ⚪ **Code with backticks**: ````markdown containing ```
- ⚪ **Language identifiers**: Valid, invalid, case variations
- ⚪ **Very large code blocks**: 1000+ lines
- ⚪ **Code blocks with unicode**: Non-ASCII characters
- ⚪ **Mixed code styles**: Fenced and indented in same file

## Lists & Nesting
- ⚪ **Mixed list types**: Numbered, bulleted, task lists combined
- ⚪ **Deep nesting**: 5+ levels of indentation
- ⚪ **Inconsistent markers**: `*`, `-`, `+` mixed
- ⚪ **Lists in blockquotes**: `> - item`
- ⚪ **Lists with code blocks**: Proper indentation preservation
- ⚪ **Lists with images**: Links and media in list items
- ⚪ **Loose vs tight lists**: With/without blank lines
- ⚪ **Lists with line breaks**: Hard breaks within items

## Images & Media
- ⚪ **Relative image paths**: `![alt](./img/pic.jpg)`
- ⚪ **Absolute image paths**: `![alt](/assets/image.png)`
- ⚪ **External images**: `![alt](https://example.com/img.jpg)`
- ⚪ **Images with titles**: `![alt](pic.jpg "Title")`
- ⚪ **Images without alt text**: `![](image.png)`
- ⚪ **Broken image links**: Non-existent files
- ⚪ **Images in links**: `[![alt](img.jpg)](link.md)`
- ⚪ **Base64 images**: Data URLs
- ⚪ **Special image formats**: SVG, WebP, etc.

## Blockquotes & Formatting
- ⚪ **Nested blockquotes**: `> > text`
- ⚪ **Blockquotes with code**: Code blocks inside quotes
- ⚪ **Blockquotes with lists**: Nested list structures
- ⚪ **Empty blockquotes**: `>` with no content
- ⚪ **Multi-paragraph blockquotes**: Continued quotes
- ⚪ **Blockquotes with headers**: `> # Header`

## Edge Cases & Malformed Content
- ⚪ **Empty files**: Zero-byte files
- ⚪ **Files with only whitespace**: Spaces, tabs, newlines
- ⚪ **Binary files with .md extension**: Images, executables
- ⚪ **Very large files**: 10MB+ markdown files
- ⚪ **Files with BOM**: UTF-8 BOM markers
- ⚪ **Mixed line endings**: CRLF, LF, CR
- ⚪ **Encoding issues**: Non-UTF8 files, invalid unicode
- ⚪ **Malformed markdown**: Unclosed emphasis, broken tables
- ⚪ **HTML entities**: `&amp;`, `&lt;`, `&quot;`
- ⚪ **Raw HTML**: `<script>`, `<style>`, dangerous content

## Special Characters & Unicode
- ⚪ **Unicode in filenames**: `café.md`, `文档.md`, `🚀.md`
- ⚪ **RTL text**: Arabic, Hebrew content
- ⚪ **Mathematical symbols**: LaTeX-style, Unicode math
- ⚪ **Zero-width characters**: ZWSP, ZWNJ
- ⚪ **Control characters**: Tab, form feed, etc.
- ⚪ **Combining characters**: Accented letters
- ⚪ **Emoji variations**: Text vs emoji presentation

## Scope & File Discovery
- ✅ **Files outside scope**: `--scope` boundary testing *(scope-boundary, scope-option - PASS)*
- ⚪ **Symlinks**: To files, directories, broken links
- ⚪ **Hidden files**: `.hidden.md`, files in `.git/`
- ⚪ **Permission issues**: Unreadable files/directories
- ⚪ **Case-insensitive filesystems**: macOS, Windows behavior
- ⚪ **Long paths**: Windows 260+ character limits
- ⚪ **Special directories**: `.`, `..`, system folders

## Performance & Limits
- ✅ **Deep link chains**: 100+ linked files *(traversal-order - PASS)*
- ⚪ **Wide link graphs**: Files linking to 50+ others
- ⚪ **Recursive directories**: Very deep folder structures
- ⚪ **Memory pressure**: Processing very large documents
- ⚪ **Processing time**: Performance with complex inputs

## Infrastructure & Configuration
- ✅ **Output file option**: `--output` flag *(output-option - PASS)*
- ✅ **Unicode filenames**: Cyrillic, special characters *(complex-filenames - PASS)*
- ✅ **Graceful error handling**: Missing files, broken links *(error-handling - PASS)*
- ✅ **File inclusion order**: Deterministic traversal *(traversal-order - PASS)*
