# Traversal Order Test

This test verifies that files are processed in depth-first (prefix) traversal order:

1. **Depth-first traversal**: Files are processed in the order they would appear in a prefix tree walk
2. **Link order preserved**: Multiple links from the same file are processed in document order
3. **Correct concatenation sequence**: Output reflects the logical document flow
4. **Not breadth-first**: Verifies the implementation uses stack (LIFO) behavior, not queue (FIFO)

This ensures the concatenated output follows a logical, predictable structure.
