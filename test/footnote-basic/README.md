# Basic Footnote Test

This test verifies footnote inlining functionality:

1. **Footnote detection**: `[^id]` references and `[^id]: content` definitions are identified
2. **Inline expansion**: Footnote references become inline text (e.g., `[^1]` → ` (content)`)
3. **Definition removal**: Footnote definitions are removed from the output
4. **No footnotes in output**: The final result contains no footnote syntax

This ensures the output is a clean, self-contained document without footnote dependencies.
