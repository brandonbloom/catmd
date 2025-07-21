# Contributing to catmd

Thank you for your interest in contributing to catmd! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites

- **Go 1.24 or later** - Check with `go version`
- **Git** - For version control and testing

### Local Development

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/catmd
   cd catmd
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the Project**
   ```bash
   go build -o catmd
   ```

4. **Run Tests**
   ```bash
   ./test.sh
   ```

## Making Changes

### Code Style

- Follow standard Go conventions (`go fmt`, `go vet`)
- Use descriptive variable names
- Add godoc comments to exported functions
- Keep functions focused and testable

### Testing Strategy

catmd uses **snapshot testing** with shell scripts:

- **Test Location**: All tests are in `test/*/` directories
- **Test Structure**: Each test directory contains:
  - Input markdown files
  - `README.md` explaining the test purpose
  - `expected.md` with expected output (auto-generated)

### Adding New Tests

1. **Create Test Directory**
   ```bash
   mkdir test/your-feature/
   cd test/your-feature/
   ```

2. **Add Test Files**
   - Create input markdown files
   - Add `README.md` explaining what the test covers

3. **Generate Expected Output**
   ```bash
   ../../catmd input.md > expected.md
   ```

4. **Run All Tests**
   ```bash
   cd ../..
   ./test.sh
   ```

### Updating Test Expectations

When test output changes legitimately:
```bash
./test.sh --update  # Updates all expected.md files
```

## Development Workflow

### 1. Test-Driven Development

catmd follows TDD practices:

1. **Write a failing test** for new functionality
2. **Implement minimal code** to make it pass
3. **Refactor** while keeping tests green
4. **Commit** with descriptive messages

### 2. Architecture Overview

Understanding the codebase structure:

- **`parser.go`** - Goldmark-based markdown parsing
- **`traversal.go`** - File discovery and traversal logic
- **`transform.go`** - Content transformation (links, headers, footnotes)
- **`main.go`** - CLI interface and orchestration

### 3. Key Implementation Details

- **Link Classification**: Internal vs external link detection
- **Header Rules**: Synthetic header generation logic
- **Footnote Processing**: Inline expansion with content extraction
- **Error Handling**: Graceful degradation with warnings

## Submitting Changes

### Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Add tests for new functionality
   - Ensure all tests pass
   - Follow code style guidelines

3. **Commit Changes**
   ```bash
   git add .
   git commit -m "Add feature: brief description"
   ```

4. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Guidelines

- Use present tense ("Add feature" not "Added feature")
- Keep first line under 50 characters
- Reference issues when applicable
- Use imperative mood

Examples:
```
Add support for complex filename handling
Fix header detection for edge cases
Update test expectations for footnote inlining
```

## Common Tasks

### Running Individual Tests

```bash
cd test/specific-test/
../../catmd README.md
```

### Debugging Parse Issues

```bash
# Add debug output to see AST structure
ast.Dump(node, source)  # In Go code
```

### Performance Testing

```bash
# Test with large files
time ./catmd large-file.md > /dev/null
```

## Getting Help

- **Issues**: Check existing [GitHub issues](https://github.com/brandonbloom/catmd/issues)
- **Questions**: Open a discussion or issue
- **Documentation**: Review `DEVELOPING.md` for detailed implementation notes

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a welcoming environment

## Recognition

Contributors are automatically credited in commit messages. Significant contributions may be recognized in release notes.

---

Thank you for helping make catmd better! 🐱‍⚕️
