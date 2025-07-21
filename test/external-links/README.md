# External Links Test

This test verifies that external links are preserved unchanged:

1. **HTTP/HTTPS links**: Absolute URLs remain as external links
2. **Mailto links**: Email addresses are not processed
3. **Absolute file paths**: Absolute paths are treated as external
4. **Fragment-only links**: Anchor links (`#section`) remain unchanged

External links should never be converted to section links and should never trigger file inclusion.
