# `rex` Design Document

## Overview

`rex` is a command-line utility written in Go that parses Go source code to
extract information about struct types and their fields. It generates an
interactive, multi-column HTML interface for visualizing these data structures
and their relationships, similar to macOS Finder's column view. The tool can
also output the extracted data as a JSON object for testing or consumption by
other tools.

The primary goal is to provide developers with an intuitive way to browse and
understand complex Go data structures within a project or across multiple
packages.

## Architecture

The application is composed of two main parts: a Go-based backend for parsing
and data extraction, and a self-contained HTML/CSS/JavaScript frontend for
visualization.

## Backend (Go)

The backend is responsible for processing command-line arguments, finding and
parsing Go source files, building a graph of type information, and generating
the final output.

### Key Components:

*   **Command-Line Interface (CLI):** Built using the standard `flag`
    package. It handles:
    *   `-output <file>`: Specifies the destination for the generated HTML file
        (defaults to `godoc.html`, `-` for stdout).
    *   `-json`: If present, switches the output format from HTML to JSON.
    *   `-type <type>`: the type to start with.
    *   Positional arguments: A list of Go package directories to parse.

*   **Parsing Logic (`parsePackage`):** This is the core of the tool.
    *   It uses the standard `go/parser` package to build an Abstract Syntax
        Tree (AST) for all source files in a package directory.
    *   It then uses the `go/doc` package to analyze the AST, which is more
        powerful as it associates documentation and related declarations (like
        constants) with their respective types.
    *   It iterates through all exported types found by `go/doc`. For each type,
        it determines if it's a struct or a potential enum.
    *   For **struct types**, it constructs a fully qualified type name (e.g.,
        `package/path.TypeName`) which serves as a unique identifier. It
        extracts all exported fields, their types (resolving them to fully
        qualified names), and their documentation. It handles various field
        type expressions, including pointers (`*ast.StarExpr`), arrays
        (`*ast.ArrayType`), maps (`*ast.MapType`), and qualified types from
        other packages (`*ast.SelectorExpr`).
    *   For **enum types**, it identifies types defined on a string or numeric
        base type (e.g., `type Status string`) that have one or more exported
        constants of that type associated with them. It extracts the names and
        documentation of these constants as the enum values.
    *   **Docstring Parsing:** Instead of treating docstrings as plain text, the
        tool parses them into a structured format (`GoDocString`). This
        structure separates paragraphs, headings, lists, and code blocks,
        allowing the frontend to render them with rich formatting.
    *   While parsing struct fields, it discovers imported packages and adds
        them to a queue to be parsed recursively.

*   **Package Discovery and Traversal:**
    *   The tool starts with the set of packages provided as command-line
        arguments.
    *   It uses a queue to manage a list of packages to parse. As it parses a
        package, it discovers imported packages that may contain relevant types.
    *   The `resolvePkgDir` function uses the `go list -f '{{.Dir}}'` command to
        find the absolute file path of an imported package, which is then added
        to the parsing queue. This allows the tool to recursively parse all
        relevant local dependencies.
    *   To avoid parsing the Go standard library, it uses a heuristic: any
        package where the first component of the path does not contain a dot
        (e.g., `fmt`, `net/http`) is ignored.

