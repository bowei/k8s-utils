# Godoc formatting

Godoc formatting is based on a set of simple, text-based conventions for comments directly preceding top-level declarations (packages, functions, types, constants, and variables). It is intentionally not Markdown and has a much simpler rule set.

A renderer should parse comment text according to the following rules.

-----

### 1\. Blocks and Paragraphs

  * **Paragraphs**: A paragraph is a sequence of one or more consecutive, non-blank lines. Paragraphs are separated by one or more blank lines.

  * **Preformatted Text (Code Blocks)**: A block of text is considered preformatted if it is indented relative to the surrounding non-indented text. Any line starting with a space or tab, where the preceding paragraph was not indented, begins a preformatted block. The block continues until a non-indented line is found. Whitespace within this block is preserved.

    ```go
    // This is a normal paragraph.
    //
    //  // This is a preformatted code block.
    //  func main() {
    //      fmt.Println("Hello")
    //  }
    //
    // This is another normal paragraph.
    ```

-----

### 2\. Headings

A **heading** is a single-line paragraph that is not indented and ends with a colon (`:`). Headings are used to create distinct sections in the documentation.

```go
// MyFunc does something important.
//
// Usage:
// This section describes how to use MyFunc.
//
// Another Heading:
// More details go here.
```

The `Deprecated:` directive is a special case. A line starting with `Deprecated:` followed by a space signals that the declaration is deprecated. The following text explains the deprecation.

-----

### 3\. Lists

  * **Bulleted Lists**: A line starting with `*`, `+`, `-`, or `•`, followed by a space or tab, begins a bulleted list item.
  * **Numbered Lists**: A line starting with a number or letter, followed by a period (`.`) or parenthesis (`)`), and then a space or tab, begins a numbered list item.

For both list types, the text of a single list item can span multiple lines. Subsequent lines must be indented to the same level as the first line's text. The list ends after a blank line or a line with less indentation.

```go
// This function performs several steps:
//
// 1. It initializes the frobulator.
//    This may take a moment.
// 2. It connects to the database.
//
// We support these protocols:
//  * TCP
//  * UDP
```

-----

### 4\. Links and Identifiers

  * **URLs**: Any text that is a valid URL (e.g., `http://example.com`) should be converted into a hyperlink.
  * **Identifier Links**: When rendering, exported identifiers from the Go source (e.g., `MyType`, `AnotherFunc`) found in the comment text should be automatically linked to their own documentation. The modern `pkgsite` renderer also supports explicit links in the format `[pkg.Identifier]` or `[Identifier.Method]`.
