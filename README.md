# catmd 🐱‍⚕️

**Concatenates Markdown files intelligently.**

`catmd` recursively traverses markdown files through internal links, intelligently
combining them into a single document while preserving external references and
converting internal links to section anchors.

## Installation

```bash
go install github.com/user/catmd@latest
```

Or build from source:

```bash
git clone https://github.com/user/catmd
cd catmd
go build
```

## Usage

```bash
catmd [options] <root>
```

### Options

- `-o, --output <file>` - Output file (default: stdout)
- `--scope <directory>` - Directory containing eligible files (default: root file's directory)

### Example

Given these files:

**index.md**
```markdown
# My Project

Welcome to [the guide](./guide.md)!
```

**guide.md**
```markdown
## Getting Started

See the [API docs](./api.md) for details.
```

**api.md**
```markdown
# API Reference

...content...
```

Running `catmd index.md` produces:

```markdown
# My Project

Welcome to [the guide](#guide.md)!

# guide.md

## Getting Started

See the [API docs](#api.md) for details.

# api.md

...content...
```

## Key Features

- **Smart Link Conversion**: Internal links become section anchors (`./file.md` → `#file.md`)
- **Header Management**: Ensures exactly one top-level header per file
- **Footnote Inlining**: Expands `[^1]` references directly into text
- **Scope Boundaries**: External links and files outside scope are preserved
- **Cycle Detection**: Prevents infinite loops in circular references
- **Graceful Errors**: Continues processing when individual files are missing

## License

MIT

Note: The MIT license expressly denies any kind of warranty. Beyond that, this
project was aggressively "vibe coded" as a learning experience, and so offers
even less warranty than none at all.