*   **Data Model:** The extracted information is stored in a graph structure
    represented by a map where keys are fully qualified type names. The values
    are `TypeInfo` objects with the following structure:
    *   `TypeInfo`:
        *   `Package`: The package import path.
        *   `TypeName`: The unqualified type name.
        *   `DocString`: The raw documentation string for the type.
        *   `ParsedDocString`: A structured representation of the docstring,
            used for rich rendering.
        *   `IsRoot`: A boolean indicating if the type is a "root" type. A type
            is considered a root if it contains `TypeMeta` and `ObjectMeta`
            fields (like a Kubernetes resource) or if its name ends with
            `Request` or `Response`.
        *   `Fields`: A slice of `FieldInfo` objects for struct types.
        *   `EnumValues`: A slice of `EnumInfo` objects for enum types.
    *   `FieldInfo`:
        *   `FieldName`: The name of the struct field.
        *   `TypeName`: The fully qualified type of the field.
        *   `TypeDecorators`: A list of strings representing type modifiers like
            pointers (`"Ptr"`), lists (`"List"`), or maps (`"Map[keyType]"`).
        *   `DocString`: The raw documentation string for the field.
        *   `ParsedDocString`: A structured representation of the docstring.
    *   `EnumInfo`:
        *   `Name`: The name of the enum constant.
        *   `DocString`: The raw documentation string for the enum constant.
        *   `ParsedDocString`: A structured representation of the docstring.
    *   A `map[string]TypeInfo` is used to store all discovered types, with the
        *fully qualified* type name as the key (e.g.,
        `"github.com/user/project/package.MyStruct"`).
    *   When traversing packages, ignore packages in the Go standard library.
    *   Non-exported types and fields should not be traversed.
    *   Enum types: an enum type is an Exported type that has a string or numeric
      underlying type that has one or more exported `const` values.

*   **Output Generation:**
    *   **HTML (`generateHTML`):** The `map[string]TypeInfo` is marshaled into a
        JSON string. This JSON is then embedded directly into a Go
        `html/template`. The template contains all the necessary HTML, CSS, and
        JavaScript for the frontend, creating a single, self-contained `.html`
        file.
    *   **JSON:** If the `--json` flag is used, the `map[string]TypeInfo` is
        simply marshaled into a formatted JSON string and printed to standard
        output.

### Generation Workflow

1.  User executes `rex <package-dirs...>`.
2.  The `main` function parses flags and initializes the package queue with the
    provided directories.
3.  The program enters a loop that continues as long as the queue is not empty.
4.  In the loop, it dequeues a package path and calls `parsePackage`.
5.  `parsePackage` reads all `.go` files in the directory, builds an AST, and
    inspects it for struct definitions, populating the global `allTypes` map. It
    returns a list of any new external packages it discovered via imports.
6.  The `main` function calls `resolvePkgDir` for each new external package to
    get its file system path and adds it to the queue.
7.  Once the loop finishes, the `allTypes` map contains the complete type graph.
8.  Based on the `--json` flag, the program either marshals `allTypes` to JSON
    and prints it, or it calls `generateHTML`.
9.  `generateHTML` marshals `allTypes` to JSON, injects it into the HTML
    template, and writes the final, self-contained HTML file to the location
    specified by `--output`.

## Frontend (HTML/CSS/JavaScript)

The frontend is a single-page application embedded within the generated HTML
file. It is responsible for rendering the interactive multi-column view.

*   **Structure (HTML):** A main container (`<div id="main-container">`) holds a
    series of `.column` divs.
*   **Styling (CSS):** The CSS provides a clean, modern look. It styles the
    columns, headers, and list items, including hover and selection states. It
    uses Flexbox to manage the layout of the columns.
*   **Rich Docstrings:** The frontend consumes the `parsedDocString` data to
    render comments with their original formatting, including paragraphs,
    headings, lists, and code blocks. It also automatically detects and
    converts URLs into clickable links.

### Field browsing

*   **Data:** The Go type data is available globally as a JavaScript object
    (`typeData`), injected by the Go template.
*   **Initialization (`init`):** On page load, it checks for a list of starting
    types provided by the Go backend (from the `--type` flag). If starting types
    are specified, it creates a column for each. Otherwise, it defaults to
    creating a column for the first type found in the `typeData` object.
*   **Column Creation (`createColumn`):** This function dynamically creates the
    DOM elements for a new column based on a given type name. It populates a
    list with the fields of that type. If a field's type is also a known struct,
    a chevron `>` is added to indicate it can be expanded.
