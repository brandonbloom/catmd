# Goldmark Overview

goldmark is a Markdown parser library for Go with several key features:

## Key Characteristics
- Fully compliant with CommonMark 0.31.2 specification
- Designed to be highly extensible
- Pure Go implementation
- Performance comparable to reference C implementations

## Core Motivation
The library was created to address limitations in existing Markdown parsers, specifically seeking a parser that is:
- Easy to extend
- Standards-compliant
- Well-structured with an Abstract Syntax Tree (AST)

## Basic Usage Example
```go
import "github.com/yuin/goldmark"

var buf bytes.Buffer
err := goldmark.Convert(source, &buf)
```

## Customization Options
- Supports multiple built-in extensions (tables, strikethrough, task lists)
- Allows custom parsing and rendering
- Provides options for HTML output

## Notable Extensions
- GitHub Flavored Markdown
- Definition Lists
- Footnotes
- Typographer
- CJK language support

## Performance
- Benchmarks show performance on par with reference implementations
- Designed to be memory-efficient

## Security
- By default, does not render raw HTML or potentially dangerous URLs
- Recommends using additional HTML sanitization for untrusted content

The library emphasizes flexibility, allowing developers to easily customize and extend Markdown parsing for their specific needs.