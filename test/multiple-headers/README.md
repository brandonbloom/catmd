# Multiple Headers Test

This test verifies header rewriting rules when a file has multiple top-level headers:

1. **Multiple `#` headers detected**: Files with more than one level-1 header
2. **Synthetic header generation**: A new header is created from the filename
3. **Original headers preserved**: Existing headers become level-2 and below
4. **One top-level rule enforced**: Each file contributes exactly one `#` header to output

This ensures consistent document structure in the concatenated output.