*   **Interaction (`handleFieldClick`):** An event listener on each field
    handles user clicks. When a field is clicked:
    1.  It removes any columns to the right of the currently active column
        to reset the view from that point forward.
    1.  It visually marks the clicked field as `selected`.
    1.  If the clicked field's type exists as a key in the `typeData` object, it
        calls `createColumn` to render a new column for that type to the
        right of the current one.
    1.  The view is automatically scrolled to bring the new column into view.

### Search dialog

* If the user types "/", then bring a dialog to switch which type is being
  shown.
* The dialog should show a list of all of the available "root" types (types that
  have `isRoot: true` in the type data, e.g. Kubernetes resources).
* The user can filter the types displayed by incrementally typing a string,
  which will limit the types shown to those that substring match to what the
  user typed.
* The user can either hit "enter" or click on the type they want to
  display. Pressing "escape" will exit the dialog, canceling the search.
* Pressing "enter" will select the first type on the list if there are multiple
  choices.

### Anchor-based browsing

We want to store the state of the columns (which types are currently viewed and
selected) in the page's anchor state. This will allow the user to share links to
go to specific fields and use the browser history function to navigate.

We need to:

* Save the current state of the columns and selections in the
  `window.location.hash`.
* Update the state when the user opens and closes the column selections.
* Restore the column state from the hash.

## Graph data structure

* A map[key]value should have type of value.
* We should ignore non-POD types (e.g. chan, interface).
* We should ignore Go standard library (not needed).
* Pointers, list, slice and map should be represented in the typeDecorators
  field in a list:
  * ["Ptr"]: pointer to Type.
  * ["List"]: List of Type.
  * ["Ptr" "List"]: pointer to list of Type.
  * ["Map[string]" "List"]: map[string][] of Type.

Example:

```json
{
  "github.com/bowei/rex/sample.Content": {
    "package": "github.com/bowei/rex/sample",
    "typeName": "Content",
    "fields": [
      {
        "fieldName": "NumField",
        "typeName": "int",
        "typeDecorators": [],
        "docString": "NumField is...",
        "parsedDocString": { "elements": [{"type": "p", "content": ["NumField is..."]}] }
      },
      {
        "fieldName": "SubContent",
        "typeName": "github.com/bowei/rex/sample.SubContent",
        "typeDecorators": [],
        "docString": "...",
        "parsedDocString": { "elements": [{"type": "p", "content": ["..."]}] }
      }
    ],
    "docString": "Content is...",
    "parsedDocString": { "elements": [{"type": "p", "content": ["Content is..."]}] },
    "isRoot": false
  },
  "github.com/bowei/rex/sample.MyStruct": {
    "package": "github.com/bowei/rex/sample",
    "typeName": "MyStruct",
    "fields": [
      {
        "fieldName": "Content",
        "typeName": "github.com/bowei/rex/sample.Content",
        "typeDecorators": [],
        "docString": "...",
        "parsedDocString": { "elements": [{"type": "p", "content": ["..."]}] }
      }
    ],
    "docString": "MyStruct is...",
    "parsedDocString": { "elements": [{"type": "p", "content": ["MyStruct is..."]}] },
    "isRoot": true
  },
  "github.com/bowei/rex/sample.MyEnum": {
    "package": "github.com/bowei/rex/sample",
    "typeName": "MyEnum",
    "enumValues": [
      {
        "name": "MyEnumA",
        "docString": "Doc for A",
        "parsedDocString": { "elements": [{"type": "p", "content": ["Doc for A"]}] }
      },
      {
        "name": "MyEnumB",
        "docString": "Doc for B",
        "parsedDocString": { "elements": [{"type": "p", "content": ["Doc for B"]}] }
      },
      {
        "name": "MyEnumC",
        "docString": "Doc for C",
        "parsedDocString": { "elements": [{"type": "p", "content": ["Doc for C"]}] }
      }
    ],
    "docString": "MyEnum is...",
    "parsedDocString": { "elements": [{"type": "p", "content": ["MyEnum is..."]}] }
  }
}
```
