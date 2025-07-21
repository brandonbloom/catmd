# Cyclic References Test

This test verifies that circular references between files are handled gracefully:

1. **Cycle detection**: Files that reference each other don't cause infinite loops
2. **Visited set tracking**: Each file is processed exactly once
3. **Link conversion still works**: Circular links are still converted to section anchors
4. **No duplicate content**: Files appear only once in the output

This prevents infinite recursion while maintaining correct link behavior in circular document structures.
