Create a new CLI tool called "catmd" with the following requirements:

# Implementation

- Programming language: Go
- Assumes "Github Flavored Markdown"

# Description

"Concatenates Markdown files intelligently."

Logo: cute cat as a doctor

# Usage

```bash
catmd [options] <root>
```

## Example:

```bash
catmd index.md > combined.md
```

### Options

- `--output <output>` (shorthand: `-o`) -- Output file to write (default: "/dev/stdout")
- `--scope <path>` -- directory containing all files eligible for concatenation (default: dirname of the root file)

# Behavior

- Using a robust markdown file parser, recursively visits all files reachable from links starting with the index file.
- Files are concatenated to the output file in prefix traversal order.
- Files are only included once (maintain a visited set while traversing).
- Output is streaming on a per-file basis. After each file is processed in memory, it is written before moving on to the next file.
- Each outputted file should produce exactly one top level (single `#`) header. See the headers section below for details.

## Headers

Headers should be rewritten according to the following rules:

- If a file has exactly one `#` header and it is at the start of the file, that is the file's sole top level header.
- If the file has more than zero or more than one such header, a synthetic header is created from the filename.

### Internal Links

An internal link is a relative path link to a file within the scope directory.
These should be converted to section links, as defined by Github Flavored
Markdown. See [Basic Writing and Formatting Syntax](./basic-writing-and-formatting-syntax.md) for "Section Links".

When an internal link is detected, it should be queued for concatenation (if it has not been seen before).

### External Links

Paths outside of the "scope" directory are considered "external" links. These should be preserved as is. External links should
not be visited.

### Footnotes

Footnote links should all be inlined and then treated as normal internal or external links. The output should not contain any footnotes.

# Example

Consider these two files:

**`first.md`**

```
# The First Document

## Some Section

[second](./second.md)
```

**`second.md`**

```
## Apple
one two three
## Banana
a b c
```

Would be concatenated as:

```
# The First Document

## Some Section

[second](#second.md)

# second.md

## Apple
one two three
## Banana
a b c
```

# Testing

We should use a shell script and "snapshot testing" with git as the snapshot of record. `./test.sh` should walk a test directory containing individual named test case directories. The output should be committed. Each feature described in this document should be tested.
