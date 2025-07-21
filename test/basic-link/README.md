# Basic Link Test

This test verifies the core functionality of catmd:

1. **Internal link detection**: Links to local markdown files are identified as internal
2. **Section link conversion**: Internal links are converted to section anchors (e.g., `./second.md` → `#second.md`)
3. **File inclusion**: Linked files are included in the output in traversal order
4. **Content concatenation**: Multiple files are combined with proper separation

This is the foundational test case that demonstrates the primary catmd workflow.
