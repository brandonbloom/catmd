# catmd 🐱⚕️

**Concatenates Markdown files intelligently.**

Perfect for LLM context preparation and generating [llm.txt](https://llmstxt.org) files. `catmd` recursively traverses markdown files through internal links, combining them into a single document while preserving external references and converting internal links to section anchors. Unlike pandoc's complex filters and configuration, `catmd` works with sane defaults out of the box.

## Use Cases

- **llm.txt Generation**: Automated build step for creating [llm.txt](https://llmstxt.org) files
- **Documentation Search**: Eliminate file traversal to search across linked markdown easily
- **Agent Workflows**: Create single files for feeding to Claude Code and similar tools

## Installation

**Requirements:** Go 1.24 or later

```bash
go install github.com/brandonbloom/catmd@latest
```

Or build from source:

```bash
git clone https://github.com/brandonbloom/catmd
cd catmd
go build
```

## Usage

```bash
catmd [options] <root>
```

### Options

- `-o, --output <file>` - Output file (default: stdout)
- `--scope <directory>` - Only include files within this directory (default: root file's directory)

### Example

Given these files:

**index.md**
```markdown
# My Project

Welcome to [the guide](./guide.md)!
```

**guide.md**
```markdown
# User Guide

See the [setup instructions](./setup.md) for details.
```

**setup.md**
```markdown
## Installation

Run the installer...

## Configuration

Edit the config file...
```

Running `catmd index.md` produces:

```markdown
# My Project

Welcome to [the guide](#user-guide)!

# User Guide

See the [setup instructions](#setup.md) for details.

# setup.md

## Installation

Run the installer...

## Configuration

Edit the config file...
```

## Key Features

- **Intelligent File Discovery**: Follows internal links in depth-first order (not alphabetical like `cat *.md`)
- **Smart Link Conversion**: Internal links become section anchors (`./file.md` → `#file.md`)
- **Built-in Cycle Detection**: Prevents infinite loops in circular references
- **Footnote Inlining**: Expands `[^1]` references directly into text for LLM readability
- **Scope Boundaries**: External links and files outside scope are preserved
- **Graceful Errors**: Continues processing when individual files are missing

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

**Note:** This project was heavily Vibe-Coded as a learning experience and is primarily for personal use. While tested, it prioritizes simplicity over comprehensive edge case handling.
