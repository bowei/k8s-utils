package pkg

import (
	"reflect"
	"testing"
)

func TestParseGoDocString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *GoDocString
	}{
		{
			name:  "Empty",
			input: "",
			expected: &GoDocString{
				Elements: nil,
			},
		},
		{
			name:  "Simple Paragraph",
			input: "This is a simple paragraph.",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"This is a simple paragraph."}},
				},
			},
		},
		{
			name:  "Multiple Paragraphs",
			input: "Paragraph one.\n\nParagraph two.",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"Paragraph one."}},
					{Type: GoDocParagraph, Content: []string{"Paragraph two."}},
				},
			},
		},
		{
			name:  "Markdown Heading 1",
			input: "\n# This is a heading\n",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocHeading, Content: []string{"This is a heading"}},
				},
			},
		},
		{
			name:  "Markdown Heading 4",
			input: "\n#### This is a heading\n",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocHeading, Content: []string{"This is a heading"}},
				},
			},
		},
		{
			name:  "Not heading 1",
			input: "# not a heading",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"# not a heading"}},
				},
			},
		},
		{
			name:  "Not heading 2",
			input: "\n#\n",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"#"}},
				},
			},
		},
		{
			name:  "Not heading 3",
			input: "\n#text\n",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"#text"}},
				},
			},
		},
		{
			name:  "Not heading 4",
			input: "\n# text",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"# text"}},
				},
			},
		},
		{
			name:  "Paragraph with colon",
			input: "This is a line with a colon:\nbut it's part of a paragraph.",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"This is a line with a colon: but it's part of a paragraph."}},
				},
			},
		},
		{
			name:  "Plus Directive",
			input: "+directive: Do not use.",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocDirective, Content: []string{"+directive: Do not use."}},
				},
			},
		},
		{
			name:  "Code Block",
			input: "  code line 1\n  code line 2",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocCode, Content: []string{"code line 1\ncode line 2"}},
				},
			},
		},
		{
			name:  "Code block with blank lines",
			input: "  line 1\n  \n  line 3",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocCode, Content: []string{"line 1\n\nline 3"}},
				},
			},
		},
		{
			name:  "Code block followed by paragraph",
			input: "  code\n\npara",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocCode, Content: []string{"code\n"}},
					{Type: GoDocParagraph, Content: []string{"para"}},
				},
			},
		},
		{
			name:  "Bulleted List",
			input: "* item 1\n* item 2",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocElementList, Content: []string{"item 1", "item 2"}},
				},
			},
		},
		{
			name:  "Numbered List",
			input: "1. item 1\na) item 2",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocElementList, Content: []string{"item 1", "item 2"}},
				},
			},
		},
		{
			name:  "List with multi-line items",
			input: "* item 1\n  more text for item 1\n* item 2",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocElementList, Content: []string{"item 1\nmore text for item 1", "item 2"}},
				},
			},
		},
		{
			name:  "List with blank line between items",
			input: "* item 1\n\n* item 2",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocElementList, Content: []string{"item 1"}},
					{Type: GoDocElementList, Content: []string{"item 2"}},
				},
			},
		},
		{
			name:  "Mixed Content",
			input: "This is a paragraph.\n\n# A Heading\n\n* list item 1\n* list item 2\n\n  code block\n\nAnother paragraph.",
			expected: &GoDocString{
				Elements: []GoDocElem{
					{Type: GoDocParagraph, Content: []string{"This is a paragraph."}},
					{Type: GoDocHeading, Content: []string{"A Heading"}},
					{Type: GoDocElementList, Content: []string{"list item 1", "list item 2"}},
					{Type: GoDocCode, Content: []string{"code block\n"}},
					{Type: GoDocParagraph, Content: []string{"Another paragraph."}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			act := parseGoDocString(tc.input)
			if !reflect.DeepEqual(act, tc.expected) {
				t.Errorf("parseGoDocString() = %+v, want %+v", act, tc.expected)
			}
		})
	}
}
