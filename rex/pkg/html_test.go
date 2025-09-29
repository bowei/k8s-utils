package pkg

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateHTML(t *testing.T) {
	types := map[string]TypeInfo{
		"main.MyStruct": {
			Package:  "main",
			TypeName: "MyStruct",
			Fields: []FieldInfo{
				{
					FieldName: "Field1",
					TypeName:  "string",
				},
			},
			DocString: "This is a test struct.",
			IsRoot:    true,
		},
	}
	startType := "main.MyStruct"

	var buf bytes.Buffer
	err := GenerateHTML(types, &buf, startType)
	if err != nil {
		t.Fatalf("GenerateHTML() error = %v", err)
	}

	htmlOutput := buf.String()

	// Check for presence of JSON data
	if !strings.Contains(htmlOutput, `"main.MyStruct":`) {
		t.Errorf("GenerateHTML() output does not contain expected JSON data for types")
	}
	if !strings.Contains(htmlOutput, `"typeName":"MyStruct"`) {
		t.Errorf("GenerateHTML() output does not contain expected JSON data for types")
	}

	// Check for presence of start type
	if !strings.Contains(htmlOutput, `const startTypes = 'main.MyStruct'`) {
		t.Errorf("GenerateHTML() output does not contain expected startType")
	}

	// Check for some basic HTML structure
	if !strings.Contains(htmlOutput, "<!DOCTYPE html>") {
		t.Errorf("GenerateHTML() output does not look like an HTML file")
	}
	if !strings.Contains(htmlOutput, "<body>") {
		t.Errorf("GenerateHTML() output does not contain a body tag")
	}
}
