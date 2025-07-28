# Goldmark Extensions

Goldmark provides a rich extension system for adding additional Markdown functionality beyond the CommonMark specification.

## Key Extensions

### Footnote Extension
- Implements PHP Markdown Extra style footnotes
- Supports `[^1]` reference syntax and `[^1]: content` definitions
- Configurable options:
  - Custom ID prefix for footnote links
  - Customizable CSS classes for links and backlinks  
  - Optional title attributes for accessibility
  - Backlink rendering control

### GFM (GitHub Flavored Markdown) Extension
Comprehensive GitHub-specific Markdown features:
- **Tables**: Full table support with alignment
- **Task Lists**: `- [ ]` and `- [x]` checkbox syntax
- **Strikethrough**: `~~text~~` deletion syntax
- **Autolinks**: Automatic URL and email detection
- **Disallowed Raw HTML**: Security filtering

### Other Notable Extensions
- **Definition Lists**: Term and definition pairs
- **Typographer**: Smart quotes and punctuation
- **CJK**: Chinese/Japanese/Korean language support

## Extension Architecture

### Configuration Pattern
Extensions use functional options for configuration:
```go
Footnote = &footnote{
    options: []FootnoteOption{},
}
```

### Integration
Extensions are modular and easily integrated:
```go
goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,
        extension.Footnote,
    ),
)
```

## Footnote Extension Details

### AST Nodes
- `Footnote`: Definition nodes containing content
- `FootnoteLink`: Reference nodes linking to definitions
- `FootnoteList`: Container for all footnote definitions

### Key Properties
- Index-based referencing system
- Support for named references (e.g., `[^note]`)
- Automatic backlink generation
- Configurable HTML output

## Extension Development
- Clean separation between parsing and rendering
- AST-based transformations
- Configurable HTML output generation
- Plugin-friendly architecture

The extension system demonstrates goldmark's flexibility and makes it suitable for diverse Markdown processing needs beyond basic CommonMark compliance.