# Self-Reference Test

This test verifies how catmd handles files that reference themselves:

1. **Self-link conversion**: Links to the same file (`index.md`, `./index.md`) should be converted to anchor links pointing to the document's H1 header (`#self-reference-test`)
2. **No duplicate content**: The file appears only once in the output, not duplicated
3. **Fragment links preserved**: Internal fragment links (`#conclusion`) remain unchanged
4. **No infinite loops**: Self-references don't cause infinite traversal
5. **Working anchors**: Self-reference links should create functional anchor links, not broken ones like `#index.md`

**Current behavior**: Creates broken `#index.md` anchors
**Expected behavior**: Creates working `#self-reference-test` anchors that link to the document's H1 header

This ensures that documents with self-referential links (common in documentation) are processed correctly and create meaningful, functional links.