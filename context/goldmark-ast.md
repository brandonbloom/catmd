# Goldmark AST Package

The AST (Abstract Syntax Tree) package defines node structures for parsing Markdown documents.

## Node Types

### Block Nodes
- Document: Root container for the entire document
- Paragraph: Text paragraph blocks  
- Heading: Headers (H1-H6)
- List: Ordered and unordered lists
- CodeBlock: Fenced and indented code blocks

### Inline Nodes
- Text: Plain text content with source segments
- Link: Hyperlinks with destination and title
- Image: Images with alt text and source
- CodeSpan: Inline code snippets
- Emphasis: Bold, italic, and other formatting

### Fundamental Types
- Block: Base interface for block-level elements
- Inline: Base interface for inline elements  
- Document: Root node containing metadata

## Core Capabilities

### Tree Traversal
- `Walk()` function for depth-first traversal
- `WalkStatus` controls traversal flow (Continue, Stop, SkipChildren)
- Custom walker functions with entering/exiting callbacks

### Node Manipulation
- `AppendChild()`: Add child nodes
- `InsertBefore()`, `InsertAfter()`: Position-specific insertion
- `RemoveChild()`: Remove child nodes
- `ReplaceChild()`: Replace existing children

### Attributes & Metadata
- Attribute management for HTML output
- Document-level metadata support
- Source segment tracking for text nodes

## Key Interfaces

### Node Interface
```go
type Node interface {
    Kind() NodeKind
    Parent() Node
    SetParent(Node)
    FirstChild() Node
    LastChild() Node
    NextSibling() Node
    PreviousSibling() Node
    // ... manipulation methods
}
```

### Walker Interface
Enables custom traversal logic with entering/exiting phases.

## Important Node Kinds
- `KindDocument`
- `KindParagraph` 
- `KindHeading`
- `KindLink`
- `KindText`
- `KindCodeBlock`
- `KindEmphasis`

## Unique Features

### Source Segment Tracking
- Text nodes store segments pointing to original source
- Enables accurate text extraction and transformation
- Critical for preserving markdown syntax during processing

### Line Break Handling
- Distinguishes hard vs soft line breaks
- Maintains formatting semantics

### Extensibility
- Supports custom node types via extensions
- Extensible attribute system
- Plugin-friendly architecture

The AST package provides a flexible, comprehensive representation of Markdown document structure optimized for parsing, transformation, and rendering operations.