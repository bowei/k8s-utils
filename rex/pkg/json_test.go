package pkg

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteJSON(t *testing.T) {
	allTypes := map[string]TypeInfo{
		"main.MyStruct": {
			Package:  "main",
			TypeName: "MyStruct",
			Fields: []FieldInfo{
				{
					FieldName: "BField",
					TypeName:  "string",
					Package:   "",
				},
				{
					FieldName: "AField",
					TypeName:  "int",
					Package:   "",
				},
			},
			EnumValues: []EnumInfo{
				{Name: "EnumVal2"},
				{Name: "EnumVal1"},
			},
			DocString: "A test struct.",
		},
	}

	var buf bytes.Buffer
	err := WriteJSON(allTypes, &buf)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	expectedJSON := `{
  "main.MyStruct": {
    "package": "main",
    "typeName": "MyStruct",
    "fields": [
      {
        "fieldName": "AField",
        "typeName": "int",
        "package": "",
        "typeDecorators": null,
        "docString": ""
      },
      {
        "fieldName": "BField",
        "typeName": "string",
        "package": "",
        "typeDecorators": null,
        "docString": ""
      }
    ],
    "enumValues": [
      {
        "name": "EnumVal1",
        "docString": ""
      },
      {
        "name": "EnumVal2",
        "docString": ""
      }
    ],
    "docString": "A test struct.",
    "isRoot": false
  }
}`

	var got, want map[string]TypeInfo
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expectedJSON), &want); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("WriteJSON() mismatch (-want +got):\n%s", diff)
	}
}
