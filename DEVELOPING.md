# Development Guide

This guide covers the implementation approach, architecture, and development workflow for `catmd`.

## Architecture Overview

The codebase is structured into distinct modules:

- **`parser.go`** - Markdown parsing using Goldmark, extracts links/headers/footnotes
- **`traversal.go`** - File discovery and traversal with cycle detection  
- **`transform.go`** - Content transformation (link rewriting, header generation)
- **`main.go`** - CLI interface and orchestration

### Key Dependencies

- **[Goldmark](https://github.com/yuin/goldmark)** - CommonMark/GFM compliant parser
- **Extensions**: `github.com/yuin/goldmark/extension` for GFM support

## Implementation Approach

### 1. Parsing Phase
Files are parsed into an AST to extract:
- Links (with internal/external classification)
- Headers (to determine top-level header rules)
- Footnotes (for content extraction and inlining)

### 2. Traversal Phase  
Starting from the root file, performs depth-first traversal:
- Maintains visited set to prevent cycles
- Queues internal links for processing
- Respects scope boundaries

### 3. Transform Phase
For each file in traversal order:
- Applies header rewriting rules
- Converts internal links to section anchors
- Inlines footnotes with content expansion
- Preserves external links unchanged

### 4. Output Phase
Streams processed content with file separators.

## Development Workflow

### Building

```bash
go build -o catmd
```

### Testing

The project uses snapshot testing with shell scripts:

```bash
./test.sh
```

This runs all test cases in `test/*/` directories, comparing actual output against
committed expected output.

### Adding Tests

1. Create new directory: `test/feature-name/`
2. Add input files and `README.md` explaining the test
3. Run `./test.sh` to generate initial output
4. Review and commit the expected output

### TDD Workflow

1. Write failing test case
2. Implement minimal code to make it pass  
3. Refactor while keeping tests green
4. Commit with descriptive message

## Key Implementation Details

### Link Classification

Internal links are relative paths within the scope directory. Classification happens during parsing:

```go
func classifyLink(linkURL, currentFile, scopeDir string) (bool, error) {
    // Resolve relative to current file
    // Check if resolved path is within scope
}
```

### Header Rules

Files get exactly one top-level header based on these rules:

1. **Single `#` header at start** → Keep as-is
2. **Multiple or zero `#` headers** → Generate from filename  
3. **Single `#` header not at start** → Generate from filename

### Footnote Processing

Footnotes are expanded inline during transformation:

- `[^1]` becomes ` (content of footnote 1)`
- Footnote definitions are removed
- Footnote content is processed for links

### Error Handling

The tool continues processing despite individual file errors:
- Missing files: Log warning, preserve link as external
- Parse errors: Log warning, skip file
- Permission errors: Log warning, continue

## Testing Strategy

### Test Coverage

The test suite covers all specification requirements:

- **Basic functionality**: Link conversion, file concatenation
- **Header rules**: All combinations of header scenarios  
- **Edge cases**: Complex filenames, fragments, cycles
- **Error conditions**: Missing files, permissions
- **Options**: Output files, scope directories

### Snapshot Testing

Each test case directory contains:
- Input markdown files
- `README.md` explaining the test purpose
- `expected.md` with expected output (auto-generated)

The `test.sh` script:
1. Runs `catmd` on test inputs
2. Compares against expected output
3. Reports any differences

## Code Style

- Follow standard Go conventions
- Use descriptive variable names
- Add comments for complex logic only
- Keep functions focused and testable
- Prefer composition over inheritance

## Debugging Tips

### Verbose Output

Add debug prints to see traversal order:

```go
fmt.Fprintf(os.Stderr, "Processing: %s\n", filename)
```

### AST Inspection

Use Goldmark's AST dump for debugging parsing issues:

```go
ast.Dump(node, source)
```

### Test Isolation

Run individual tests:

```bash
cd test/specific-test/
../../catmd README.md
```
