package pkg

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// GoDocString represents a parsed godoc comment.
type GoDocString struct {
	Elements []GoDocElem `json:"elements"`
}

// GoDocElemType defines the type of a documentation element.
type GoDocElemType string

const (
	// GoDocParagraph represents a paragraph of text.
	GoDocParagraph GoDocElemType = "p"
	// GoDocHeading represents a heading.
	GoDocHeading GoDocElemType = "h"
	// GoDocElementList represents a list (bulleted or numbered).
	GoDocElementList GoDocElemType = "l"
	// GoDocCode represents a preformatted code block.
	GoDocCode GoDocElemType = "c"
	// GoDocDirective represents a special directive, like "Deprecated:".
	GoDocDirective GoDocElemType = "d"
)

// GoDocElem represents a single element in a godoc comment, like a paragraph,
// a heading, or a list.
type GoDocElem struct {
	// Type of this element.
	Type GoDocElemType `json:"type"`
	// Content of the element. For lists, each item is an element in the
	// slice. For other types, it's typically a single string in the slice.
	Content []string `json:"content"`
}

// parseGoDocString parses a raw godoc comment string into a structured GoDocString.
func parseGoDocString(comment string) *GoDocString {
	doc := &GoDocString{}
	lines := strings.Split(comment, "\n")

	p := &docParser{lines: lines}

	for p.pos < len(p.lines) {
		line := p.peek()

		// Skip blank lines between blocks.
		if strings.TrimSpace(line) == "" {
			p.pos++
			continue
		}

		if isListItem(line) {
			doc.Elements = append(doc.Elements, p.parseList())
			continue
		}

		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			doc.Elements = append(doc.Elements, p.parseCodeBlock())
			continue
		}

		// '+' are special and will get special styling.
		if strings.HasPrefix(line, "+") {
			p.pos++
			doc.Elements = append(doc.Elements,
				GoDocElem{Type: GoDocDirective, Content: []string{line}})
			continue
		}

		if p.isHeading() {
			p.pos++
			doc.Elements = append(doc.Elements,
				GoDocElem{
					Type: GoDocHeading,
					Content: []string{
						strings.TrimSpace(strings.TrimLeft(line, "#")),
					},
				})
			continue
		}

		// Otherwise, it's a paragraph.
		doc.Elements = append(doc.Elements, p.parseParagraph())
	}

	return doc
}

type docParser struct {
	lines []string
	pos   int
}

func (p *docParser) peek() string {
	if p.pos >= len(p.lines) {
		return ""
	}
	return p.lines[p.pos]
}

func (p *docParser) isHeading() bool {
	// A heading:
	// - Begins with '#' (We allow Markdown multiple heading
	//   markers, Golang does not.)
	// - Must be bracketed by empty line before and after.
	// - Has no leading space.
	// - '#' by itself is not a heading.
	if p.pos == 0 || p.pos == len(p.lines)-1 {
		return false
	}

	before := p.lines[p.pos-1]
	after := p.lines[p.pos+1]
	line := p.lines[p.pos]

	if !(strings.TrimSpace(before) == "" && strings.TrimSpace(after) == "") {
		return false
	}
	if line == "#" {
		return false
	}
	if !strings.HasPrefix(line, "#") {
		return false
	}
	line = strings.TrimLeft(line, "#")

	return strings.HasPrefix(line, " ")
}

func (p *docParser) parseParagraph() GoDocElem {
	var content []string
	for p.pos < len(p.lines) && strings.TrimSpace(p.lines[p.pos]) != "" {
		line := p.lines[p.pos]
		if strings.HasPrefix(line, "+") {
			break
		}
		content = append(content, line)
		p.pos++
	}
	return GoDocElem{Type: GoDocParagraph, Content: []string{strings.Join(content, " ")}}
}

func (p *docParser) parseCodeBlock() GoDocElem {
	var content []string

	firstLine := p.peek()
	indent := 0
	for _, r := range firstLine {
		if r == ' ' || r == '\t' {
			indent++
		} else {
			break
		}
	}

	startPos := p.pos
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		if strings.TrimSpace(line) != "" && !(line[0] == ' ' || line[0] == '\t') {
			break // non-blank, non-indented line
		}
		p.pos++
	}

	for i := startPos; i < p.pos; i++ {
		line := p.lines[i]
		if len(line) >= indent {
			content = append(content, line[indent:])
		} else {
			content = append(content, line)
		}
	}

	return GoDocElem{Type: GoDocCode, Content: []string{strings.Join(content, "\n")}}
}

func (p *docParser) parseList() GoDocElem {
	var items []string

	for p.pos < len(p.lines) && isListItem(p.lines[p.pos]) {
		var currentItem strings.Builder
		line := p.lines[p.pos]

		trimmed := strings.TrimLeft(line, " 	")
		markerEnd := 0
		r, size := utf8.DecodeRuneInString(trimmed)
		if r == '*' || r == '+' || r == '-' || r == '•' {
			markerEnd = size
		} else {
			i := 0
			for i < len(trimmed) && (unicode.IsDigit(rune(trimmed[i])) || unicode.IsLetter(rune(trimmed[i]))) {
				i++
			}
			if i < len(trimmed) && (trimmed[i] == '.' || trimmed[i] == ')') {
				markerEnd = i + 1
			}
		}

		textPart := strings.TrimLeft(trimmed[markerEnd:], " 	")
		currentItem.WriteString(textPart)
		textIndent := len(line) - len(textPart)
		p.pos++

		for p.pos < len(p.lines) {
			nextLine := p.lines[p.pos]
			if strings.TrimSpace(nextLine) == "" {
				break
			}
			if isListItem(nextLine) {
				break
			}

			nextTrimmed := strings.TrimLeft(nextLine, " 	")
			nextIndent := len(nextLine) - len(nextTrimmed)

			if nextIndent >= textIndent {
				currentItem.WriteString("\n")
				currentItem.WriteString(nextLine[textIndent:])
				p.pos++
			} else {
				break
			}
		}
		items = append(items, currentItem.String())
	}

	return GoDocElem{Type: GoDocElementList, Content: items}
}

func isListItem(line string) bool {
	trimmed := strings.TrimLeft(line, " 	")
	if len(trimmed) == 0 {
		return false
	}

	// Bulleted lists
	if strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "*\t") ||
		strings.HasPrefix(trimmed, "+ ") || strings.HasPrefix(trimmed, "+\t") ||
		strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "-\t") ||
		strings.HasPrefix(trimmed, "• ") || strings.HasPrefix(trimmed, "•\t") {
		return true
	}

	// Numbered lists
	i := 0
	for i < len(trimmed) && (unicode.IsDigit(rune(trimmed[i])) || unicode.IsLetter(rune(trimmed[i]))) {
		i++
	}
	if i > 0 && i < len(trimmed) && (trimmed[i] == '.' || trimmed[i] == ')') {
		if i+1 < len(trimmed) && (trimmed[i+1] == ' ' || trimmed[i+1] == '\t') {
			return true
		}
	}

	return false
}
